package report

import (
	"testing"
	"time"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
)

func TestCommitTextRecord(t *testing.T) {
	t.Parallel()

	commit := gitlog.Commit{
		Hash:                    "abc123",
		CommittedAt:             time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
		GitHubAuthorHandle:      "octocat",
		GitHubAuthorDisplayName: "The Octocat",
		Title:                   "feat: add parser",
		Body:                    "with tests",
		CombinedText:            "feat: add parser with tests",
		FilesChanged:            2,
		LinesAdded:              7,
		LinesDeleted:            3,
	}

	record := CommitTextRecord(commit)
	if got, want := record[0], "2026-04-01T10:00:00Z"; got != want {
		t.Fatalf("record[0] = %q, want %q", got, want)
	}
	if got, want := record[3], "octocat"; got != want {
		t.Fatalf("record[3] = %q, want %q", got, want)
	}
	if got, want := record[4], "The Octocat"; got != want {
		t.Fatalf("record[4] = %q, want %q", got, want)
	}
	if got, want := record[5], "feat: add parser with tests"; got != want {
		t.Fatalf("record[5] = %q, want %q", got, want)
	}
	if got, want := record[9], "10"; got != want {
		t.Fatalf("record[9] = %q, want %q", got, want)
	}
}
