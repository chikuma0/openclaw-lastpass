package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const EnvVar = "OPENCLAW_LASTPASS_CONFIG"

type FieldKind string

const (
	FieldPassword FieldKind = "password"
	FieldUsername FieldKind = "username"
	FieldNote     FieldKind = "note"
	FieldCustom   FieldKind = "custom"
)

type FieldSelector struct {
	Kind FieldKind
	Name string
}

type Mapping struct {
	Entry string `json:"entry"`
	Field string `json:"field"`
}

type Config struct {
	Path     string
	Mappings map[string]Mapping
}

type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid mapping file: %s", strings.Join(e.Issues, "; "))
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	mappings := make(map[string]Mapping)
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	cfg := &Config{
		Path:     path,
		Mappings: mappings,
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func LoadOptional(path string) (*Config, bool, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return &Config{
			Path:     path,
			Mappings: make(map[string]Mapping),
		}, false, nil
	}
	return nil, false, err
}

func (c *Config) Validate() error {
	var issues []string

	for id, mapping := range c.Mappings {
		if strings.TrimSpace(id) == "" {
			issues = append(issues, "mapping IDs must not be empty")
		}
		if strings.TrimSpace(mapping.Entry) == "" {
			issues = append(issues, fmt.Sprintf("%q: entry must not be empty", id))
		}
		if _, err := ParseField(mapping.Field); err != nil {
			issues = append(issues, fmt.Sprintf("%q: %v", id, err))
		}
	}

	if len(issues) > 0 {
		return &ValidationError{Issues: issues}
	}

	return nil
}

func (c *Config) Lookup(id string) (Mapping, bool) {
	mapping, ok := c.Mappings[id]
	return mapping, ok
}

func (c *Config) SortedIDs() []string {
	ids := make([]string, 0, len(c.Mappings))
	for id := range c.Mappings {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func ParseField(raw string) (FieldSelector, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return FieldSelector{}, errors.New("field must not be empty")
	}

	switch strings.ToLower(trimmed) {
	case "password":
		return FieldSelector{Kind: FieldPassword, Name: "Password"}, nil
	case "username":
		return FieldSelector{Kind: FieldUsername, Name: "Username"}, nil
	case "note", "notes", "secure note", "secure notes", "note body", "secure note body":
		return FieldSelector{Kind: FieldNote, Name: "Note"}, nil
	default:
		return FieldSelector{Kind: FieldCustom, Name: trimmed}, nil
	}
}

func Write(path string, mappings map[string]Mapping) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory for %q: %w", path, err)
	}

	data, err := marshalMappings(mappings)
	if err != nil {
		return fmt.Errorf("marshal config %q: %w", path, err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(path), "mapping-*.json")
	if err != nil {
		return fmt.Errorf("create temp config for %q: %w", path, err)
	}

	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("write temp config for %q: %w", path, err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp config for %q: %w", path, err)
	}
	if err := os.Chmod(tempPath, 0o600); err != nil {
		return fmt.Errorf("chmod temp config for %q: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace config %q: %w", path, err)
	}

	return nil
}

func (f FieldSelector) DisplayName() string {
	if f.Name != "" {
		return f.Name
	}

	switch f.Kind {
	case FieldPassword:
		return "Password"
	case FieldUsername:
		return "Username"
	case FieldNote:
		return "Note"
	default:
		return "Custom Field"
	}
}

func ResolvePath(explicit string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}

	return ResolvePathWithInputs(explicit, os.Getenv(EnvVar), runtime.GOOS, home, os.Getenv("XDG_CONFIG_HOME"))
}

func ResolvePathWithInputs(explicit, envPath, goos, home, xdgConfigHome string) (string, error) {
	switch {
	case strings.TrimSpace(explicit) != "":
		return normalizePath(explicit, home), nil
	case strings.TrimSpace(envPath) != "":
		return normalizePath(envPath, home), nil
	default:
		return DefaultPath(goos, home, xdgConfigHome)
	}
}

func DefaultPath(goos, home, xdgConfigHome string) (string, error) {
	if strings.TrimSpace(home) == "" {
		return "", errors.New("home directory is required to determine the default config path")
	}

	if strings.TrimSpace(xdgConfigHome) != "" {
		return filepath.Join(xdgConfigHome, "openclaw-lastpass", "mapping.json"), nil
	}

	if goos == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "openclaw-lastpass", "mapping.json"), nil
	}

	return filepath.Join(home, ".config", "openclaw-lastpass", "mapping.json"), nil
}

func normalizePath(path, home string) string {
	if path == "~" {
		return filepath.Clean(home)
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Clean(filepath.Join(home, path[2:]))
	}
	return filepath.Clean(path)
}

func marshalMappings(mappings map[string]Mapping) ([]byte, error) {
	ids := make([]string, 0, len(mappings))
	for id := range mappings {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, id := range ids {
		value, err := json.Marshal(mappings[id])
		if err != nil {
			return nil, err
		}

		buf.WriteString("  ")
		key, err := json.Marshal(id)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(": ")

		var indented bytes.Buffer
		if err := json.Indent(&indented, value, "  ", "  "); err != nil {
			return nil, err
		}
		buf.Write(indented.Bytes())

		if i < len(ids)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("}\n")

	return buf.Bytes(), nil
}
