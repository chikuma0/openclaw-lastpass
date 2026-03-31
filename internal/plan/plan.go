package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/openclaw/openclaw-lastpass/internal/config"
)

const Version1 = 1

type DraftPlan struct {
	Version     int          `json:"version"`
	GeneratedAt string       `json:"generated_at"`
	Entries     []DraftEntry `json:"entries"`
}

type DraftEntry struct {
	LastPassID     string `json:"lastpass_id"`
	Name           string `json:"name"`
	FullName       string `json:"fullname"`
	Group          string `json:"group,omitempty"`
	URL            string `json:"url,omitempty"`
	Username       string `json:"username,omitempty"`
	SuggestedRefID string `json:"suggested_ref_id,omitempty"`
	SuggestedField string `json:"suggested_field,omitempty"`
	Confidence     string `json:"confidence,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Approved       bool   `json:"approved"`
	Disabled       bool   `json:"disabled,omitempty"`
}

type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid draft plan: %s", strings.Join(e.Issues, "; "))
}

func New(now time.Time, entries []DraftEntry) *DraftPlan {
	return &DraftPlan{
		Version:     Version1,
		GeneratedAt: now.UTC().Format(time.RFC3339),
		Entries:     entries,
	}
}

func Load(path string) (*DraftPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read plan %q: %w", path, err)
	}

	var plan DraftPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("parse plan %q: %w", path, err)
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	return &plan, nil
}

func Write(path string, draft *DraftPlan) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create plan directory for %q: %w", path, err)
	}

	data, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plan %q: %w", path, err)
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o600)
}

func (p *DraftPlan) Validate() error {
	var issues []string

	if p.Version != Version1 {
		issues = append(issues, fmt.Sprintf("unsupported version %d", p.Version))
	}
	if strings.TrimSpace(p.GeneratedAt) == "" {
		issues = append(issues, "generated_at must not be empty")
	}

	seenLastPassIDs := make(map[string]struct{})
	seenApprovedRefs := make(map[string]struct{})

	for index, entry := range p.Entries {
		label := fmt.Sprintf("entries[%d]", index)
		if strings.TrimSpace(entry.LastPassID) == "" {
			issues = append(issues, label+": lastpass_id must not be empty")
		}
		if _, ok := seenLastPassIDs[entry.LastPassID]; ok && strings.TrimSpace(entry.LastPassID) != "" {
			issues = append(issues, label+": duplicate lastpass_id "+entry.LastPassID)
		}
		seenLastPassIDs[entry.LastPassID] = struct{}{}

		if entry.Confidence != "" && !isConfidence(entry.Confidence) {
			issues = append(issues, label+": confidence must be one of high, medium, low")
		}

		if entry.Disabled {
			continue
		}
		if !entry.Approved {
			continue
		}

		if strings.TrimSpace(entry.SuggestedRefID) == "" {
			issues = append(issues, label+": approved entries must set suggested_ref_id")
		}
		if _, err := config.ParseField(entry.SuggestedField); err != nil {
			issues = append(issues, label+": invalid suggested_field: "+err.Error())
		}

		if _, ok := seenApprovedRefs[entry.SuggestedRefID]; ok && strings.TrimSpace(entry.SuggestedRefID) != "" {
			issues = append(issues, label+": duplicate approved suggested_ref_id "+entry.SuggestedRefID)
		}
		seenApprovedRefs[entry.SuggestedRefID] = struct{}{}
	}

	if len(issues) > 0 {
		sort.Strings(issues)
		return &ValidationError{Issues: issues}
	}

	return nil
}

func (p *DraftPlan) ActiveApprovedEntries() []DraftEntry {
	entries := make([]DraftEntry, 0, len(p.Entries))
	for _, entry := range p.Entries {
		if entry.Disabled || !entry.Approved {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

func isConfidence(value string) bool {
	switch value {
	case "high", "medium", "low":
		return true
	default:
		return false
	}
}
