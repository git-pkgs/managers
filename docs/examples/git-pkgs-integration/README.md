# git-pkgs Integration Example

This example shows how the managers library integrates with git-pkgs to add dependency update capabilities.

## The Problem

git-pkgs can detect outdated packages using the ecosyste.ms API:

```bash
$ git-pkgs outdated
Found 3 outdated dependencies:

Minor updates:
  rails 7.0.0 -> 7.1.0

Patch updates:
  lodash 4.17.20 -> 4.17.21
  serde 1.0.190 -> 1.0.195
```

But it can't actually update them. That's where the managers library comes in.

## The Solution

The managers library translates generic "update package X" requests into the correct CLI commands for each package manager:

```go
translator.BuildCommand("npm", "update", managers.CommandInput{
    Args: map[string]string{"package": "lodash"},
})
// Returns: ["npm", "update", "lodash"]

translator.BuildCommand("bundler", "update", managers.CommandInput{
    Args: map[string]string{"package": "rails"},
})
// Returns: ["bundle", "update", "rails"]
```

## Integration Pattern

The key insight is that git-pkgs uses ecosyste.ms ecosystem names (npm, rubygems, cargo, go, pypi, packagist, hex, pub, cocoapods) while our managers library uses package manager names (npm, pnpm, yarn, bun, bundler, cargo, gomod, uv, poetry, composer, mix, pub, cocoapods).

The integration handles this mapping:

| Ecosystem | Manager | Notes |
|-----------|---------|-------|
| npm | npm/pnpm/yarn/bun | Detected from lockfile |
| rubygems | bundler | |
| cargo | cargo | |
| go | gomod | |
| pypi | uv/poetry | Detected from lockfile |
| packagist | composer | |
| hex | mix | |
| pub | pub | |
| cocoapods | cocoapods | |

## Files

- `main.go` - Simple demo showing the concept
- `apply.go` - Production-ready integration with smart lockfile detection

## Usage

The apply command would work like this:

```bash
# Update all outdated packages
git-pkgs deps apply

# Dry run to see what would happen
git-pkgs deps apply --dry-run

# Update only patch versions (safe)
git-pkgs deps apply --update-type=patch

# Update a specific package
git-pkgs deps apply --package lodash
```

## Smart Lockfile Detection

The npm ecosystem presents a challenge: four different package managers (npm, pnpm, yarn, bun) all manage packages from npmjs.org. The integration uses lockfile detection to pick the right one:

| Lockfile | Manager |
|----------|---------|
| bun.lock | bun |
| pnpm-lock.yaml | pnpm |
| yarn.lock | yarn |
| package-lock.json | npm |

This means if git-pkgs reports an outdated "npm" ecosystem package, but the repo has `pnpm-lock.yaml`, the integration will run `pnpm update` rather than `npm update`.

## Future Work

To fully integrate this into git-pkgs:

1. Add managers as a dependency in go.mod
2. Create cmd/deps.go with subcommands (apply, sync, etc.)
3. Wire up the Apply function from this example
4. Add tests using the Docker setup from managers

The managers library handles the per-manager complexity so git-pkgs can focus on its core value: tracking dependencies across git history.
