package lastpass

import (
	"strings"
	"testing"
)

func TestParseMetadataList(t *testing.T) {
	t.Parallel()

	output := joinMetadataRecord(
		"123",
		"OPENAI_API_KEY",
		"API Keys/OPENAI_API_KEY",
		"API Keys",
		"https://platform.openai.com",
		"svc-openai@example.com",
	) + "\n" + joinMetadataRecord(
		"456",
		"Shared Folder",
		"Shared Folder/",
		"",
		"http://group",
		"",
	) + "\n"

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

func TestParseMetadataListAllowsDelimiterCollisionInsideMetadata(t *testing.T) {
	t.Parallel()

	name := "OPENAI" + metadataSeparatorIDName + "API_KEY"
	fullName := "API Keys/" + name
	output := joinMetadataRecord(
		"123",
		name,
		fullName,
		"API Keys",
		"https://platform.openai.com",
		"svc-openai@example.com",
	) + "\n"

	entries, err := ParseMetadataList(output)
	if err != nil {
		t.Fatalf("ParseMetadataList() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Name != name {
		t.Fatalf("entries[0].Name = %q, want %q", entries[0].Name, name)
	}
	if entries[0].FullName != fullName {
		t.Fatalf("entries[0].FullName = %q, want %q", entries[0].FullName, fullName)
	}
}

func TestParseMetadataListRejectsMalformedLineWithPreview(t *testing.T) {
	t.Parallel()

	badLine := "123" +
		metadataSeparatorIDName + "OPENAI_API_KEY" +
		metadataSeparatorNameFullName + "API Keys/OPENAI_API_KEY" +
		metadataSeparatorFullNameGroup + "API Keys" +
		metadataSeparatorGroupURL + "https://platform.openai.com"

	_, err := ParseMetadataList(badLine + "\n")
	if err == nil {
		t.Fatal("ParseMetadataList() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "parse metadata line 1:") {
		t.Fatalf("ParseMetadataList() error = %q, want line number context", err)
	}
	if !strings.Contains(err.Error(), "preview=") {
		t.Fatalf("ParseMetadataList() error = %q, want preview", err)
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("ParseMetadataList() error = %q, want entry preview", err)
	}
}

func joinMetadataRecord(fields ...string) string {
	if len(fields) != 6 {
		panic("joinMetadataRecord requires exactly 6 fields")
	}

	return fields[0] +
		metadataSeparatorIDName +
		fields[1] +
		metadataSeparatorNameFullName +
		fields[2] +
		metadataSeparatorFullNameGroup +
		fields[3] +
		metadataSeparatorGroupURL +
		fields[4] +
		metadataSeparatorURLUsername +
		fields[5]
}
