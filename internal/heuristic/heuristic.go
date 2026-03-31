package heuristic

import (
	"strings"
)

type Metadata struct {
	Name     string
	FullName string
	Group    string
	URL      string
}

type Suggestion struct {
	RefID      string
	Field      string
	Confidence string
	Reason     string
}

func Suggest(metadata Metadata) (Suggestion, bool) {
	text := strings.ToLower(strings.Join([]string{metadata.Name, metadata.FullName, metadata.Group}, " "))

	switch {
	case hasAny(text, "openai", "chatgpt") && hasAny(text, "api key", "apikey", "token", "key"):
		return Suggestion{
			RefID:      "providers/openai/apiKey",
			Field:      suggestField(text, metadata.URL),
			Confidence: "high",
			Reason:     "matched openai/chatgpt naming with api key or token keywords in entry name/path",
		}, true
	case hasAll(text, "gpt", "api key") || hasAll(text, "gpt", "token"):
		return Suggestion{
			RefID:      "providers/openai/apiKey",
			Field:      suggestField(text, metadata.URL),
			Confidence: "high",
			Reason:     "matched gpt naming with api key or token keywords in entry name/path",
		}, true
	case hasAny(text, "openai", "chatgpt", "gpt"):
		return Suggestion{
			RefID:      "providers/openai/apiKey",
			Field:      suggestField(text, metadata.URL),
			Confidence: "medium",
			Reason:     "matched openai/chatgpt/gpt naming in entry name/path",
		}, true
	case hasAny(text, "anthropic", "claude") && hasAny(text, "api key", "apikey", "token", "key"):
		return Suggestion{
			RefID:      "providers/anthropic/apiKey",
			Field:      suggestField(text, metadata.URL),
			Confidence: "high",
			Reason:     "matched anthropic/claude naming with api key or token keywords in entry name/path",
		}, true
	case hasAny(text, "anthropic", "claude"):
		return Suggestion{
			RefID:      "providers/anthropic/apiKey",
			Field:      suggestField(text, metadata.URL),
			Confidence: "medium",
			Reason:     "matched anthropic/claude naming in entry name/path",
		}, true
	case hasAny(text, "github_pat", "gh_pat") || (hasAny(text, "github") && hasAny(text, "pat", "personal access token", "token")):
		return Suggestion{
			RefID:      "github/token",
			Field:      suggestField(text, metadata.URL),
			Confidence: "high",
			Reason:     "matched github personal access token naming in entry name/path",
		}, true
	case hasAny(text, "supabase") && hasAny(text, "service_role", "service role"):
		return Suggestion{
			RefID:      "supabase/serviceRole",
			Field:      suggestField(text, metadata.URL),
			Confidence: "high",
			Reason:     "matched supabase service role naming in entry name/path",
		}, true
	case hasAny(text, "discord") && hasAny(text, "bot token", "bot_token", "bottoken", "token"):
		return Suggestion{
			RefID:      "channels.discord.token",
			Field:      suggestField(text, metadata.URL),
			Confidence: "high",
			Reason:     "matched discord bot token naming in entry name/path",
		}, true
	case hasAny(text, "slack") && hasAny(text, "bot token", "bot_token", "bottoken", "token"):
		return Suggestion{
			RefID:      "channels.slack.token",
			Field:      suggestField(text, metadata.URL),
			Confidence: "high",
			Reason:     "matched slack bot token naming in entry name/path",
		}, true
	default:
		return Suggestion{}, false
	}
}

func suggestField(text, url string) string {
	switch {
	case hasAny(text, "secure note", "secure notes", " note ", " notes ", "api key", "api keys", "token", "tokens", "service role", "service_role", "pat"):
		return "notes"
	case strings.TrimSpace(url) == "":
		return "notes"
	default:
		return "password"
	}
}

func hasAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func hasAll(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if !strings.Contains(text, keyword) {
			return false
		}
	}
	return true
}
