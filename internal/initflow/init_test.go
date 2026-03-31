package initflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/openclaw/openclaw-lastpass/internal/lastpass"
)

type fakeLastPassClient struct {
	lookPath        string
	lookPathErr     error
	statusErr       error
	metadata        []lastpass.MetadataEntry
	listMetadataErr error
	listCalls       int
}

func (f *fakeLastPassClient) LookPath() (string, error) {
	if f.lookPathErr != nil {
		return "", f.lookPathErr
	}
	return f.lookPath, nil
}

func (f *fakeLastPassClient) Status(context.Context) error {
	return f.statusErr
}

func (f *fakeLastPassClient) ListMetadata(context.Context) ([]lastpass.MetadataEntry, error) {
	f.listCalls++
	if f.listMetadataErr != nil {
		return nil, f.listMetadataErr
	}
	return f.metadata, nil
}

func TestRunMissingLPass(t *testing.T) {
	t.Parallel()

	_, err := Run(context.Background(), Paths{}, false, &fakeLastPassClient{lookPathErr: errors.New("missing")}, "")
	if !errors.Is(err, ErrLPassMissing) {
		t.Fatalf("Run() error = %v, want ErrLPassMissing", err)
	}
}

func TestRunRequiresLogin(t *testing.T) {
	t.Parallel()

	_, err := Run(context.Background(), Paths{}, false, &fakeLastPassClient{
		lookPath:  "/usr/bin/lpass",
		statusErr: errors.New("not logged in"),
	}, "")
	if !errors.Is(err, ErrLastPassLoginRequired) {
		t.Fatalf("Run() error = %v, want ErrLastPassLoginRequired", err)
	}
}

func TestRunGeneratesDiscoveryFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	paths := Paths{
		ConfigDir:     tempDir,
		DiscoveryPath: filepath.Join(tempDir, "discovery.json"),
		DraftPath:     filepath.Join(tempDir, "mapping.draft.json"),
	}

	result, err := Run(context.Background(), paths, false, &fakeLastPassClient{
		lookPath: "/usr/bin/lpass",
		metadata: []lastpass.MetadataEntry{
			{
				ID:       "123",
				Name:     "OPENAI_API_KEY",
				FullName: "API Keys/OPENAI_API_KEY",
				Group:    "API Keys",
			},
		},
	}, "/usr/local/bin/openclaw")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Generated {
		t.Fatal("Generated = false, want true")
	}
	if result.Suggested != 1 {
		t.Fatalf("Suggested = %d, want 1", result.Suggested)
	}
	if _, err := os.Stat(paths.DiscoveryPath); err != nil {
		t.Fatalf("os.Stat(discovery) error = %v", err)
	}
	if _, err := os.Stat(paths.DraftPath); err != nil {
		t.Fatalf("os.Stat(draft) error = %v", err)
	}
}

func TestRunKeepsExistingDraftWithoutRefresh(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	paths := Paths{
		ConfigDir:     tempDir,
		DiscoveryPath: filepath.Join(tempDir, "discovery.json"),
		DraftPath:     filepath.Join(tempDir, "mapping.draft.json"),
	}

	if err := os.WriteFile(paths.DiscoveryPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(discovery) error = %v", err)
	}
	if err := os.WriteFile(paths.DraftPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(draft) error = %v", err)
	}

	client := &fakeLastPassClient{lookPath: "/usr/bin/lpass"}
	result, err := Run(context.Background(), paths, false, client, "")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Generated {
		t.Fatal("Generated = true, want false")
	}
	if client.listCalls != 0 {
		t.Fatalf("listCalls = %d, want 0", client.listCalls)
	}
}
