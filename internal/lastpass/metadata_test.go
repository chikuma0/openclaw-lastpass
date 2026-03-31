package lastpass

import "testing"

func TestParseMetadataList(t *testing.T) {
	t.Parallel()

	output := "123\x1fOPENAI_API_KEY\x1fAPI Keys/OPENAI_API_KEY\x1fAPI Keys\x1fhttps://platform.openai.com\x1fsvc-openai@example.com\n456\x1fShared Folder\x1fShared Folder/\x1f\x1fhttp://group\x1f\n"

	entries, err := ParseMetadataList(output)
	if err != nil {
		t.Fatalf("ParseMetadataList() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].ID != "123" {
		t.Fatalf("entries[0].ID = %q, want %q", entries[0].ID, "123")
	}
	if entries[0].Username != "svc-openai@example.com" {
		t.Fatalf("entries[0].Username = %q, want %q", entries[0].Username, "svc-openai@example.com")
	}
}

func TestParseMetadataListRejectsBadFieldCount(t *testing.T) {
	t.Parallel()

	if _, err := ParseMetadataList("123\x1fOPENAI_API_KEY\n"); err == nil {
		t.Fatal("ParseMetadataList() error = nil, want parse error")
	}
}
