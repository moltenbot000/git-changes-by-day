# git-changes-by-day

`git-changes-by-day` is a Go command-line tool that reads git history from a target repository and writes a CSV file with one row per commit.

## Build

```bash
go build -o ./bin/git-changes-by-day .
```

Run the compiled binary:

```bash
./bin/git-changes-by-day -repo /path/to/repo -text-out ./commit-text.csv
```

## Usage

```bash
go run . -repo /path/to/repo -text-out ./commit-text.csv
```

Flags:

- `-repo`: repository to inspect. Defaults to the current directory.
- `-text-out`: output path for the per-commit text CSV.

## Output

`commit-text.csv` columns:

- `datetime` (UTC, RFC3339)
- `date` (UTC, YYYY-MM-DD)
- `commit_hash`
- `author_email`
- `github_author_handle`
- `github_author_display_name`
- `text`
- `files_changed`
- `lines_added`
- `lines_deleted`
- `lines_changed`

## Development

```bash
go build ./...
go test ./...
```

GitHub Actions includes:

- `ci`: runs tests on pushes and pull requests.
- `publish-artifact`: generates and uploads the commit CSV artifact after `ci` succeeds for an update to `main`.
