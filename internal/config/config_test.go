package config

import (
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "testdata", "mapping.valid.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Mappings) != 2 {
		t.Fatalf("len(cfg.Mappings) = %d, want 2", len(cfg.Mappings))
	}

	ids := cfg.SortedIDs()
	want := []string{"providers/anthropic/apiKey", "providers/openai/apiKey"}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids[%d] = %q, want %q", i, ids[i], want[i])
		}
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "testdata", "mapping.empty-field.json")
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "testdata", "mapping.invalid.json")
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
}

func TestResolvePathWithInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		explicit string
		envPath  string
		goos     string
		home     string
		xdg      string
		want     string
		wantErr  bool
	}{
		{
			name:     "explicit wins",
			explicit: "~/custom/mapping.json",
			envPath:  "/ignored.json",
			goos:     "linux",
			home:     "/home/alice",
			want:     filepath.Clean("/home/alice/custom/mapping.json"),
		},
		{
			name:    "env override",
			envPath: "~/env/mapping.json",
			goos:    "linux",
			home:    "/home/alice",
			want:    filepath.Clean("/home/alice/env/mapping.json"),
		},
		{
			name: "mac default",
			goos: "darwin",
			home: "/Users/alice",
			want: filepath.Clean("/Users/alice/Library/Application Support/openclaw-lastpass/mapping.json"),
		},
		{
			name: "linux xdg default",
			goos: "linux",
			home: "/home/alice",
			xdg:  "/tmp/xdg",
			want: filepath.Clean("/tmp/xdg/openclaw-lastpass/mapping.json"),
		},
		{
			name:    "missing home",
			goos:    "linux",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ResolvePathWithInputs(tt.explicit, tt.envPath, tt.goos, tt.home, tt.xdg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ResolvePathWithInputs() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolvePathWithInputs() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ResolvePathWithInputs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		kind  FieldKind
		name  string
	}{
		{input: "password", kind: FieldPassword, name: "Password"},
		{input: "Password", kind: FieldPassword, name: "Password"},
		{input: "username", kind: FieldUsername, name: "Username"},
		{input: "note", kind: FieldNote, name: "Note"},
		{input: "Notes", kind: FieldNote, name: "Note"},
		{input: "API Key", kind: FieldCustom, name: "API Key"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, err := ParseField(tt.input)
			if err != nil {
				t.Fatalf("ParseField() error = %v", err)
			}
			if got.Kind != tt.kind || got.Name != tt.name {
				t.Fatalf("ParseField() = %#v, want kind=%q name=%q", got, tt.kind, tt.name)
			}
		})
	}
}
