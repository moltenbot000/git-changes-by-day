package report

import (
	"encoding/json"
	"strconv"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
)

func CommitTextRecord(commit gitlog.Commit) []string {
	committedAtUTC := commit.CommittedAt.UTC()
	coAuthorEmails := make([]string, 0, len(commit.CoAuthors))
	coAuthorHandles := make([]string, 0, len(commit.CoAuthors))
	coAuthorDisplayNames := make([]string, 0, len(commit.CoAuthors))
	for _, coAuthor := range commit.CoAuthors {
		coAuthorEmails = append(coAuthorEmails, coAuthor.Email)
		coAuthorHandles = append(coAuthorHandles, coAuthor.GitHubHandle)
		coAuthorDisplayNames = append(coAuthorDisplayNames, coAuthor.GitHubDisplayName)
	}

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
		jsonStrings(coAuthorEmails),
		jsonStrings(coAuthorHandles),
		jsonStrings(coAuthorDisplayNames),
	}
}

func jsonStrings(values []string) string {
	encoded, _ := json.Marshal(values)
	return string(encoded)
}
