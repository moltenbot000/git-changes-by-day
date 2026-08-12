package gitlog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const prettyFormat = "%x1e%H%x00%cI%x00%aN%x00%aE%x00%s%x00%b%x00"

type Commit struct {
	Hash                    string
	CommittedAt             time.Time
	AuthorDisplayName       string
	AuthorEmail             string
	GitHubAuthorHandle      string
	GitHubAuthorDisplayName string
	CoAuthors               []CoAuthor
	Title                   string
	Body                    string
	CombinedText            string
	FilesChanged            int
	LinesAdded              int
	LinesDeleted            int
}

type CoAuthor struct {
	Email             string
	GitHubHandle      string
	GitHubDisplayName string
}

func (c Commit) LinesChanged() int {
	return c.LinesAdded + c.LinesDeleted
}

type Runner interface {
	Run(ctx context.Context, repoPath string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, repoPath string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

type Collector struct {
	runner Runner
}

func NewCollector(runner Runner) Collector {
	return Collector{runner: runner}
}

func (c Collector) Collect(ctx context.Context, repoPath string) ([]Commit, error) {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("resolve repo path %q: %w", repoPath, err)
	}

	if _, err := c.runner.Run(ctx, absPath, "rev-parse", "--is-inside-work-tree"); err != nil {
		return nil, fmt.Errorf("validate git repository %q: %w", absPath, err)
	}

	out, err := c.runner.Run(ctx, absPath, "log", "--date=iso-strict", "--numstat", "--pretty=format:"+prettyFormat)
	if err != nil {
		return nil, fmt.Errorf("collect git log from %q: %w", absPath, err)
	}

	commits, err := ParseLog(out)
	if err != nil {
		return nil, err
	}

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].CommittedAt.Before(commits[j].CommittedAt)
	})
	return commits, nil
}

func ParseLog(data []byte) ([]Commit, error) {
	rawRecords := bytes.Split(data, []byte{0x1e})
	commits := make([]Commit, 0, len(rawRecords))

	for _, rawRecord := range rawRecords {
		rawRecord = bytes.TrimSpace(rawRecord)
		if len(rawRecord) == 0 {
			continue
		}

		headerEnd := bytes.IndexByte(rawRecord, 0x00)
		if headerEnd == -1 {
			return nil, fmt.Errorf("parse git log record %q: %w", string(rawRecord), errors.New("missing header terminator"))
		}

		parts := bytes.SplitN(rawRecord, []byte{0x00}, 7)
		if len(parts) != 7 {
			return nil, fmt.Errorf("parse git log record %q: %w", string(rawRecord), errors.New("unexpected field count"))
		}

		committedAt, err := time.Parse(time.RFC3339, string(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("parse commit time %q: %w", string(parts[1]), err)
		}

		commit := Commit{
			Hash:              string(parts[0]),
			CommittedAt:       committedAt,
			AuthorDisplayName: strings.TrimSpace(string(parts[2])),
			AuthorEmail:       strings.TrimSpace(string(parts[3])),
			Title:             strings.TrimSpace(string(parts[4])),
			Body:              strings.TrimSpace(string(parts[5])),
		}
		commit.GitHubAuthorHandle = githubAuthorHandle(commit.AuthorEmail)
		commit.GitHubAuthorDisplayName = commit.AuthorDisplayName
		commit.CoAuthors = parseCoAuthors(commit.Body)
		commit.CombinedText = joinCommitText(commit.Title, commit.Body)

		for _, line := range strings.Split(strings.TrimSpace(string(parts[6])), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			fields := strings.Split(line, "\t")
			if len(fields) != 3 {
				continue
			}

			added, deleted, ok := parseNumstat(fields[0], fields[1])
			if !ok {
				continue
			}

			commit.FilesChanged++
			commit.LinesAdded += added
			commit.LinesDeleted += deleted
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func parseCoAuthors(body string) []CoAuthor {
	var coAuthors []CoAuthor
	for _, line := range strings.Split(body, "\n") {
		key, value, found := strings.Cut(strings.TrimSpace(line), ":")
		if !found || !strings.EqualFold(strings.TrimSpace(key), "co-authored-by") {
			continue
		}

		value = strings.TrimSpace(value)
		open := strings.LastIndex(value, "<")
		if open < 0 || !strings.HasSuffix(value, ">") {
			continue
		}

		displayName := strings.TrimSpace(value[:open])
		email := strings.TrimSpace(value[open+1 : len(value)-1])
		if displayName == "" || email == "" {
			continue
		}

		coAuthors = append(coAuthors, CoAuthor{
			Email:             email,
			GitHubHandle:      githubAuthorHandle(email),
			GitHubDisplayName: displayName,
		})
	}
	return coAuthors
}

func parseNumstat(addedRaw, deletedRaw string) (int, int, bool) {
	if addedRaw == "-" || deletedRaw == "-" {
		return 0, 0, false
	}

	added, err := strconv.Atoi(addedRaw)
	if err != nil {
		return 0, 0, false
	}
	deleted, err := strconv.Atoi(deletedRaw)
	if err != nil {
		return 0, 0, false
	}
	return added, deleted, true
}

func joinCommitText(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		nonEmpty = append(nonEmpty, strings.Join(strings.Fields(part), " "))
	}

	return strings.Join(nonEmpty, " ")
}

func githubAuthorHandle(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.HasSuffix(email, "@users.noreply.github.com") {
		return ""
	}

	localPart := strings.TrimSuffix(email, "@users.noreply.github.com")
	if localPart == "" {
		return ""
	}

	if plus := strings.LastIndex(localPart, "+"); plus >= 0 {
		localPart = localPart[plus+1:]
	}

	return strings.TrimSpace(localPart)
}
