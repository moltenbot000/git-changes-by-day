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
			Hash:         "abc123",
			CommittedAt:  committedAt,
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
	if got, want := rows[1][3], "feat: add reporting body text"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
	if got, want := rows[1][7], "9"; got != want {
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
