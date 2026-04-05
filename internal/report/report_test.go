package report

import (
	"testing"
	"time"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
)

func TestAggregateByDay(t *testing.T) {
	t.Parallel()

	commits := []gitlog.Commit{
		{CommittedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC), LinesAdded: 3, LinesDeleted: 1},
		{CommittedAt: time.Date(2026, 4, 1, 15, 0, 0, 0, time.UTC), LinesAdded: 5, LinesDeleted: 2},
		{CommittedAt: time.Date(2026, 4, 2, 9, 0, 0, 0, time.UTC), LinesAdded: 4, LinesDeleted: 4},
	}

	summaries := AggregateByDay(commits)
	if len(summaries) != 2 {
		t.Fatalf("len(summaries) = %d, want 2", len(summaries))
	}

	if got, want := summaries[0].CommitCount, 2; got != want {
		t.Fatalf("summaries[0].CommitCount = %d, want %d", got, want)
	}
	if got, want := summaries[0].LinesChanged(), 11; got != want {
		t.Fatalf("summaries[0].LinesChanged() = %d, want %d", got, want)
	}
	if got, want := summaries[1].Date, "2026-04-02"; got != want {
		t.Fatalf("summaries[1].Date = %q, want %q", got, want)
	}
}

func TestCommitTextRecord(t *testing.T) {
	t.Parallel()

	commit := gitlog.Commit{
		Hash:         "abc123",
		CommittedAt:  time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
		Title:        "feat: add parser",
		Body:         "with tests",
		CombinedText: "feat: add parser\nwith tests",
		FilesChanged: 2,
		LinesAdded:   7,
		LinesDeleted: 3,
	}

	record := CommitTextRecord(commit)
	if got, want := record[0], "2026-04-01T10:00:00Z"; got != want {
		t.Fatalf("record[0] = %q, want %q", got, want)
	}
	if got, want := record[9], "10"; got != want {
		t.Fatalf("record[9] = %q, want %q", got, want)
	}
}
