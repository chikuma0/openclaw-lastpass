package lastpass

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/openclaw/openclaw-lastpass/internal/config"
)

const defaultTimeout = 10 * time.Second

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
}

type Result struct {
	Stdout string
	Stderr string
}

type Client struct {
	command string
	timeout time.Duration
	runner  Runner
}

type execRunner struct{}

func NewClient(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	return &Client{
		command: "lpass",
		timeout: timeout,
		runner:  execRunner{},
	}
}

func NewClientWithRunner(command string, timeout time.Duration, runner Runner) *Client {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if command == "" {
		command = "lpass"
	}
	if runner == nil {
		runner = execRunner{}
	}

	return &Client{
		command: command,
		timeout: timeout,
		runner:  runner,
	}
}

func (execRunner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, err
}

func (c *Client) LookPath() (string, error) {
	return exec.LookPath(c.command)
}

func (c *Client) Status(ctx context.Context) error {
	result, err := c.run(ctx, "status", "--quiet", "--color=never")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("%s is not installed or not found in PATH", c.command)
		}
		if strings.TrimSpace(result.Stdout) == "" && strings.TrimSpace(result.Stderr) == "" {
			return fmt.Errorf("%s is not logged in or is otherwise unusable on this machine", c.command)
		}
		return c.wrapError("status", "", "", result, err)
	}
	return nil
}

func (c *Client) Resolve(ctx context.Context, entry string, field config.FieldSelector) (string, error) {
	result, err := c.run(ctx, buildShowArgs(field, entry)...)
	if err != nil {
		return "", c.wrapError("show", entry, field.DisplayName(), result, err)
	}
	if isMultipleMatchesOutput(result.Stdout) {
		return "", fmt.Errorf("entry %q matched multiple LastPass items; use a unique entry name or entry ID in the mapping", entry)
	}

	return stripTrailingCommandNewline(result.Stdout), nil
}

func (c *Client) run(ctx context.Context, args ...string) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	runCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.runner.Run(runCtx, c.command, args...)
}

func buildShowArgs(field config.FieldSelector, target string) []string {
	args := []string{"show", "--sync=auto", "--color=never"}

	switch field.Kind {
	case config.FieldPassword:
		args = append(args, "--password")
	case config.FieldUsername:
		args = append(args, "--username")
	case config.FieldNote:
		args = append(args, "--notes")
	case config.FieldCustom:
		args = append(args, "--field="+field.Name)
	default:
		args = append(args, "--password")
	}

	args = append(args, target)
	return args
}

func stripTrailingCommandNewline(value string) string {
	return strings.TrimSuffix(value, "\n")
}

func isMultipleMatchesOutput(value string) bool {
	return strings.HasPrefix(value, "Multiple matches found.\n")
}

func (c *Client) wrapError(command, entry, field string, result Result, err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("%s is not installed or not found in PATH", c.command)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		if entry != "" {
			return fmt.Errorf("%s timed out after %s while accessing entry %q", c.command, c.timeout, entry)
		}
		return fmt.Errorf("%s timed out after %s while running %s", c.command, c.timeout, command)
	}

	message := firstNonEmpty(stripWhitespace(result.Stderr), stripWhitespace(result.Stdout), stripWhitespace(err.Error()))
	switch {
	case strings.Contains(message, "Could not find specified account"):
		return fmt.Errorf("entry %q not found", entry)
	case strings.Contains(message, "Could not find specified field"):
		return fmt.Errorf("field %q not found in entry %q", field, entry)
	case entry != "" && field != "":
		return fmt.Errorf("%s failed for entry %q field %q: %s", c.command, entry, field, message)
	case entry != "":
		return fmt.Errorf("%s failed for entry %q: %s", c.command, entry, message)
	default:
		return fmt.Errorf("%s %s failed: %s", c.command, command, message)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func stripWhitespace(value string) string {
	return strings.TrimSpace(value)
}
