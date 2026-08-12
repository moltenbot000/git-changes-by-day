package gitlog

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommitLinesChanged(t *testing.T) {
	t.Parallel()
	commit := Commit{LinesAdded: 7, LinesDeleted: 4}
	if got, want := commit.LinesChanged(), 11; got != want {
		t.Fatalf("LinesChanged() = %d, want %d", got, want)
	}
}

func TestExecRunner(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if out, err := (ExecRunner{}).Run(context.Background(), dir, "init"); err != nil {
		t.Fatalf("Run(init) error = %v, output = %q", err, out)
	}
	if out, err := (ExecRunner{}).Run(context.Background(), dir, "rev-parse", "--is-inside-work-tree"); err != nil || strings.TrimSpace(string(out)) != "true" {
		t.Fatalf("Run(rev-parse) = (%q, %v), want true, nil", out, err)
	}
	if _, err := (ExecRunner{}).Run(context.Background(), dir, "definitely-not-a-command"); err == nil || !strings.Contains(err.Error(), "git definitely-not-a-command") {
		t.Fatalf("Run(invalid) error = %v, want contextual error", err)
	}
}

func TestParseLog(t *testing.T) {
	t.Parallel()

	raw := strings.Join([]string{
		"\x1eabc123\x002026-04-01T12:00:00Z\x00The Octocat\x00123+octocat@users.noreply.github.com\x00feat: add cli\x00first body line\nsecond body line\n\nCo-authored-by: Hubot <456+hubot@users.noreply.github.com>\nCo-Authored-By: Mona Lisa <mona@example.com>\x00",
		"5\t3\tmain.go",
		"2\t1\tREADME.md",
		"\x1edef456\x002026-04-02T08:15:00Z\x00Monalisa Octocat\x00monalisa@example.com\x00fix: handle binary\x00\x00",
		"-\t-\timage.png",
		"3\t0\tinternal/gitlog/gitlog.go",
	}, "\n")

	commits, err := ParseLog([]byte(raw))
	if err != nil {
		t.Fatalf("ParseLog() error = %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("len(commits) = %d, want 2", len(commits))
	}

	first := commits[0]
	if got, want := first.FilesChanged, 2; got != want {
		t.Fatalf("first.FilesChanged = %d, want %d", got, want)
	}
	if got, want := first.LinesAdded, 7; got != want {
		t.Fatalf("first.LinesAdded = %d, want %d", got, want)
	}
	if got, want := first.LinesDeleted, 4; got != want {
		t.Fatalf("first.LinesDeleted = %d, want %d", got, want)
	}
	if got, want := first.CombinedText, "feat: add cli first body line second body line Co-authored-by: Hubot <456+hubot@users.noreply.github.com> Co-Authored-By: Mona Lisa <mona@example.com>"; got != want {
		t.Fatalf("first.CombinedText = %q, want %q", got, want)
	}
	if got, want := first.AuthorEmail, "123+octocat@users.noreply.github.com"; got != want {
		t.Fatalf("first.AuthorEmail = %q, want %q", got, want)
	}
	if got, want := first.GitHubAuthorHandle, "octocat"; got != want {
		t.Fatalf("first.GitHubAuthorHandle = %q, want %q", got, want)
	}
	if got, want := first.GitHubAuthorDisplayName, "The Octocat"; got != want {
		t.Fatalf("first.GitHubAuthorDisplayName = %q, want %q", got, want)
	}
	if got, want := len(first.CoAuthors), 2; got != want {
		t.Fatalf("len(first.CoAuthors) = %d, want %d", got, want)
	}
	if got, want := first.CoAuthors[0].Email, "456+hubot@users.noreply.github.com"; got != want {
		t.Fatalf("first.CoAuthors[0].Email = %q, want %q", got, want)
	}
	if got, want := first.CoAuthors[0].GitHubHandle, "hubot"; got != want {
		t.Fatalf("first.CoAuthors[0].GitHubHandle = %q, want %q", got, want)
	}
	if got, want := first.CoAuthors[1].GitHubDisplayName, "Mona Lisa"; got != want {
		t.Fatalf("first.CoAuthors[1].GitHubDisplayName = %q, want %q", got, want)
	}

	second := commits[1]
	if got, want := second.AuthorEmail, "monalisa@example.com"; got != want {
		t.Fatalf("second.AuthorEmail = %q, want %q", got, want)
	}
	if got, want := second.FilesChanged, 1; got != want {
		t.Fatalf("second.FilesChanged = %d, want %d", got, want)
	}
	if got := second.GitHubAuthorHandle; got != "" {
		t.Fatalf("second.GitHubAuthorHandle = %q, want empty", got)
	}
}

func TestParseLogRejectsBadHeader(t *testing.T) {
	t.Parallel()

	_, err := ParseLog([]byte("\x1eabc123\x002026-04-01T12:00:00Z\x00The Octocat\x00octocat@users.noreply.github.com\x00missing-body"))
	if err == nil {
		t.Fatal("ParseLog() error = nil, want error")
	}
}

func TestParseLogRejectsMissingHeaderTerminator(t *testing.T) {
	t.Parallel()
	_, err := ParseLog([]byte("\x1eno-null-fields"))
	if err == nil || !strings.Contains(err.Error(), "missing header terminator") {
		t.Fatalf("ParseLog() error = %v, want missing header terminator", err)
	}
}

func TestParseLogRejectsBadCommitTime(t *testing.T) {
	t.Parallel()
	raw := "\x1eabc\x00not-a-time\x00Name\x00email@example.com\x00title\x00body\x00"
	_, err := ParseLog([]byte(raw))
	if err == nil || !strings.Contains(err.Error(), "parse commit time") {
		t.Fatalf("ParseLog() error = %v, want commit time error", err)
	}
}

func TestParseLogSkipsEmptyAndMalformedNumstat(t *testing.T) {
	t.Parallel()
	raw := "\n\x1eabc\x002026-04-01T12:00:00Z\x00 Name \x00 email@example.com \x00 title \x00 body \x00\nmalformed\n \nnope\t2\ta.go\n2\tnope\tb.go\n\n"
	commits, err := ParseLog([]byte(raw))
	if err != nil {
		t.Fatalf("ParseLog() error = %v", err)
	}
	if len(commits) != 1 || commits[0].FilesChanged != 0 {
		t.Fatalf("ParseLog() = %#v, want one commit with no changed files", commits)
	}
}

func TestCollectorCollectSortsCommits(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		outputs: map[string][]byte{
			"rev-parse --is-inside-work-tree": []byte("true\n"),
			"log --date=iso-strict --numstat --pretty=format:%x1e%H%x00%cI%x00%aN%x00%aE%x00%s%x00%b%x00": []byte(strings.Join([]string{
				"\x1eb\x002026-04-02T12:00:00Z\x00B\x00b@users.noreply.github.com\x00second\x00\x00",
				"1\t1\tb.go",
				"\x1ea\x002026-04-01T12:00:00Z\x00A\x00a@users.noreply.github.com\x00first\x00\x00",
				"1\t0\ta.go",
			}, "\n")),
		},
	}

	commits, err := NewCollector(runner).Collect(context.Background(), ".")
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("len(commits) = %d, want 2", len(commits))
	}
	if got, want := commits[0].Hash, "a"; got != want {
		t.Fatalf("commits[0].Hash = %q, want %q", got, want)
	}
}

