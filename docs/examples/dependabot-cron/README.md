# Dependabot-like Updater Example

This example shows how to build a simple dependency updater using the managers library, running as a scheduled GitHub Action.

## How it works

1. Detects which package manager the repository uses (npm, yarn, pnpm, bun, bundler, cargo, go, uv, poetry, composer, mix, pub, cocoapods)
2. Runs the `outdated` command to find dependencies with available updates
3. For each outdated dependency, creates a branch, updates it, and opens a PR

## Files

- `main.go` - The updater program
- `.github/workflows/deps-update.yml` - GitHub Action workflow (copy to your repo)

## Usage

### As a GitHub Action

Copy the workflow file to your repository:

```bash
mkdir -p .github/workflows
cp .github/workflows/deps-update.yml your-repo/.github/workflows/
```

The action runs weekly by default. Edit the cron schedule to change frequency:

```yaml
schedule:
  - cron: '0 9 * * 1'  # Every Monday at 9am UTC
```

### Running locally

```bash
cd docs/examples/dependabot-cron
go build -o deps-update .
./deps-update /path/to/your/repo
```

## Comparison with Dependabot

| Feature | This example | Dependabot |
|---------|--------------|------------|
| Package managers | 13 | 10+ |
| Update strategy | Latest version | Configurable |
| Grouping | One PR per dep | Configurable groups |
| Security updates | No | Yes |
| Commit signing | No | Yes |
| Customizable | Fully (it's your code) | Via config file |

This is a starting point. For production use, you'd want to add:

- Version constraints (don't update major versions automatically)
- Grouping related updates into single PRs
- Rate limiting to avoid too many PRs
- Better error handling and logging
- Integration with your CI to verify updates don't break things
