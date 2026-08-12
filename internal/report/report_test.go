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
		AuthorEmail:             "octocat@example.com",
		GitHubAuthorHandle:      "octocat",
		GitHubAuthorDisplayName: "The Octocat",
		CoAuthors: []gitlog.CoAuthor{
			{Email: "hubot@users.noreply.github.com", GitHubHandle: "hubot", GitHubDisplayName: "Hubot"},
			{Email: "mona@example.com", GitHubDisplayName: "Mona Lisa"},
		},
		Title:        "feat: add parser",
		Body:         "with tests",
		CombinedText: "feat: add parser with tests",
		FilesChanged: 2,
		LinesAdded:   7,
		LinesDeleted: 3,
	}

	record := CommitTextRecord(commit)
	if got, want := len(record), len(CommitTextHeader); got != want {
		t.Fatalf("len(record) = %d, want header length %d", got, want)
	}
	if got, want := record[0], "2026-04-01T10:00:00Z"; got != want {
		t.Fatalf("record[0] = %q, want %q", got, want)
	}
	if got, want := record[3], "octocat@example.com"; got != want {
		t.Fatalf("record[3] = %q, want %q", got, want)
	}
	if got, want := record[4], "octocat"; got != want {
		t.Fatalf("record[4] = %q, want %q", got, want)
	}
	if got, want := record[5], "The Octocat"; got != want {
		t.Fatalf("record[5] = %q, want %q", got, want)
	}
	if got, want := record[11], `["hubot@users.noreply.github.com","mona@example.com"]`; got != want {
		t.Fatalf("record[11] = %q, want %q", got, want)
	}
	if got, want := record[12], `["hubot",""]`; got != want {
		t.Fatalf("record[12] = %q, want %q", got, want)
	}
	if got, want := record[13], `["Hubot","Mona Lisa"]`; got != want {
		t.Fatalf("record[13] = %q, want %q", got, want)
	}
	if got, want := record[6], "feat: add parser with tests"; got != want {
		t.Fatalf("record[6] = %q, want %q", got, want)
	}
	if got, want := record[10], "10"; got != want {
		t.Fatalf("record[10] = %q, want %q", got, want)
	}
}

func TestCommitTextRecordNormalizesDateToUTC(t *testing.T) {
	t.Parallel()

	offsetMinusSeven := time.FixedZone("UTC-07", -7*60*60)
	commit := gitlog.Commit{
		Hash:               "utc-shift",
		CommittedAt:        time.Date(2026, 4, 6, 19, 1, 25, 0, offsetMinusSeven),
		GitHubAuthorHandle: "octocat",
	}

	record := CommitTextRecord(commit)
	if got, want := record[0], "2026-04-07T02:01:25Z"; got != want {
		t.Fatalf("record[0] = %q, want %q", got, want)
	}
	if got, want := record[1], "2026-04-07"; got != want {
		t.Fatalf("record[1] = %q, want %q", got, want)
	}
}
