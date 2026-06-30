# Replacing a dependency

The `replace` operation points a dependency at something other than its registry version: a local checkout on disk, a git URL at a specific ref, or back to the registry by dropping the override. The main consumer is [downstream](https://github.com/git-pkgs/downstream), which clones a library's dependents, swaps the library for a patched copy, and reruns their tests to find breakage before a release. The same primitive is useful for any "develop against a local fork" workflow or for `git pkgs replace` in the CLI.

```go
mgr, _ := detector.Detect(dependentDir, managers.DetectOptions{})
result, err := mgr.Replace(ctx, "github.com/spf13/cobra", managers.ReplaceOptions{
    Path: "../cobra",
})
```

`ReplaceOptions` takes exactly one of `Path`, `Git`, or `Drop`. `Ref` is a branch, tag, or revision and is only valid alongside `Git`. `Extra` passes raw arguments through to the underlying CLI where there is one.

Capabilities `replace_path`, `replace_git`, and `replace_drop` declare which targets a manager supports. Check with `mgr.Supports(managers.CapReplacePath)` before calling, or just call and handle `ErrUnsupportedOperation`.

## Per-manager behaviour

Most package managers expose this as a CLI command. Three (cargo, uv, bundler) only support source overrides through manifest syntax, so for those `Replace` edits the manifest file directly and `Result.Command` is nil. Either way `Replace` only redirects the source; run `Install` afterwards to actually resolve and fetch.

| Manager | Path | Git | Drop | Mechanism |
|---|---|---|---|---|
| gomod | yes | yes | yes | `go mod edit -replace` / `-dropreplace` |
| npm | yes | yes | no | `npm install pkg@file:...` / `pkg@git+url#ref` |
| pnpm | yes | yes | no | `pnpm add pkg@file:...` / `pkg@git+url#ref` |
| yarn | yes | yes | no | `yarn add pkg@file:...` / `pkg@git+url#ref` |
| bun | yes | yes | no | `bun add pkg@file:...` / `pkg@git+url#ref` |
| composer | yes | yes | yes | `composer config repositories.<key>` then `require` |
| cargo | yes | yes | yes | edits `Cargo.toml` `[patch.crates-io]` |
| uv | yes | yes | yes | edits `pyproject.toml` `[tool.uv.sources]` |
| bundler | yes | yes | yes | edits the `gem` line in `Gemfile` |

Anything not listed returns `ErrUnsupportedOperation`. pip has no override mechanism short of `pip install -e`, which changes what's installed rather than what's declared; poetry and the JVM managers are doable but not wired up yet.

### gomod

```go
mgr.Replace(ctx, "github.com/spf13/cobra", managers.ReplaceOptions{Path: "../cobra"})
// go mod edit -replace github.com/spf13/cobra=../cobra

mgr.Replace(ctx, "github.com/spf13/cobra", managers.ReplaceOptions{
    Git: "https://github.com/fork/cobra",
    Ref: "v1.8.1-0.20240520123456-abcdef123456",
})
// go mod edit -replace github.com/spf13/cobra=github.com/fork/cobra@v1.8.1-0.20240520...

mgr.Replace(ctx, "github.com/spf13/cobra", managers.ReplaceOptions{Drop: true})
// go mod edit -dropreplace github.com/spf13/cobra
```

The git URL is normalised to a module path (scheme and `.git` stripped, `git@host:path` converted). `Ref` must be a Go module version or pseudo-version; `go.mod` replace directives don't accept branch names, and `Replace` rejects them rather than letting `go mod edit` fail later. To target a branch, resolve it first:

```bash
go list -m github.com/fork/cobra@my-branch
# github.com/fork/cobra v1.8.1-0.20240520123456-abcdef123456
```

or clone the fork and use `Path`. For downstream testing where the patched library is already checked out locally, `Path` is the normal route anyway.

### npm, pnpm, yarn, bun

```go
mgr.Replace(ctx, "lodash", managers.ReplaceOptions{Path: "../lodash"})
// npm install lodash@file:../lodash

mgr.Replace(ctx, "lodash", managers.ReplaceOptions{
    Git: "https://github.com/fork/lodash",
    Ref: "fix-something",
})
// npm install lodash@git+https://github.com/fork/lodash#fix-something
```

This rewrites the `dependencies` entry in `package.json`, which means it only redirects a direct dependency. If the package you're replacing is pulled in transitively, the install will succeed but the registry copy stays in the tree. npm `overrides`, pnpm `pnpm.overrides`, and yarn `resolutions` can redirect transitives, but those need `package.json` edits and aren't covered here yet.

`Drop` is not supported because installing from `file:` overwrites the original version constraint and there's no record of what it was. To restore the registry version, call `Add` with an explicit version.

For path replacements the local package usually needs to be built first (`npm run build` or whatever its `prepare` script does); `Replace` doesn't run that for you.

### composer

```go
mgr.Replace(ctx, "monolog/monolog", managers.ReplaceOptions{Path: "../monolog"})
// composer config repositories.git-pkgs-monolog-monolog '{"type":"path","url":"../monolog"}'

mgr.Replace(ctx, "monolog/monolog", managers.ReplaceOptions{
    Git: "https://github.com/fork/monolog",
    Ref: "feature",
})
// composer config repositories.git-pkgs-monolog-monolog '{"type":"vcs","url":"..."}'
// composer require monolog/monolog:dev-feature
```

The repository entry is keyed `git-pkgs-<vendor>-<package>` so `Drop` can find it again. A path repository on its own only adds a candidate source; whether composer actually picks it depends on the version in the local `composer.json` satisfying the dependent's constraint. With a git ref, `Replace` also runs `composer require pkg:dev-<ref>` to force the constraint to the branch.

### cargo

```go
mgr.Replace(ctx, "serde", managers.ReplaceOptions{Path: "../serde"})
```

writes to `Cargo.toml`:

```toml
[patch.crates-io]
serde = { path = "../serde" }
```

A git ref that looks like a hex SHA (7 to 40 chars) is written as `rev`, anything else as `branch`. `[patch]` overrides apply across the whole dependency tree, including transitives, which is what you want for downstream testing. `Drop` removes the entry and, if it was the last one, the section header.

The editor is line-based and only touches the `[patch.crates-io]` table. It will misbehave if that table contains a multi-line value for the target key, which in practice it never does for patch entries.

### uv

Same shape as cargo, written to `pyproject.toml` under `[tool.uv.sources]`:

```toml
[tool.uv.sources]
demo-pkg = { path = "../demo" }
```

uv has `[tool.uv.sources]` as its documented override mechanism; `uv add --editable` exists but changes the declared dependency rather than overlaying it.

### bundler

```go
mgr.Replace(ctx, "rails", managers.ReplaceOptions{Path: "../rails"})
```

rewrites the existing line in `Gemfile`:

```ruby
gem "rails", "~> 7.0", require: false           # before
gem "rails", "~> 7.0", require: false, path: "../rails"   # after
```

The version constraint, `require:`, group options, and any trailing comment are preserved; existing `path:`, `git:`, `github:`, `branch:`, `tag:`, `ref:` keys are stripped before the new source is appended. If the gem isn't in the Gemfile at all a new line is added at the end. `Drop` strips the source keys and leaves the rest, so the registry constraint comes back into effect.

Because the version constraint stays, bundler will refuse to resolve if the local checkout's gemspec version falls outside it. For ephemeral downstream testing that's usually fine since the patch is a small change on top of a release; for a major-version branch you may need to widen the constraint in the dependent's Gemfile yourself.

The editor handles the common single-line `gem "name", ...` form. A gem declared inside a `git ... do` or `path ... do` block, or split across lines, won't be matched.

## Calling the translator directly

`Replace` on a `Manager` is the high-level entry point. If you're driving the translator yourself (the way `git pkgs` does for `--dry-run` output), the command shape per manager is:

```go
// gomod
tr.BuildCommand("gomod", "replace", managers.CommandInput{
    Args: map[string]string{"spec": "github.com/spf13/cobra=../cobra"},
})
// [go mod edit -replace github.com/spf13/cobra=../cobra]

tr.BuildCommand("gomod", "replace", managers.CommandInput{
    Args:  map[string]string{"package": "github.com/spf13/cobra"},
    Flags: map[string]any{"drop": true},
})
// [go mod edit -dropreplace github.com/spf13/cobra]

// npm-family
tr.BuildCommand("pnpm", "replace", managers.CommandInput{
    Args: map[string]string{"spec": "lodash@file:../lodash"},
})
// [pnpm add lodash@file:../lodash]

// composer
tr.BuildCommand("composer", "replace", managers.CommandInput{
    Args: map[string]string{
        "repository": "repositories.git-pkgs-vendor-pkg",
        "payload":    `{"type":"path","url":"../pkg"}`,
    },
})
// [composer config repositories.git-pkgs-vendor-pkg {"type":"path","url":"../pkg"}]
```

cargo, uv and bundler have no `replace` command in the translator since there's no subprocess; use `Manager.Replace` for those.

## Adding a manager

If the manager has a CLI for it, add a `replace` command to its YAML alongside `add`/`remove` and list the supported `replace_*` capabilities. The command should accept a single `spec` arg (the pre-built target string) since the schema can't compose multiple inputs into one argument. Then add a case to `buildReplaceInputs` in `replace.go` that turns `(pkg, ReplaceOptions)` into that spec.

If the manager only supports overrides through manifest syntax, add the capabilities to its YAML, add the manifest filename to `manifestReplaceFiles` in `replace.go`, and add a case to `replaceInManifest`. Reuse `editTOMLReplaceEntry` if the manifest is TOML with a single override section.

Either way, add a translator test in `translator_test.go` for any new YAML command and a `Replace` test in `replace_test.go` driving it through `MockRunner` or a temp file.
