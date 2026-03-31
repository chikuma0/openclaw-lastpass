package plan

import (
	"testing"
	"time"
)

func TestValidateApprovedEntries(t *testing.T) {
	t.Parallel()

	draft := New(time.Unix(0, 0), []DraftEntry{
		{
			LastPassID:     "123",
			Name:           "OPENAI_API_KEY",
			SuggestedRefID: "providers/openai/apiKey",
			SuggestedField: "notes",
			Confidence:     "high",
			Approved:       true,
		},
	})

	if err := draft.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsDuplicateApprovedRefIDs(t *testing.T) {
	t.Parallel()

	draft := New(time.Unix(0, 0), []DraftEntry{
		{
			LastPassID:     "123",
			Name:           "OPENAI_API_KEY",
			SuggestedRefID: "providers/openai/apiKey",
			SuggestedField: "notes",
			Approved:       true,
		},
		{
			LastPassID:     "456",
			Name:           "OPENAI_API_KEY_2",
			SuggestedRefID: "providers/openai/apiKey",
			SuggestedField: "password",
			Approved:       true,
		},
	})

	if err := draft.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want duplicate approved ref error")
	}
}

func TestValidateAllowsUnapprovedIncompleteEntries(t *testing.T) {
	t.Parallel()

	draft := New(time.Unix(0, 0), []DraftEntry{
		{
			LastPassID: "123",
			Name:       "OPENAI_API_KEY",
			Approved:   false,
		},
	})

	if err := draft.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
