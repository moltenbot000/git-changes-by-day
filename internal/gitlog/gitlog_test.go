package gitlog

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParseLog(t *testing.T) {
	t.Parallel()

	raw := strings.Join([]string{
		"\x1eabc123\x002026-04-01T12:00:00Z\x00The Octocat\x00123+octocat@users.noreply.github.com\x00feat: add cli\x00first body line\nsecond body line\x00",
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
	if got, want := first.CombinedText, "feat: add cli first body line second body line"; got != want {
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

func TestParseNumstat(t *testing.T) {
	t.Parallel()

	if _, _, ok := parseNumstat("-", "-"); ok {
		t.Fatal("parseNumstat() ok = true, want false for binary file")
	}
	if added, deleted, ok := parseNumstat("4", "2"); !ok || added != 4 || deleted != 2 {
		t.Fatalf("parseNumstat() = (%d, %d, %t), want (4, 2, true)", added, deleted, ok)
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