func TestCollectorCollectValidatesRepo(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		errs: map[string]error{
			"rev-parse --is-inside-work-tree": errors.New("fatal: not a git repository"),
		},
	}

	_, err := NewCollector(runner).Collect(context.Background(), ".")
	if err == nil {
		t.Fatal("Collect() error = nil, want error")
	}
}

func TestCollectorCollectReportsLogAndParseErrors(t *testing.T) {
	t.Parallel()

	t.Run("log", func(t *testing.T) {
		runner := &stubRunner{outputs: map[string][]byte{"rev-parse --is-inside-work-tree": []byte("true")}, errs: map[string]error{"log --date=iso-strict --numstat --pretty=format:" + prettyFormat: errors.New("log failed")}}
		_, err := NewCollector(runner).Collect(context.Background(), ".")
		if err == nil || !strings.Contains(err.Error(), "collect git log") {
			t.Fatalf("Collect() error = %v, want log context", err)
		}
	})

	t.Run("parse", func(t *testing.T) {
		runner := &stubRunner{outputs: map[string][]byte{"rev-parse --is-inside-work-tree": []byte("true"), "log --date=iso-strict --numstat --pretty=format:" + prettyFormat: []byte("\x1ebad")}}
		_, err := NewCollector(runner).Collect(context.Background(), ".")
		if err == nil || !strings.Contains(err.Error(), "missing header terminator") {
			t.Fatalf("Collect() error = %v, want parse error", err)
		}
	})
}

