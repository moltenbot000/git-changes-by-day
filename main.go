package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
	"github.com/moltenbot000/git-changes-by-day/internal/report"
)

const helpText = `git-changes-by-day - export git commit history as CSV

USAGE
  git-changes-by-day [OPTIONS]
  git-changes-by-day help
  git-changes-by-day /help

QUICK START
  git-changes-by-day -repo /path/to/repo -text-out ./commit-text.csv

OPTIONS
  -repo PATH
        Git repository to inspect. Default: current directory.
  -text-out PATH
        CSV output path. Missing parent directories are created.
        Default: commit-text.csv.
  -h, -help, --help
        Print this reference without reading a repository or writing a file.

OUTPUT
  One CSV row per commit, ordered from oldest to newest. Header columns:
  datetime,date,commit_hash,author_email,github_author_handle,
  github_author_display_name,text,files_changed,lines_added,lines_deleted,
  lines_changed

  datetime                    Commit time normalized to UTC (RFC3339).
  date                        UTC calendar date (YYYY-MM-DD).
  commit_hash                 Full git commit hash.
  author_email                Author email from git history.
  github_author_handle        Handle parsed from a GitHub noreply email;
                              empty for other email forms.
  github_author_display_name  Author name from git history.
  text                        Subject and body joined as one line.
  files_changed               Count of text files reported by git numstat.
  lines_added                 Added text lines.
  lines_deleted               Deleted text lines.
  lines_changed               lines_added + lines_deleted.

BEHAVIOR AND EXIT STATUS
  Requires git and a readable git work tree at -repo. Binary-file numstat
  entries are excluded from file and line totals. Existing -text-out files
  are replaced. Success exits 0; invalid options, repository errors, and
  output errors exit 1 and print details to stderr.

AUTOMATION BEST PRACTICES
  Use explicit absolute paths for -repo and -text-out. Write each concurrent
  run to a distinct output path. Check exit status before consuming CSV.
  Treat author_email as personal data when storing or sharing exports.

EXAMPLES
  git-changes-by-day -repo . -text-out ./artifacts/commit-text.csv
  go run . --help
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("git-changes-by-day", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprint(stderr, helpText)
	}

	if isHelpRequest(args) {
		_, err := fmt.Fprint(stdout, helpText)
		return err
	}

	repoPath := fs.String("repo", ".", "path to the git repository to summarize")
	textOut := fs.String("text-out", "commit-text.csv", "path to write the commit text CSV")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %q; run 'git-changes-by-day help' for usage", fs.Args())
	}

	collector := gitlog.NewCollector(gitlog.ExecRunner{})
	commits, err := collector.Collect(context.Background(), *repoPath)
	if err != nil {
		return err
	}

	if err := writeCommitText(*textOut, commits); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "wrote %d commits to %s\n", len(commits), *textOut)
	return nil
}

func isHelpRequest(args []string) bool {
	if len(args) != 1 {
		return false
	}

	switch args[0] {
	case "help", "/help", "-h", "-help", "--help":
		return true
	default:
		return false
	}
}

func writeCommitText(path string, commits []gitlog.Commit) error {
	file, err := createOutputFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(report.CommitTextHeader); err != nil {
		return err
	}
	for _, commit := range commits {
		if err := writer.Write(report.CommitTextRecord(commit)); err != nil {
			return err
		}
	}
	if err := writer.Error(); err != nil {
		return err
	}
	return nil
}

func createOutputFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create output directory for %q: %w", path, err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create output file %q: %w", path, err)
	}
	return file, nil
}
