package lastpass

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const (
	metadataSeparatorIDName        = "\x1f"
	metadataSeparatorNameFullName  = "\x1e"
	metadataSeparatorFullNameGroup = "\x1d"
	metadataSeparatorGroupURL      = "\x1c"
	metadataSeparatorURLUsername   = "\x1b"
)

var metadataFieldSeparators = []string{
	metadataSeparatorIDName,
	metadataSeparatorNameFullName,
	metadataSeparatorFullNameGroup,
	metadataSeparatorGroupURL,
	metadataSeparatorURLUsername,
}

var metadataBoundaryNames = []string{
	"id",
	"name",
	"fullname",
	"group",
	"url",
}

type MetadataEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Group    string `json:"group"`
	URL      string `json:"url,omitempty"`
	Username string `json:"username,omitempty"`
}

func (c *Client) ListMetadata(ctx context.Context) ([]MetadataEntry, error) {
	// Use a different control-character separator for each boundary so a delimiter
	// collision inside one metadata field does not break the whole record.
	result, err := c.run(ctx, "ls", "--sync=auto", "--color=never", "--format",
		"%ai"+metadataSeparatorIDName+
			"%an"+metadataSeparatorNameFullName+
			"%aN"+metadataSeparatorFullNameGroup+
			"%ag"+metadataSeparatorGroupURL+
			"%al"+metadataSeparatorURLUsername+
			"%au")
	if err != nil {
		return nil, c.wrapError("ls", "", "", result, err)
	}

	entries, err := ParseMetadataList(result.Stdout)
	if err != nil {
		return nil, err
	}

	filtered := make([]MetadataEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.TrimSpace(entry.Name) == "" {
			continue
		}
		if entry.URL == "http://group" {
			continue
		}
		filtered = append(filtered, entry)
	}

	return filtered, nil
}

func ParseMetadataList(output string) ([]MetadataEntry, error) {
	lines := strings.Split(output, "\n")
	entries := make([]MetadataEntry, 0, len(lines))

	for index, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}

		entry, err := parseMetadataLine(line, index+1)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func sanitizeUsername(username string) string {
	trimmed := strings.TrimSpace(username)
	switch {
	case trimmed == "":
		return ""
	case len(trimmed) > 128:
		return ""
	case strings.ContainsAny(trimmed, "\n\r\t"):
		return ""
	case strings.Contains(strings.ToLower(trimmed), "-----begin"):
		return ""
	default:
		return trimmed
	}
}

func parseMetadataLine(line string, lineNumber int) (MetadataEntry, error) {
	fields := make([]string, 0, len(metadataFieldSeparators)+1)
	remainder := line

	for index, separator := range metadataFieldSeparators {
		field, rest, ok := strings.Cut(remainder, separator)
		if !ok {
			return MetadataEntry{}, metadataParseError(lineNumber, line,
				fmt.Sprintf("missing separator %s after %s field", strconv.QuoteToASCII(separator), metadataBoundaryNames[index]))
		}
		fields = append(fields, field)
		remainder = rest
	}

	fields = append(fields, remainder)
	entry := MetadataEntry{
		ID:       strings.TrimSpace(fields[0]),
		Name:     fields[1],
		FullName: fields[2],
		Group:    fields[3],
		URL:      strings.TrimSpace(fields[4]),
		Username: sanitizeUsername(fields[5]),
	}

	if entry.ID == "" {
		return MetadataEntry{}, metadataParseError(lineNumber, line, "entry ID must not be empty")
	}

	return entry, nil
}

func metadataParseError(lineNumber int, line, detail string) error {
	return fmt.Errorf("parse metadata line %d: %s; preview=%s", lineNumber, detail, metadataLinePreview(line))
}

func metadataLinePreview(line string) string {
	const maxPreviewBytes = 160

	preview := line
	if len(preview) > maxPreviewBytes {
		preview = preview[:maxPreviewBytes] + "..."
	}

	return strconv.QuoteToASCII(preview)
}
