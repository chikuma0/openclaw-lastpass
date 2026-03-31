package discovery

import (
	"testing"
	"time"

	"github.com/chikuma0/openclaw-lastpass/internal/lastpass"
)

func TestBuildDraftPlan(t *testing.T) {
	t.Parallel()

	draft := BuildDraftPlan(time.Unix(0, 0), []lastpass.MetadataEntry{
		{
			ID:       "123",
			Name:     "OPENAI_API_KEY",
			FullName: "API Keys/dera-next/OPENAI_API_KEY",
			Group:    "API Keys/dera-next",
		},
		{
			ID:       "456",
			Name:     "Personal Netflix",
			FullName: "Personal/Netflix",
			Group:    "Personal",
		},
	})

	if len(draft.Entries) != 1 {
		t.Fatalf("len(draft.Entries) = %d, want 1", len(draft.Entries))
	}
	if draft.Entries[0].LastPassID != "123" {
		t.Fatalf("draft.Entries[0].LastPassID = %q, want %q", draft.Entries[0].LastPassID, "123")
	}
	if draft.Entries[0].SuggestedRefID != "providers/openai/apiKey" {
		t.Fatalf("draft.Entries[0].SuggestedRefID = %q, want %q", draft.Entries[0].SuggestedRefID, "providers/openai/apiKey")
	}
}
