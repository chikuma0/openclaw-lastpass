package heuristic

import "testing"

func TestSuggestOpenAI(t *testing.T) {
	t.Parallel()

	suggestion, ok := Suggest(Metadata{
		Name:     "OPENAI_API_KEY",
		FullName: "API Keys/dera-next/OPENAI_API_KEY",
		Group:    "API Keys/dera-next",
	})
	if !ok {
		t.Fatal("Suggest() ok = false, want true")
	}
	if suggestion.RefID != "providers/openai/apiKey" {
		t.Fatalf("RefID = %q, want %q", suggestion.RefID, "providers/openai/apiKey")
	}
	if suggestion.Confidence != "high" {
		t.Fatalf("Confidence = %q, want %q", suggestion.Confidence, "high")
	}
}

func TestSuggestGithubPAT(t *testing.T) {
	t.Parallel()

	suggestion, ok := Suggest(Metadata{
		Name:     "github_pat_repo_admin",
		FullName: "Tokens/github_pat_repo_admin",
		Group:    "Tokens",
	})
	if !ok {
		t.Fatal("Suggest() ok = false, want true")
	}
	if suggestion.RefID != "github/token" {
		t.Fatalf("RefID = %q, want %q", suggestion.RefID, "github/token")
	}
}

func TestSuggestNoMatch(t *testing.T) {
	t.Parallel()

	if _, ok := Suggest(Metadata{Name: "Personal Netflix"}); ok {
		t.Fatal("Suggest() ok = true, want false")
	}
}
