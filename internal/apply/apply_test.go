package apply

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chikuma0/openclaw-lastpass/internal/config"
	"github.com/chikuma0/openclaw-lastpass/internal/plan"
)

type fakeValidator struct {
	calls int
	err   error
}

func (f *fakeValidator) Resolve(_ context.Context, _ string, _ config.FieldSelector) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return "discarded", nil
}

func TestExecuteMergesApprovedEntries(t *testing.T) {
	t.Parallel()

	existing := &config.Config{
		Mappings: map[string]config.Mapping{
			"providers/anthropic/apiKey": {
				Entry: "existing-id",
				Field: "password",
			},
		},
	}

	draft := plan.New(time.Unix(0, 0), []plan.DraftEntry{
		{
			LastPassID:     "123",
			Name:           "OPENAI_API_KEY",
			SuggestedRefID: "providers/openai/apiKey",
			SuggestedField: "notes",
			Confidence:     "high",
			Approved:       true,
		},
	})

	result, err := Execute(context.Background(), existing, draft, nil, Options{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Added != 1 {
		t.Fatalf("Added = %d, want 1", result.Added)
	}
	if result.FinalMappings["providers/openai/apiKey"].Entry != "123" {
		t.Fatalf("providers/openai/apiKey entry = %q, want %q", result.FinalMappings["providers/openai/apiKey"].Entry, "123")
	}
	if result.FinalMappings["providers/anthropic/apiKey"].Entry != "existing-id" {
		t.Fatalf("existing mapping was not preserved")
	}
}

func TestExecuteValidationCallsResolver(t *testing.T) {
	t.Parallel()

	draft := plan.New(time.Unix(0, 0), []plan.DraftEntry{
		{
			LastPassID:     "123",
			Name:           "OPENAI_API_KEY",
			SuggestedRefID: "providers/openai/apiKey",
			SuggestedField: "password",
			Confidence:     "high",
			Approved:       true,
		},
	})

	validator := &fakeValidator{}
	if _, err := Execute(context.Background(), &config.Config{Mappings: map[string]config.Mapping{}}, draft, validator, Options{Validate: true}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if validator.calls != 1 {
		t.Fatalf("validator.calls = %d, want 1", validator.calls)
	}
}

func TestExecuteValidationPropagatesError(t *testing.T) {
	t.Parallel()

	draft := plan.New(time.Unix(0, 0), []plan.DraftEntry{
		{
			LastPassID:     "123",
			Name:           "OPENAI_API_KEY",
			SuggestedRefID: "providers/openai/apiKey",
			SuggestedField: "password",
			Confidence:     "high",
			Approved:       true,
		},
	})

	wantErr := errors.New("lpass failed")
	validator := &fakeValidator{err: wantErr}
	if _, err := Execute(context.Background(), &config.Config{Mappings: map[string]config.Mapping{}}, draft, validator, Options{Validate: true}); !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
}
