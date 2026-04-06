# git-changes-by-day

`git-changes-by-day` is a Go command-line tool that reads git history from a target repository and writes a CSV file with one row per commit.

## Usage

```bash
go run . -repo /path/to/repo -text-out ./commit-text.csv
```

Flags:

- `-repo`: repository to inspect. Defaults to the current directory.
- `-text-out`: output path for the per-commit text CSV.

## Output

`commit-text.csv` columns:

- `datetime`
- `date`
- `commit_hash`
- `github_author_handle`
- `github_author_display_name`
- `text`
- `files_changed`
- `lines_added`
- `lines_deleted`
- `lines_changed`

## Development

```bash
go test ./...
```

GitHub Actions includes:

- `ci`: runs tests on pushes and pull requests.
- `daily-artifact`: generates the commit CSV artifact on a schedule or manual dispatch and uploads it as a workflow artifact.
