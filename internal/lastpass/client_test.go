package lastpass

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/openclaw/openclaw-lastpass/internal/config"
)

type runnerCall struct {
	name string
	args []string
}

type runnerResponse struct {
	result Result
	err    error
}

type fakeRunner struct {
	calls     []runnerCall
	responses []runnerResponse
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) (Result, error) {
	f.calls = append(f.calls, runnerCall{
		name: name,
		args: append([]string(nil), args...),
	})

	if len(f.responses) == 0 {
		return Result{}, errors.New("unexpected Run call")
	}

	response := f.responses[0]
	f.responses = f.responses[1:]
	return response.result, response.err
}

func TestResolvePassword(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		responses: []runnerResponse{
			{result: Result{Stdout: "super-secret\n"}},
		},
	}

	client := NewClientWithRunner("lpass", 2*time.Second, runner)
	field, err := config.ParseField("Password")
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	value, err := client.Resolve(context.Background(), "OpenClaw/OpenAI", field)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if value != "super-secret" {
		t.Fatalf("Resolve() value = %q, want %q", value, "super-secret")
	}

	want := []runnerCall{
		{
			name: "lpass",
			args: []string{"show", "--sync=auto", "--color=never", "--password", "OpenClaw/OpenAI"},
		},
	}

	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestResolveCustomField(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		responses: []runnerResponse{
			{result: Result{Stdout: "token-value\n"}},
		},
	}

	client := NewClientWithRunner("lpass", 2*time.Second, runner)
	field, err := config.ParseField("API Token")
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	if _, err := client.Resolve(context.Background(), "OpenClaw/Anthropic", field); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	got := runner.calls[0].args
	want := []string{"show", "--sync=auto", "--color=never", "--field=API Token", "OpenClaw/Anthropic"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("custom field args = %#v, want %#v", got, want)
	}
}

func TestResolveRejectsAmbiguousEntry(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		responses: []runnerResponse{
			{result: Result{Stdout: "Multiple matches found.\n[id: 1] OpenClaw/OpenAI\n[id: 2] OpenClaw/OpenAI\n"}},
		},
	}

	client := NewClientWithRunner("lpass", 2*time.Second, runner)
	field, err := config.ParseField("Password")
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	if _, err := client.Resolve(context.Background(), "OpenClaw/OpenAI", field); err == nil {
		t.Fatal("Resolve() error = nil, want ambiguity error")
	}
}

func TestResolvePropagatesMissingField(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		responses: []runnerResponse{
			{
				result: Result{Stderr: "Could not find specified field 'API Token'.\n"},
				err:    errors.New("exit status 1"),
			},
		},
	}

	client := NewClientWithRunner("lpass", 2*time.Second, runner)
	field, err := config.ParseField("API Token")
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	if _, err := client.Resolve(context.Background(), "OpenClaw/OpenAI", field); err == nil {
		t.Fatal("Resolve() error = nil, want missing field error")
	}
}

func TestResolveBuiltInFieldMappings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		field     string
		wantToken string
	}{
		{name: "password lower", field: "password", wantToken: "--password"},
		{name: "password title", field: "Password", wantToken: "--password"},
		{name: "username lower", field: "username", wantToken: "--username"},
		{name: "username title", field: "Username", wantToken: "--username"},
		{name: "note lower", field: "note", wantToken: "--notes"},
		{name: "notes title", field: "Notes", wantToken: "--notes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &fakeRunner{
				responses: []runnerResponse{
					{result: Result{Stdout: "value\n"}},
				},
			}

			client := NewClientWithRunner("lpass", 2*time.Second, runner)
			field, err := config.ParseField(tt.field)
			if err != nil {
				t.Fatalf("ParseField() error = %v", err)
			}

			if _, err := client.Resolve(context.Background(), "OPENAI_API_KEY22", field); err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			gotArgs := strings.Join(runner.calls[0].args, " ")
			if !strings.Contains(gotArgs, tt.wantToken) {
				t.Fatalf("args = %q, want token %q", gotArgs, tt.wantToken)
			}
			if strings.Contains(gotArgs, "--field=") {
				t.Fatalf("args = %q, did not expect custom field lookup", gotArgs)
			}
		})
	}
}

func TestStripTrailingCommandNewline(t *testing.T) {
	t.Parallel()

	got := stripTrailingCommandNewline("line one\nline two\n")
	want := "line one\nline two"
	if got != want {
		t.Fatalf("stripTrailingCommandNewline() = %q, want %q", got, want)
	}
}
