package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chikuma0/openclaw-lastpass/internal/heuristic"
	"github.com/chikuma0/openclaw-lastpass/internal/lastpass"
	"github.com/chikuma0/openclaw-lastpass/internal/plan"
)

const Version1 = 1

type Snapshot struct {
	Version     int                      `json:"version"`
	GeneratedAt string                   `json:"generated_at"`
	Entries     []lastpass.MetadataEntry `json:"entries"`
}

type Summary struct {
	Scanned     int
	Suggested   int
	High        int
	Medium      int
	Low         int
	DiscoveryTo string
	DraftTo     string
}

func NewSnapshot(now time.Time, entries []lastpass.MetadataEntry) *Snapshot {
	return &Snapshot{
		Version:     Version1,
		GeneratedAt: now.UTC().Format(time.RFC3339),
		Entries:     entries,
	}
}

func BuildDraftPlan(now time.Time, entries []lastpass.MetadataEntry) *plan.DraftPlan {
	planEntries := make([]plan.DraftEntry, 0, len(entries))

	for _, entry := range entries {
		suggestion, ok := heuristic.Suggest(heuristic.Metadata{
			Name:     entry.Name,
			FullName: entry.FullName,
			Group:    entry.Group,
			URL:      entry.URL,
		})
		if !ok {
			continue
		}

		planEntries = append(planEntries, plan.DraftEntry{
			LastPassID:     entry.ID,
			Name:           entry.Name,
			FullName:       entry.FullName,
			Group:          entry.Group,
			URL:            entry.URL,
			Username:       entry.Username,
			SuggestedRefID: suggestion.RefID,
			SuggestedField: suggestion.Field,
			Confidence:     suggestion.Confidence,
			Reason:         suggestion.Reason,
			Approved:       false,
		})
	}

	return plan.New(now, planEntries)
}

func WriteSnapshot(path string, snapshot *Snapshot) error {
	return writeJSONFile(path, snapshot)
}

func WriteDraft(path string, draft *plan.DraftPlan) error {
	return plan.Write(path, draft)
}

func BuildSummary(snapshot *Snapshot, draft *plan.DraftPlan, discoveryPath, draftPath string) Summary {
	summary := Summary{
		Scanned:     len(snapshot.Entries),
		Suggested:   len(draft.Entries),
		DiscoveryTo: discoveryPath,
		DraftTo:     draftPath,
	}

	for _, entry := range draft.Entries {
		switch entry.Confidence {
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
	}

	return summary
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", path, err)
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %q: %w", path, err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}

	return nil
}
