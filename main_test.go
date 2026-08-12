package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
)

func TestWriteCommitText(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "commit-text.csv")
	committedAt := time.Date(2026, 4, 1, 12, 30, 0, 0, time.UTC)

	commits := []gitlog.Commit{
		{
			Hash:                    "abc123",
			CommittedAt:             committedAt,
			AuthorEmail:             "octocat@example.com",
			GitHubAuthorHandle:      "octocat",
			GitHubAuthorDisplayName: "The Octocat",
			CoAuthors: []gitlog.CoAuthor{
				{Email: "hubot@users.noreply.github.com", GitHubHandle: "hubot", GitHubDisplayName: "Hubot"},
			},
			Title:        "feat: add reporting",
			Body:         "body text",
			CombinedText: "feat: add reporting body text",
			FilesChanged: 1,
			LinesAdded:   7,
			LinesDeleted: 2,
		},
	}

	if err := writeCommitText(path, commits); err != nil {
		t.Fatalf("writeCommitText() error = %v", err)
	}

	rows := readCSV(t, path)
	if got, want := rows[1][2], "abc123"; got != want {
		t.Fatalf("commit_hash = %q, want %q", got, want)
	}
	if got, want := rows[0][3], "author_email"; got != want {
		t.Fatalf("header author_email = %q, want %q", got, want)
	}
	if got, want := rows[0][4], "github_author_handle"; got != want {
		t.Fatalf("header github_author_handle = %q, want %q", got, want)
	}
	if got, want := rows[0][5], "github_author_display_name"; got != want {
		t.Fatalf("header github_author_display_name = %q, want %q", got, want)
	}
	if got, want := rows[0][6], "co_author_emails"; got != want {
		t.Fatalf("header co_author_emails = %q, want %q", got, want)
	}
	if got, want := rows[1][3], "octocat@example.com"; got != want {
		t.Fatalf("author_email = %q, want %q", got, want)
	}
	if got, want := rows[1][4], "octocat"; got != want {
		t.Fatalf("github_author_handle = %q, want %q", got, want)
	}
	if got, want := rows[1][5], "The Octocat"; got != want {
		t.Fatalf("github_author_display_name = %q, want %q", got, want)
	}
	if got, want := rows[1][6], `["hubot@users.noreply.github.com"]`; got != want {
		t.Fatalf("co_author_emails = %q, want %q", got, want)
	}
	if got, want := rows[1][9], "feat: add reporting body text"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
	if got, want := rows[1][13], "9"; got != want {
		t.Fatalf("lines_changed = %q, want %q", got, want)
	}
}

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer file.Close()

	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	return rows
}
