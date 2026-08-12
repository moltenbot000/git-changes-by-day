package report

import (
	"strconv"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
)

func CommitTextRecord(commit gitlog.Commit) []string {
	committedAtUTC := commit.CommittedAt.UTC()

	return []string{
		committedAtUTC.Format("2006-01-02T15:04:05Z07:00"),
		committedAtUTC.Format("2006-01-02"),
		commit.Hash,
		commit.AuthorEmail,
		commit.GitHubAuthorHandle,
		commit.GitHubAuthorDisplayName,
		commit.CombinedText,
		strconv.Itoa(commit.FilesChanged),
		strconv.Itoa(commit.LinesAdded),
		strconv.Itoa(commit.LinesDeleted),
		strconv.Itoa(commit.LinesChanged()),
	}
}
