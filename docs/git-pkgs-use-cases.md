# git-pkgs Integration with managers

This document describes how git-pkgs could use the managers library to add dependency management commands alongside its existing analysis capabilities.

## Current state

git-pkgs currently provides read-only analysis of dependencies across git history:

- `git-pkgs outdated` - find packages with newer versions (via ecosyste.ms API)
- `git-pkgs list` - list dependencies at any commit
- `git-pkgs vulns` - check for security vulnerabilities
- `git-pkgs tree` - show dependency tree
- `git-pkgs licenses` - audit licenses
- `git-pkgs diff` - compare dependencies between commits

These commands parse lockfiles (using the manifests library) and query external APIs, but they don't modify anything.

## What managers enables

The managers library lets git-pkgs run package manager commands without knowing the specifics of each CLI.

### git-pkgs install

Install all dependencies from the lockfile.

```bash
$ git-pkgs install
Detected: bundler (Gemfile.lock)
Running: bundle install
```

With frozen lockfile for CI:

```bash
$ git-pkgs install --frozen
Detected: npm (package-lock.json)
Running: npm ci
```

### git-pkgs update

Update a specific package:

```bash
$ git-pkgs update lodash
Detected: npm (package-lock.json)
Running: npm update lodash
```

Or apply updates found by `git-pkgs outdated`:

```bash
$ git-pkgs outdated
Found 3 outdated dependencies:
  rails 7.0.0 -> 7.1.0 (minor)
  lodash 4.17.20 -> 4.17.21 (patch)
  serde 1.0.190 -> 1.0.195 (patch)

$ git-pkgs update --patch-only
Updating lodash...
Running: npm update lodash
Updating serde...
Running: cargo update serde
Done. Updated 2 packages.
```

Options:
- `--patch-only` - only apply patch updates (safe)
- `--minor` - apply patch and minor updates
- `--all` - apply all updates including major (careful)
- `--dry-run` - show what would happen

### git-pkgs add

Add a package to any project, regardless of ecosystem.

```bash
$ git-pkgs add lodash
Detected: pnpm (pnpm-lock.yaml)
Running: pnpm add lodash

$ git-pkgs add rails --dev
Detected: bundler (Gemfile.lock)
Running: bundle add rails --group development
```

### git-pkgs remove

```bash
$ git-pkgs remove lodash
Detected: yarn (yarn.lock)
Running: yarn remove lodash
```

### git-pkgs sync

Ensure dependencies are installed and lockfile is current.

```bash
$ git-pkgs sync
Detected: gomod (go.sum)
Running: go mod download
Running: go mod tidy
```

## Implementation approach

The integration is straightforward because git-pkgs already detects ecosystems from lockfiles:

```go
// git-pkgs already knows the ecosystem from parsing
dep := database.Dependency{
    Ecosystem: "npm",
    Name:      "lodash",
}

// Map ecosystem to manager (handles npm/pnpm/yarn/bun)
manager := detectManagerFromLockfiles(repoPath)

// Build and run the command
cmd, _ := translator.BuildCommand(manager, "update", managers.CommandInput{
    Args: map[string]string{"package": dep.Name},
})
runner.Run(ctx, cmd, managers.RunOptions{Dir: repoPath})
```

The key mapping is ecosystem (from ecosyste.ms/manifests) to manager (from managers library):

| Ecosystem | Possible managers | Detection |
|-----------|------------------|-----------|
| npm | npm, pnpm, yarn, bun | Lockfile present |
| rubygems | bundler | Gemfile.lock |
| cargo | cargo | Cargo.lock |
| go | gomod | go.sum |
| pypi | uv, poetry | uv.lock or poetry.lock |
| packagist | composer | composer.lock |
| hex | mix | mix.lock |
| pub | pub | pubspec.lock |
| cocoapods | cocoapods | Podfile.lock |

## Multi-ecosystem repos

git-pkgs already handles repos with multiple ecosystems (e.g., a Rails app with npm for frontend). These commands would operate on all detected managers:

```bash
$ git-pkgs install
Detected: bundler (Gemfile.lock), npm (package-lock.json)
Running: bundle install
Running: npm install
```

Or target a specific one:

```bash
$ git-pkgs install --ecosystem npm
Running: npm install
```

## Git hooks integration

git-pkgs already has a hooks system. With managers, it could:

**Post-checkout hook** - auto-install when switching branches with lockfile changes:

```bash
$ git checkout feature-branch
Lockfile changed. Running: bundle install
```

**Pre-commit hook** - verify lockfile is in sync:

```bash
$ git commit
Checking lockfile integrity...
Running: npm install --frozen-lockfile
Error: lockfile out of sync with package.json
```

## Workflow example

A typical workflow combining existing git-pkgs features with new deps commands:

```bash
# Check current state
$ git-pkgs outdated
Found 5 outdated dependencies

# Check for vulnerabilities in updates
$ git-pkgs vulns --include-outdated
No known vulnerabilities in available updates

# Apply safe updates
$ git-pkgs update --patch-only
Updated 3 packages

# Review what changed
$ git-pkgs diff HEAD~1
+ lodash 4.17.21 (was 4.17.20)
+ axios 1.6.1 (was 1.6.0)
+ serde 1.0.195 (was 1.0.190)

# Commit
$ git commit -am "Update patch dependencies"
```
