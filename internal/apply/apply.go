package apply

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/chikuma0/openclaw-lastpass/internal/config"
	"github.com/chikuma0/openclaw-lastpass/internal/plan"
)

type Validator interface {
	Resolve(ctx context.Context, entry string, field config.FieldSelector) (string, error)
}

type Options struct {
	Validate bool
	DryRun   bool
}

type Result struct {
	ApprovedCount int
	Added         int
	Updated       int
	FinalMappings map[string]config.Mapping
}

func Execute(ctx context.Context, existing *config.Config, draft *plan.DraftPlan, validator Validator, options Options) (Result, error) {
	if err := draft.Validate(); err != nil {
		return Result{}, err
	}
	if options.Validate && validator == nil {
		return Result{}, fmt.Errorf("validation requested but no validator was provided")
	}

	finalMappings := make(map[string]config.Mapping, len(existing.Mappings))
	for id, mapping := range existing.Mappings {
		finalMappings[id] = mapping
	}

	result := Result{
		ApprovedCount: len(draft.ActiveApprovedEntries()),
		FinalMappings: finalMappings,
	}

	for _, entry := range draft.ActiveApprovedEntries() {
		mapping := config.Mapping{
			Entry: entry.LastPassID,
			Field: entry.SuggestedField,
		}

		if current, ok := finalMappings[entry.SuggestedRefID]; ok {
			if current != mapping {
				result.Updated++
			}
		} else {
			result.Added++
		}
		finalMappings[entry.SuggestedRefID] = mapping

		if options.Validate {
			field, err := config.ParseField(entry.SuggestedField)
			if err != nil {
				return Result{}, fmt.Errorf("validate %q: %w", entry.SuggestedRefID, err)
			}
			if _, err := validator.Resolve(ctx, entry.LastPassID, field); err != nil {
				return Result{}, fmt.Errorf("validate %q: %w", entry.SuggestedRefID, err)
			}
		}
	}

	return result, nil
}

func BuildOpenClawSnippet(commandPath string) (string, error) {
	snippet := map[string]any{
		"secrets": map[string]any{
			"providers": map[string]any{
				"lastpass": map[string]any{
					"source":   "exec",
					"command":  commandPath,
					"args":     []string{"openclaw"},
					"passEnv":  []string{"HOME", "PATH"},
					"jsonOnly": true,
				},
			},
		},
	}

	data, err := json.MarshalIndent(snippet, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func OpenClawPath() (string, bool) {
	path, err := exec.LookPath("openclaw")
	return path, err == nil
}

func SortedMappingIDs(mappings map[string]config.Mapping) []string {
	ids := make([]string, 0, len(mappings))
	for id := range mappings {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func ExecutablePath() string {
	path, err := os.Executable()
	if err != nil {
		return "openclaw-lastpass"
	}
	return path
}
