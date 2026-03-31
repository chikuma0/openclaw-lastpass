package lastpass

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const metadataSeparator = "\x1f"

type MetadataEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Group    string `json:"group"`
	URL      string `json:"url,omitempty"`
	Username string `json:"username,omitempty"`
}

func (c *Client) ListMetadata(ctx context.Context) ([]MetadataEntry, error) {
	result, err := c.run(ctx, "ls", "--sync=auto", "--color=never", "--format", "%ai"+metadataSeparator+"%an"+metadataSeparator+"%aN"+metadataSeparator+"%ag"+metadataSeparator+"%al"+metadataSeparator+"%au")
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

		fields := strings.Split(line, metadataSeparator)
		if len(fields) != 6 {
			return nil, fmt.Errorf("parse metadata line %d: expected 6 fields, got %d", index+1, len(fields))
		}

		entry := MetadataEntry{
			ID:       strings.TrimSpace(fields[0]),
			Name:     fields[1],
			FullName: fields[2],
			Group:    fields[3],
			URL:      strings.TrimSpace(fields[4]),
			Username: sanitizeUsername(fields[5]),
		}

		if entry.ID == "" {
			return nil, errors.New("parse metadata: entry ID must not be empty")
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
