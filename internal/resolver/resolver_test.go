package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/chikuma0/openclaw-lastpass/internal/config"
)

type fakeClient struct {
	entry string
	field config.FieldSelector
	value string
	err   error
}

func (f *fakeClient) Resolve(_ context.Context, entry string, field config.FieldSelector) (string, error) {
	f.entry = entry
	f.field = field
	if f.err != nil {
		return "", f.err
	}
	return f.value, nil
}

func TestResolve(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Mappings: map[string]config.Mapping{
			"providers/openai/apiKey": {
				Entry: "OpenClaw/OpenAI",
				Field: "Password",
			},
		},
	}

	client := &fakeClient{value: "secret-value"}
	resolver := New(cfg, client)

	value, err := resolver.Resolve(context.Background(), "providers/openai/apiKey")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if value != "secret-value" {
		t.Fatalf("Resolve() value = %q, want %q", value, "secret-value")
	}
	if client.entry != "OpenClaw/OpenAI" {
		t.Fatalf("client entry = %q, want %q", client.entry, "OpenClaw/OpenAI")
	}
	if client.field.Kind != config.FieldPassword {
		t.Fatalf("client field kind = %q, want %q", client.field.Kind, config.FieldPassword)
	}
}

func TestResolveMissingMapping(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Mappings: map[string]config.Mapping{}}
	resolver := New(cfg, &fakeClient{})

	if _, err := resolver.Resolve(context.Background(), "providers/openai/apiKey"); err == nil {
		t.Fatal("Resolve() error = nil, want missing mapping error")
	}
}

func TestResolvePropagatesClientError(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Mappings: map[string]config.Mapping{
			"providers/openai/apiKey": {
				Entry: "OpenClaw/OpenAI",
				Field: "Password",
			},
		},
	}

	wantErr := errors.New("entry not found")
	resolver := New(cfg, &fakeClient{err: wantErr})

	if _, err := resolver.Resolve(context.Background(), "providers/openai/apiKey"); !errors.Is(err, wantErr) {
		t.Fatalf("Resolve() error = %v, want %v", err, wantErr)
	}
}
