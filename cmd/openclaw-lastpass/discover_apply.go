package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	applypkg "github.com/chikuma0/openclaw-lastpass/internal/apply"
	"github.com/chikuma0/openclaw-lastpass/internal/config"
	"github.com/chikuma0/openclaw-lastpass/internal/discovery"
	"github.com/chikuma0/openclaw-lastpass/internal/lastpass"
	"github.com/chikuma0/openclaw-lastpass/internal/plan"
)

func runDiscover(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outDir := fs.String("out-dir", "", "Directory for discovery.json and mapping.draft.json")
	discoveryOut := fs.String("discovery-out", "", "Path to write discovery metadata JSON")
	draftOut := fs.String("draft-out", "", "Path to write editable draft mapping plan JSON")
	timeout := fs.Duration("timeout", defaultTimeout, "Timeout for each lpass subprocess call")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: openclaw-lastpass discover [--out-dir path] [--discovery-out file] [--draft-out file] [--timeout 10s]\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return exitUsage
	}

	configDir, err := resolveConfigDirectory(*outDir)
	if err != nil {
		fmt.Fprintf(stderr, "discover: resolve output directory: %v\n", err)
		return exitConfig
	}

	discoveryPath := *discoveryOut
	if strings.TrimSpace(discoveryPath) == "" {
		discoveryPath = filepath.Join(configDir, "discovery.json")
	}
	draftPath := *draftOut
	if strings.TrimSpace(draftPath) == "" {
		draftPath = filepath.Join(configDir, "mapping.draft.json")
	}

	client := lastpass.NewClient(*timeout)
	entries, err := client.ListMetadata(context.Background())
	if err != nil {
		fmt.Fprintf(stderr, "discover: %v\n", err)
		return exitError
	}

	now := time.Now()
	snapshot := discovery.NewSnapshot(now, entries)
	draft := discovery.BuildDraftPlan(now, entries)

	if err := discovery.WriteSnapshot(discoveryPath, snapshot); err != nil {
		fmt.Fprintf(stderr, "discover: %v\n", err)
		return exitError
	}
	if err := discovery.WriteDraft(draftPath, draft); err != nil {
		fmt.Fprintf(stderr, "discover: %v\n", err)
		return exitError
	}

	summary := discovery.BuildSummary(snapshot, draft, discoveryPath, draftPath)
	fmt.Fprintf(stdout, "Scanned %d LastPass entries.\n", summary.Scanned)
	fmt.Fprintf(stdout, "Wrote metadata snapshot to %s.\n", summary.DiscoveryTo)
	fmt.Fprintf(stdout, "Wrote editable draft plan to %s.\n", summary.DraftTo)
	fmt.Fprintf(stdout, "Generated %d suggestions (%d high, %d medium, %d low).\n", summary.Suggested, summary.High, summary.Medium, summary.Low)
	fmt.Fprintln(stdout, "Review and edit the draft plan before running apply.")

	return exitOK
}

func runApply(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(stderr)
	planPath := fs.String("plan", "", "Path to the approved draft plan JSON")
	configPath := fs.String("config", "", "Path to the resolver mapping JSON file")
	timeout := fs.Duration("timeout", defaultTimeout, "Timeout for each lpass subprocess call")
	dryRun := fs.Bool("dry-run", false, "Preview config updates without writing files")
	validate := fs.Bool("validate", false, "Validate approved entries through lpass before writing")
	printOpenClawConfig := fs.Bool("print-openclaw-config", false, "If OpenClaw is installed, print a recommended provider snippet")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: openclaw-lastpass apply [--plan file] [--config file] [--dry-run] [--validate] [--print-openclaw-config] [--timeout 10s]\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return exitUsage
	}

	targetConfigPath, err := config.ResolvePath(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "apply: resolve config path: %v\n", err)
		return exitConfig
	}

	resolvedPlanPath := *planPath
	if strings.TrimSpace(resolvedPlanPath) == "" {
		resolvedPlanPath = filepath.Join(filepath.Dir(targetConfigPath), "mapping.draft.json")
	}

	draft, err := plan.Load(resolvedPlanPath)
	if err != nil {
		fmt.Fprintf(stderr, "apply: %v\n", err)
		return exitConfig
	}

	existing, _, err := config.LoadOptional(targetConfigPath)
	if err != nil {
		fmt.Fprintf(stderr, "apply: %v\n", err)
		return exitConfig
	}

	options := applypkg.Options{
		Validate: *validate,
		DryRun:   *dryRun,
	}

	var validator applypkg.Validator
	if *validate {
		validator = lastpass.NewClient(*timeout)
	}

	result, err := applypkg.Execute(context.Background(), existing, draft, validator, options)
	if err != nil {
		fmt.Fprintf(stderr, "apply: %v\n", err)
		return exitError
	}

	if !*dryRun {
		if err := config.Write(targetConfigPath, result.FinalMappings); err != nil {
			fmt.Fprintf(stderr, "apply: %v\n", err)
			return exitError
		}
	}

	fmt.Fprintf(stdout, "Loaded plan from %s.\n", resolvedPlanPath)
	fmt.Fprintf(stdout, "Approved entries applied: %d.\n", result.ApprovedCount)
	fmt.Fprintf(stdout, "Mappings added: %d. Mappings updated: %d.\n", result.Added, result.Updated)
	if *dryRun {
		fmt.Fprintf(stdout, "Dry run only. No changes were written to %s.\n", targetConfigPath)
	} else {
		fmt.Fprintf(stdout, "Wrote resolver mapping to %s.\n", targetConfigPath)
	}
	if *validate {
		fmt.Fprintln(stdout, "Validation ran through lpass and discarded resolved values immediately.")
	}

	if *printOpenClawConfig {
		if _, ok := applypkg.OpenClawPath(); !ok {
			fmt.Fprintln(stderr, "apply: openclaw not found in PATH; skipping provider snippet")
		} else {
			snippet, err := applypkg.BuildOpenClawSnippet(applypkg.ExecutablePath())
			if err != nil {
				fmt.Fprintf(stderr, "apply: build OpenClaw snippet: %v\n", err)
				return exitError
			}
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Recommended OpenClaw provider snippet:")
			fmt.Fprintln(stdout, snippet)
		}
	}

	return exitOK
}

func resolveConfigDirectory(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Clean(explicit), nil
	}

	defaultConfigPath, err := config.ResolvePath("")
	if err != nil {
		return "", err
	}
	return filepath.Dir(defaultConfigPath), nil
}
