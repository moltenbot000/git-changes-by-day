# git-changes-by-day

`git-changes-by-day` is a Go command-line tool that reads git history from a target repository and writes two CSV files:

- `daily-summary.csv`: one row per day with commit counts and line-change totals.
- `commit-text.csv`: one row per commit with timestamped title/body text and change counts.

## Usage

```bash
go run . -repo /path/to/repo -summary-out ./daily-summary.csv -text-out ./commit-text.csv
```

Flags:

- `-repo`: repository to inspect. Defaults to the current directory.
- `-summary-out`: output path for the per-day aggregate CSV.
- `-text-out`: output path for the per-commit text CSV.

## Output

`daily-summary.csv` columns:

- `date`
- `commit_count`
- `lines_added`
- `lines_deleted`
- `lines_changed`

`commit-text.csv` columns:

- `datetime`
- `date`
- `commit_hash`
- `title`
- `body`
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
- `daily-artifact`: generates CSV artifacts on a schedule or manual dispatch and uploads them as workflow artifacts.
