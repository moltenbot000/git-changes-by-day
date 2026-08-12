package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/moltenbot000/git-changes-by-day/internal/gitlog"
	"github.com/moltenbot000/git-changes-by-day/internal/report"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("git-changes-by-day", flag.ContinueOnError)
	fs.SetOutput(stderr)

	repoPath := fs.String("repo", ".", "path to the git repository to summarize")
	textOut := fs.String("text-out", "commit-text.csv", "path to write the commit text CSV")

	if err := fs.Parse(args); err != nil {
		return err
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

func writeCommitText(path string, commits []gitlog.Commit) error {
	file, err := createOutputFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"datetime", "date", "commit_hash", "author_email", "github_author_handle", "github_author_display_name", "text", "files_changed", "lines_added", "lines_deleted", "lines_changed", "co_author_emails", "github_co_author_handles", "github_co_author_display_names"}); err != nil {
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
