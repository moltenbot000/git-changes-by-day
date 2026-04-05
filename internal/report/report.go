package report

import (
	"strconv"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
)

type DailySummary struct {
	Date         string
	CommitCount  int
	LinesAdded   int
	LinesDeleted int
}

func (d DailySummary) LinesChanged() int {
	return d.LinesAdded + d.LinesDeleted
}

func (d DailySummary) Record() []string {
	return []string{
		d.Date,
		strconv.Itoa(d.CommitCount),
		strconv.Itoa(d.LinesAdded),
		strconv.Itoa(d.LinesDeleted),
		strconv.Itoa(d.LinesChanged()),
	}
}

func AggregateByDay(commits []gitlog.Commit) []DailySummary {
	if len(commits) == 0 {
		return nil
	}

	summaries := make([]DailySummary, 0)
	indexByDate := make(map[string]int, len(commits))

	for _, commit := range commits {
		date := commit.CommittedAt.Format("2006-01-02")
		idx, ok := indexByDate[date]
		if !ok {
			idx = len(summaries)
			indexByDate[date] = idx
			summaries = append(summaries, DailySummary{Date: date})
		}

		summaries[idx].CommitCount++
		summaries[idx].LinesAdded += commit.LinesAdded
		summaries[idx].LinesDeleted += commit.LinesDeleted
	}

	return summaries
}

func CommitTextRecord(commit gitlog.Commit) []string {
	return []string{
		commit.CommittedAt.Format("2006-01-02T15:04:05Z07:00"),
		commit.CommittedAt.Format("2006-01-02"),
		commit.Hash,
		commit.Title,
		commit.Body,
		commit.CombinedText,
		strconv.Itoa(commit.FilesChanged),
		strconv.Itoa(commit.LinesAdded),
		strconv.Itoa(commit.LinesDeleted),
		strconv.Itoa(commit.LinesChanged()),
	}
}
