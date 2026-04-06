package report

import (
	"strconv"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
)

func CommitTextRecord(commit gitlog.Commit) []string {
	return []string{
		commit.CommittedAt.Format("2006-01-02T15:04:05Z07:00"),
		commit.CommittedAt.Format("2006-01-02"),
		commit.Hash,
		commit.CombinedText,
		strconv.Itoa(commit.FilesChanged),
		strconv.Itoa(commit.LinesAdded),
		strconv.Itoa(commit.LinesDeleted),
		strconv.Itoa(commit.LinesChanged()),
	}
}
