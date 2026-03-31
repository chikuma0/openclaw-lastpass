package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	applypkg "github.com/chikuma0/openclaw-lastpass/internal/apply"
	"github.com/chikuma0/openclaw-lastpass/internal/initflow"
	"github.com/chikuma0/openclaw-lastpass/internal/lastpass"
)

func runInit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configDir := fs.String("config-dir", "", "Directory for discovery and draft files")
	refresh := fs.Bool("refresh", false, "Refresh discovery.json and mapping.draft.json even if they already exist")
	printOpenClawConfig := fs.Bool("print-openclaw-config", false, "If OpenClaw is installed, print a recommended provider snippet")
	timeout := fs.Duration("timeout", defaultTimeout, "Timeout for each lpass subprocess call")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: openclaw-lastpass init [--config-dir path] [--refresh] [--print-openclaw-config] [--timeout 10s]\n")
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

	dir, err := resolveConfigDirectory(*configDir)
	if err != nil {
		fmt.Fprintf(stderr, "init: resolve config directory: %v\n", err)
		return exitConfig
	}

	openClawPath, _ := applypkg.OpenClawPath()
	result, err := initflow.Run(context.Background(), initflow.Paths{
		ConfigDir:     dir,
		DiscoveryPath: filepath.Join(dir, "discovery.json"),
		DraftPath:     filepath.Join(dir, "mapping.draft.json"),
	}, *refresh, lastpass.NewClient(*timeout), openClawPath)
	if err != nil {
		switch {
		case errors.Is(err, initflow.ErrLPassMissing):
			fmt.Fprintln(stderr, "init: lpass is not installed or not found in PATH.")
			fmt.Fprintln(stderr, "Install the LastPass CLI first, then run:")
			fmt.Fprintln(stderr, "  lpass login you@example.com")
			fmt.Fprintln(stderr, "  openclaw-lastpass init")
		case errors.Is(err, initflow.ErrLastPassLoginRequired):
			fmt.Fprintf(stderr, "init: %v\n", err)
			fmt.Fprintln(stderr, "Run:")
			fmt.Fprintln(stderr, "  lpass login you@example.com")
			fmt.Fprintln(stderr, "Then rerun:")
			fmt.Fprintln(stderr, "  openclaw-lastpass init")
		default:
			fmt.Fprintf(stderr, "init: %v\n", err)
		}
		return exitError
	}

	fmt.Fprintf(stdout, "LastPass CLI found at %s.\n", result.LastPassPath)
	fmt.Fprintln(stdout, "LastPass login looks usable.")
	fmt.Fprintf(stdout, "Config directory ready at %s.\n", result.ConfigDir)

	if result.Generated {
		if result.Refreshed {
			fmt.Fprintf(stdout, "Refreshed discovery metadata at %s.\n", result.DiscoveryPath)
			fmt.Fprintf(stdout, "Refreshed draft mapping plan at %s.\n", result.DraftPath)
		} else {
			fmt.Fprintf(stdout, "Wrote discovery metadata to %s.\n", result.DiscoveryPath)
			fmt.Fprintf(stdout, "Wrote draft mapping plan to %s.\n", result.DraftPath)
		}
		fmt.Fprintf(stdout, "Scanned %d entries and generated %d suggestions.\n", result.Scanned, result.Suggested)
	} else {
		fmt.Fprintf(stdout, "Existing draft mapping plan kept at %s.\n", result.DraftPath)
		fmt.Fprintln(stdout, "Rerun with --refresh to regenerate discovery metadata and the draft plan.")
	}

	if result.OpenClawInstalled {
		fmt.Fprintf(stdout, "OpenClaw detected at %s.\n", result.OpenClawPath)
	} else {
		fmt.Fprintln(stdout, "OpenClaw not found in PATH. You can finish secret setup now and configure OpenClaw later.")
	}

	fmt.Fprintln(stdout, "Review is required before apply.")
	fmt.Fprintln(stdout, "Next commands:")
	fmt.Fprintf(stdout, "  cat %s\n", result.DraftPath)
	fmt.Fprintf(stdout, "  openclaw-lastpass apply --plan %s --dry-run\n", result.DraftPath)
	fmt.Fprintf(stdout, "  openclaw-lastpass apply --plan %s --validate\n", result.DraftPath)

	if result.OpenClawInstalled {
		if *printOpenClawConfig {
			snippet, err := applypkg.BuildOpenClawSnippet(applypkg.ExecutablePath())
			if err != nil {
				fmt.Fprintf(stderr, "init: build OpenClaw snippet: %v\n", err)
				return exitError
			}
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Recommended OpenClaw provider snippet:")
			fmt.Fprintln(stdout, snippet)
		} else {
			fmt.Fprintln(stdout, "Rerun with --print-openclaw-config to print the recommended OpenClaw provider snippet.")
		}
	}

	return exitOK
}
