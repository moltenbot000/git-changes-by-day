package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestRunHelp(t *testing.T) {
	t.Parallel()

	for _, arg := range []string{"help", "/help", "-h", "-help", "--help"} {
		arg := arg
		t.Run(arg, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if err := run([]string{arg}, &stdout, &stderr); err != nil {
				t.Fatalf("run(%q) error = %v", arg, err)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			for _, want := range []string{"QUICK START", "OPTIONS", "OUTPUT", "AUTOMATION BEST PRACTICES"} {
				if !strings.Contains(stdout.String(), want) {
					t.Errorf("help output missing %q", want)
				}
			}
		})
	}
}

func TestRunRejectsUnexpectedArguments(t *testing.T) {
	t.Parallel()

	err := run([]string{"unknown"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("run() error = %v, want unexpected arguments error", err)
	}
}

func TestRunRejectsInvalidFlag(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	err := run([]string{"-unknown"}, &bytes.Buffer{}, &stderr)
	if err == nil || !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("run() = %v, stderr %q; want invalid flag", err, stderr.String())
	}
}

func TestRunFlagHelpWithExtraArgument(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	if err := run([]string{"-h", "ignored"}, io.Discard, &stderr); err != nil {
		t.Fatalf("run(-h) error = %v", err)
	}
	if !strings.Contains(stderr.String(), "QUICK START") {
		t.Fatalf("stderr = %q, want usage", stderr.String())
	}
}

func TestRunHelpReportsWriteFailure(t *testing.T) {
	t.Parallel()
	if err := run([]string{"help"}, failingWriter{}, io.Discard); err == nil || err.Error() != "write failed" {
		t.Fatalf("run(help) error = %v, want write failed", err)
	}
}

func TestRunExportsRepository(t *testing.T) {
	t.Parallel()
	dir := createGitRepository(t)
	out := filepath.Join(t.TempDir(), "nested", "commits.csv")
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-repo", dir, "-text-out", out}, &stdout, &stderr); err != nil {
		t.Fatalf("run() error = %v, stderr = %q", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "wrote 1 commits") || len(readCSV(t, out)) != 2 {
		t.Fatalf("stdout = %q, rows = %#v", stdout.String(), readCSV(t, out))
	}
}

func TestRunReportsSuccessWriteFailure(t *testing.T) {
	t.Parallel()
	dir := createGitRepository(t)
	if err := run([]string{"-repo", dir, "-text-out", filepath.Join(t.TempDir(), "out.csv")}, failingWriter{}, io.Discard); err == nil || err.Error() != "write failed" {
		t.Fatalf("run() error = %v, want write failed", err)
	}
}

func TestRunReportsRepositoryAndOutputErrors(t *testing.T) {
	t.Parallel()
	if err := run([]string{"-repo", t.TempDir()}, io.Discard, io.Discard); err == nil {
		t.Fatal("run(non-repository) error = nil, want error")
	}
	dir := createGitRepository(t)
	parentFile := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(parentFile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"-repo", dir, "-text-out", filepath.Join(parentFile, "out.csv")}, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "create output directory") {
		t.Fatalf("run() error = %v, want output directory error", err)
	}
}

func TestWriteCommitTextReportsFlushFailure(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("/dev/full unavailable")
	}
	if err := writeCommitText("/dev/full", []gitlog.Commit{{Hash: "abc"}}); err == nil {
		t.Fatal("writeCommitText(/dev/full) error = nil, want flush error")
	}
}

func TestCreateOutputFileReportsCreateFailure(t *testing.T) {
	t.Parallel()
	if _, err := createOutputFile(t.TempDir()); err == nil || !strings.Contains(err.Error(), "create output file") {
		t.Fatalf("createOutputFile(directory) error = %v, want create error", err)
	}
}

func TestMainExitBehavior(t *testing.T) {
	if os.Getenv("TEST_MAIN_HELPER") == "1" {
		os.Args = []string{"git-changes-by-day", "-repo", filepath.Join(t.TempDir(), "missing")}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainExitBehavior")
	cmd.Env = append(os.Environ(), "TEST_MAIN_HELPER=1")
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 || !strings.Contains(string(out), "validate git repository") {
		t.Fatalf("main subprocess = (%q, %v), want exit 1 repository error", out, err)
	}
}

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
	if got, want := rows[0][11], "co_author_emails"; got != want {
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
	if got, want := rows[1][11], `["hubot@users.noreply.github.com"]`; got != want {
		t.Fatalf("co_author_emails = %q, want %q", got, want)
	}
	if got, want := rows[1][6], "feat: add reporting body text"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
	if got, want := rows[1][10], "9"; got != want {
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

func createGitRepository(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{{"init"}, {"config", "user.name", "Test User"}, {"config", "user.email", "test@example.com"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "file.txt"}, {"commit", "-m", "initial"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}
