package initflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/openclaw/openclaw-lastpass/internal/discovery"
	"github.com/openclaw/openclaw-lastpass/internal/lastpass"
)

var ErrLPassMissing = errors.New("lpass not installed")
var ErrLastPassLoginRequired = errors.New("lastpass login required")

type LastPassClient interface {
	LookPath() (string, error)
	Status(ctx context.Context) error
	ListMetadata(ctx context.Context) ([]lastpass.MetadataEntry, error)
}

type Paths struct {
	ConfigDir     string
	DiscoveryPath string
	DraftPath     string
}

type Result struct {
	ConfigDir         string
	LastPassPath      string
	OpenClawPath      string
	OpenClawInstalled bool
	DiscoveryPath     string
	DraftPath         string
	DraftExisted      bool
	Generated         bool
	Refreshed         bool
	Scanned           int
	Suggested         int
}

func Run(ctx context.Context, paths Paths, refresh bool, client LastPassClient, openclawPath string) (Result, error) {
	result := Result{
		ConfigDir:         paths.ConfigDir,
		DiscoveryPath:     paths.DiscoveryPath,
		DraftPath:         paths.DraftPath,
		OpenClawPath:      openclawPath,
		OpenClawInstalled: openclawPath != "",
	}

	lpassPath, err := client.LookPath()
	if err != nil {
		return result, fmt.Errorf("%w: install the LastPass CLI and make sure it is available in PATH", ErrLPassMissing)
	}
	result.LastPassPath = lpassPath

	if err := client.Status(ctx); err != nil {
		return result, fmt.Errorf("%w: %v", ErrLastPassLoginRequired, err)
	}

	if err := os.MkdirAll(paths.ConfigDir, 0o755); err != nil {
		return result, fmt.Errorf("create config directory %q: %w", paths.ConfigDir, err)
	}

	discoveryExists := fileExists(paths.DiscoveryPath)
	draftExists := fileExists(paths.DraftPath)
	result.DraftExisted = draftExists

	if !refresh && discoveryExists && draftExists {
		return result, nil
	}

	entries, err := client.ListMetadata(ctx)
	if err != nil {
		return result, err
	}

	now := time.Now()
	snapshot := discovery.NewSnapshot(now, entries)
	draft := discovery.BuildDraftPlan(now, entries)

	if err := discovery.WriteSnapshot(paths.DiscoveryPath, snapshot); err != nil {
		return result, err
	}
	if err := discovery.WriteDraft(paths.DraftPath, draft); err != nil {
		return result, err
	}

	summary := discovery.BuildSummary(snapshot, draft, paths.DiscoveryPath, paths.DraftPath)
	result.Generated = true
	result.Refreshed = refresh || discoveryExists || draftExists
	result.Scanned = summary.Scanned
	result.Suggested = summary.Suggested

	return result, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
