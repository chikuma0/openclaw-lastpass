package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/openclaw/openclaw-lastpass/internal/config"
	"github.com/openclaw/openclaw-lastpass/internal/doctor"
	"github.com/openclaw/openclaw-lastpass/internal/lastpass"
	"github.com/openclaw/openclaw-lastpass/internal/protocol"
	"github.com/openclaw/openclaw-lastpass/internal/resolver"
)

const (
	exitOK     = 0
	exitError  = 1
	exitUsage  = 2
	exitConfig = 3
)

const defaultTimeout = 10 * time.Second

type getJSONResponse struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printRootUsage(stderr)
		return exitUsage
	}

	switch args[0] {
	case "help", "-h", "--help":
		printRootUsage(stdout)
		return exitOK
	case "apply":
		return runApply(args[1:], stdout, stderr)
	case "discover":
		return runDiscover(args[1:], stdout, stderr)
	case "get":
		return runGet(args[1:], stdout, stderr)
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "list":
		return runList(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "openclaw":
		return runOpenClaw(args[1:], stdin, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printRootUsage(stderr)
		return exitUsage
	}
}

func runGet(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "Path to the mapping JSON file")
	timeout := fs.Duration("timeout", defaultTimeout, "Timeout for each lpass subprocess call")
	jsonOutput := fs.Bool("json", false, "Print JSON instead of the raw secret value")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: openclaw-lastpass get [--config path] [--timeout 10s] [--json] <secret-id>\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return exitUsage
	}

	cfgPath, err := config.ResolvePath(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "get: resolve config path: %v\n", err)
		return exitConfig
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(stderr, "get: %v\n", err)
		return exitConfig
	}

	client := lastpass.NewClient(*timeout)
	res := resolver.New(cfg, client)

	value, err := res.Resolve(context.Background(), fs.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "get: %v\n", err)
		return exitError
	}

	if *jsonOutput {
		if err := writeJSON(stdout, getJSONResponse{ID: fs.Arg(0), Value: value}); err != nil {
			fmt.Fprintf(stderr, "get: write output: %v\n", err)
			return exitError
		}
		return exitOK
	}

	if _, err := io.WriteString(stdout, value); err != nil {
		fmt.Fprintf(stderr, "get: write output: %v\n", err)
		return exitError
	}
	return exitOK
}

func runList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "Path to the mapping JSON file")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: openclaw-lastpass list [--config path]\n")
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

	cfgPath, err := config.ResolvePath(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "list: resolve config path: %v\n", err)
		return exitConfig
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(stderr, "list: %v\n", err)
		return exitConfig
	}

	for _, id := range cfg.SortedIDs() {
		if _, err := fmt.Fprintln(stdout, id); err != nil {
			fmt.Fprintf(stderr, "list: write output: %v\n", err)
			return exitError
		}
	}

	return exitOK
}

func runDoctor(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "Path to the mapping JSON file")
	timeout := fs.Duration("timeout", defaultTimeout, "Timeout for each lpass subprocess call")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: openclaw-lastpass doctor [--config path] [--timeout 10s]\n")
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

	cfgPath, err := config.ResolvePath(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "doctor: resolve config path: %v\n", err)
		return exitConfig
	}

	report := doctor.Run(context.Background(), cfgPath, lastpass.NewClient(*timeout))
	if err := report.WriteTo(stdout); err != nil {
		fmt.Fprintf(stderr, "doctor: write output: %v\n", err)
		return exitError
	}
	if report.HasFailures() {
		return exitError
	}
	return exitOK
}

func runOpenClaw(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("openclaw", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "Path to the mapping JSON file")
	timeout := fs.Duration("timeout", defaultTimeout, "Timeout for each lpass subprocess call")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: openclaw-lastpass openclaw [--config path] [--timeout 10s]\n")
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

	req, err := protocol.DecodeRequest(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "openclaw: decode request: %v\n", err)
		return exitUsage
	}
	if err := req.Validate(); err != nil {
		fmt.Fprintf(stderr, "openclaw: invalid request: %v\n", err)
		return exitUsage
	}

	resp := protocol.NewResponse()
	cfgPath, err := config.ResolvePath(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "openclaw: resolve config path: %v\n", err)
		populateGlobalErrors(resp, req.IDs, fmt.Sprintf("config path error: %v", err))
		return writeProtocolResponse(stdout, stderr, resp)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(stderr, "openclaw: %v\n", err)
		populateGlobalErrors(resp, req.IDs, err.Error())
		return writeProtocolResponse(stdout, stderr, resp)
	}

	client := lastpass.NewClient(*timeout)
	res := resolver.New(cfg, client)
	seen := make(map[string]resultOrError)

	for _, id := range req.IDs {
		if cached, ok := seen[id]; ok {
			if cached.err != nil {
				resp.Errors[id] = protocol.ErrorDetail{Message: cached.err.Error()}
				continue
			}
			resp.Values[id] = cached.value
			continue
		}

		value, err := res.Resolve(context.Background(), id)
		seen[id] = resultOrError{value: value, err: err}
		if err != nil {
			resp.Errors[id] = protocol.ErrorDetail{Message: err.Error()}
			continue
		}
		resp.Values[id] = value
	}

	if len(resp.Errors) == 0 {
		resp.Errors = nil
	}

	return writeProtocolResponse(stdout, stderr, resp)
}

type resultOrError struct {
	value string
	err   error
}

func populateGlobalErrors(resp *protocol.Response, ids []string, message string) {
	for _, id := range ids {
		resp.Errors[id] = protocol.ErrorDetail{Message: message}
	}
}

func writeProtocolResponse(stdout, stderr io.Writer, resp *protocol.Response) int {
	if len(resp.Errors) == 0 {
		resp.Errors = nil
	}
	if err := protocol.WriteResponse(stdout, resp); err != nil {
		fmt.Fprintf(stderr, "openclaw: write response: %v\n", err)
		return exitError
	}
	return exitOK
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func printRootUsage(w io.Writer) {
	commands := []string{"apply", "discover", "doctor", "get", "init", "list", "openclaw"}
	sort.Strings(commands)

	fmt.Fprintln(w, "Usage: openclaw-lastpass <command> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, command := range commands {
		switch command {
		case "apply":
			fmt.Fprintln(w, "  apply     Validate and write approved draft mappings to the resolver config")
		case "discover":
			fmt.Fprintln(w, "  discover  Scan LastPass metadata and write an editable draft mapping plan")
		case "doctor":
			fmt.Fprintln(w, "  doctor    Run local diagnostics for lpass access and mapping health")
		case "get":
			fmt.Fprintln(w, "  get       Resolve one configured secret ID")
		case "init":
			fmt.Fprintln(w, "  init      Guide first-run setup and generate discovery/draft files")
		case "list":
			fmt.Fprintln(w, "  list      List configured secret IDs")
		case "openclaw":
			fmt.Fprintln(w, "  openclaw  Resolve secrets using the OpenClaw exec-provider protocol")
		}
	}
}