func TestCollectorCollectWithRealRepository(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runner := ExecRunner{}
	commands := [][]string{{"init"}, {"config", "user.name", "Test User"}, {"config", "user.email", "test@example.com"}}
	for _, args := range commands {
		if _, err := runner.Run(context.Background(), dir, args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "file.txt"}, {"commit", "-m", "initial"}} {
		if _, err := runner.Run(context.Background(), dir, args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	commits, err := NewCollector(runner).Collect(ctx, dir)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(commits) != 1 || commits[0].Title != "initial" || commits[0].FilesChanged != 1 {
		t.Fatalf("Collect() = %#v, want initial commit", commits)
	}
}

func TestParseNumstat(t *testing.T) {
	t.Parallel()

	if _, _, ok := parseNumstat("-", "-"); ok {
		t.Fatal("parseNumstat() ok = true, want false for binary file")
	}
	if added, deleted, ok := parseNumstat("4", "2"); !ok || added != 4 || deleted != 2 {
		t.Fatalf("parseNumstat() = (%d, %d, %t), want (4, 2, true)", added, deleted, ok)
	}
	for _, values := range [][2]string{{"bad", "2"}, {"2", "bad"}, {"-", "2"}, {"2", "-"}} {
		if _, _, ok := parseNumstat(values[0], values[1]); ok {
			t.Errorf("parseNumstat(%q, %q) ok = true, want false", values[0], values[1])
		}
	}
}

func TestJoinCommitText(t *testing.T) {
	t.Parallel()

	if got, want := joinCommitText("feat: add parser", "with\n\nextra   spacing"), "feat: add parser with extra spacing"; got != want {
		t.Fatalf("joinCommitText() = %q, want %q", got, want)
	}
}

func TestGithubAuthorHandle(t *testing.T) {
	t.Parallel()

	if got, want := githubAuthorHandle("12345+octocat@users.noreply.github.com"), "octocat"; got != want {
		t.Fatalf("githubAuthorHandle() = %q, want %q", got, want)
	}
	if got, want := githubAuthorHandle("octocat@users.noreply.github.com"), "octocat"; got != want {
		t.Fatalf("githubAuthorHandle() = %q, want %q", got, want)
	}
	if got := githubAuthorHandle("octocat@example.com"); got != "" {
		t.Fatalf("githubAuthorHandle() = %q, want empty", got)
	}
	if got := githubAuthorHandle("@users.noreply.github.com"); got != "" {
		t.Fatalf("githubAuthorHandle() = %q, want empty", got)
	}
}

func TestIsTrailerLine(t *testing.T) {
	t.Parallel()
	for _, line := range []string{" continuation", "no colon", ": value", "bad_key: value"} {
		if isTrailerLine(line) {
			t.Errorf("isTrailerLine(%q) = true, want false", line)
		}
	}
	if !isTrailerLine("Reviewed-by: Person") {
		t.Error("isTrailerLine(valid) = false, want true")
	}
}

func TestParseCoAuthorsIgnoresMalformedTrailers(t *testing.T) {
	t.Parallel()

	body := "Co-authored-by: no email\nCo-authored-by: <only@example.com>\nSigned-off-by: Other <other@example.com>"
	if got := parseCoAuthors(body); len(got) != 0 {
		t.Fatalf("parseCoAuthors() = %#v, want empty", got)
	}
}

func TestParseCoAuthorsOnlyUsesTerminalTrailerBlock(t *testing.T) {
	t.Parallel()

	body := "Co-authored-by: Example <example@example.com>\nwas shown above as an example\n\nSigned-off-by: Other <other@example.com>"
	if got := parseCoAuthors(body); len(got) != 0 {
		t.Fatalf("parseCoAuthors() = %#v, want empty", got)
	}
}

func TestParseCoAuthorsRequiresTrailerSeparator(t *testing.T) {
	t.Parallel()

	body := "This change was created with\nCo-authored-by: Example <example@example.com>"
	if got := parseCoAuthors(body); len(got) != 0 {
		t.Fatalf("parseCoAuthors() = %#v, want empty", got)
	}
}

func TestParseCoAuthorsIgnoresIndentedProse(t *testing.T) {
	t.Parallel()

	body := "Description\n\n  Co-authored-by: Example <example@example.com>"
	if got := parseCoAuthors(body); len(got) != 0 {
		t.Fatalf("parseCoAuthors() = %#v, want empty", got)
	}
}

type stubRunner struct {
	outputs map[string][]byte
	errs    map[string]error
}

func (s *stubRunner) Run(_ context.Context, _ string, args ...string) ([]byte, error) {
	key := strings.Join(args, " ")
	if err := s.errs[key]; err != nil {
		return nil, err
	}
	if out := s.outputs[key]; out != nil {
		return out, nil
	}
	return nil, errors.New("unexpected git invocation: " + key)
}
