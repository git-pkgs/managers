# managers

A Go library that wraps package manager CLIs behind a common interface. Part of the [git-pkgs](https://github.com/git-pkgs) project.

## What it does

Translates generic operations (install, add, remove, list, outdated, update) into the correct CLI commands for each package manager. Define what you want to do once, and the library figures out the right command for npm, bundler, cargo, go, or any other supported manager.

```go
translator := managers.NewTranslator()
// ... register definitions

// Same operation, different managers
cmd, _ := translator.BuildCommand("npm", "add", managers.CommandInput{
    Args: map[string]string{"package": "lodash"},
    Flags: map[string]any{"dev": true},
})
// Result: ["npm", "install", "lodash", "--save-dev"]

cmd, _ = translator.BuildCommand("bundler", "add", managers.CommandInput{
    Args: map[string]string{"package": "rails"},
    Flags: map[string]any{"dev": true},
})
// Result: ["bundle", "add", "rails", "--group", "development"]
```

## Use cases

**CI/CD tooling** - A unified "install dependencies" step that works regardless of what language the repo uses. Detect the lockfile, run the right install command with the frozen flag.

**Monorepo orchestration** - Walk subdirectories, detect package managers, run operations in parallel. Update dependencies across many services without maintaining separate configs for each ecosystem.

**Security scanners** - After finding a vulnerability, automatically generate the update command or apply it. The scanner doesn't need to know npm from cargo.

**IDE plugins** - "Add package" dialog that works for any project type. User types a package name, plugin detects the manager and runs the right command.

**Dependency updaters** - Build your own Dependabot. Check for outdated packages, create branches, apply updates, open PRs. See the [dependabot-cron example](docs/examples/dependabot-cron/).

**git-pkgs integration** - Add `install`, `update`, `add`, `remove` commands to git-pkgs. See [git-pkgs use cases](docs/git-pkgs-use-cases.md).

**Audit and compliance** - "Show me all outdated packages across all our repos" for a fleet of projects in different languages. Normalize the output format.

## Supported package managers

| Manager | Ecosystem | Lockfile |
|---------|-----------|----------|
| npm | npm | package-lock.json |
| pnpm | npm | pnpm-lock.yaml |
| yarn | npm | yarn.lock |
| bun | npm | bun.lock |
| deno | deno | deno.lock |
| bundler | gem | Gemfile.lock |
| gem | gem | - |
| cargo | cargo | Cargo.lock |
| gomod | go | go.sum |
| pip | pypi | requirements.txt |
| uv | pypi | uv.lock |
| poetry | pypi | poetry.lock |
| conda | conda | conda-lock.yml |
| composer | packagist | composer.lock |
| mix | hex | mix.lock |
| rebar3 | hex | rebar.lock |
| pub | pub | pubspec.lock |
| cocoapods | cocoapods | Podfile.lock |
| swift | swift | Package.resolved |
| nuget | nuget | packages.lock.json |
| maven | maven | - |
| gradle | maven | gradle.lockfile |
| sbt | maven | - |
| cabal | hackage | cabal.project.freeze |
| stack | hackage | stack.yaml.lock |
| opam | opam | opam.locked |
| luarocks | luarocks | - |
| nimble | nimble | nimble.lock |
| shards | shards | shard.lock |
| cpanm | cpan | cpanfile.snapshot |
| lein | clojars | - |
| vcpkg | vcpkg | vcpkg.json |
| conan | conan | conan.lock |
| helm | helm | Chart.lock |
| brew | homebrew | - |

Most managers support: install, add, remove, list, outdated, update. Some managers (maven, gradle, sbt, lein) have limited CLI support for add/remove operations.

## Installation

```bash
go get github.com/git-pkgs/managers
```

## Usage

### Building commands

```go
package main

import (
    "fmt"
    "github.com/git-pkgs/managers"
    "github.com/git-pkgs/managers/definitions"
)

func main() {
    // Load embedded definitions
    defs, _ := definitions.LoadEmbedded()

    // Create translator and register definitions
    translator := managers.NewTranslator()
    for _, def := range defs {
        translator.Register(def)
    }

    // Build a command
    cmd, err := translator.BuildCommand("npm", "add", managers.CommandInput{
        Args: map[string]string{
            "package": "lodash",
            "version": "4.17.21",
        },
        Flags: map[string]any{
            "dev": true,
        },
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(cmd) // ["npm", "install", "lodash@4.17.21", "--save-dev"]
}
```

### Command chaining

Some operations require multiple commands. Use `BuildCommands` to get all of them:

```go
// Go's add operation runs "go get" then "go mod tidy"
cmds, _ := translator.BuildCommands("gomod", "add", managers.CommandInput{
    Args: map[string]string{"package": "github.com/pkg/errors"},
})
// cmds[0] = ["go", "get", "github.com/pkg/errors"]
// cmds[1] = ["go", "mod", "tidy"]
```

### Executing commands

The library builds commands but doesn't execute them by default. Use the Runner interface:

```go
runner := managers.NewExecRunner()
result, err := runner.Run(ctx, cmd, managers.RunOptions{
    Dir: "/path/to/project",
})
```

Or use MockRunner for testing:

```go
mock := managers.NewMockRunner()
mock.AddResult(managers.Result{
    ExitCode: 0,
    Stdout:   []byte(`{"dependencies": {}}`),
})
```

## Operations

| Operation | Description |
|-----------|-------------|
| `install` | Install dependencies from lockfile |
| `add` | Add a new dependency |
| `remove` | Remove a dependency |
| `list` | List installed packages |
| `outdated` | Show packages with available updates |
| `update` | Update dependencies |

### Common flags

| Flag | Description |
|------|-------------|
| `dev` | Add as development dependency |
| `frozen` | Fail if lockfile would change (CI mode) |
| `json` | Output in JSON format (where supported) |

### Escape hatch

For manager-specific flags not covered by the common interface, use `Extra`:

```go
cmd, _ := translator.BuildCommand("npm", "install", managers.CommandInput{
    Flags: map[string]any{"frozen": true},
    Extra: []string{"--legacy-peer-deps"},
})
// Result: ["npm", "install", "--ci", "--legacy-peer-deps"]
```

## How it works

Package managers are defined in YAML files that describe their commands, flags, and arguments. The translator reads these definitions and builds the correct command array for each operation.

```yaml
# definitions/npm.yaml
name: npm
binary: npm
commands:
  add:
    base: [install]
    args:
      package: {position: 0, required: true}
      version: {suffix: "@"}
    flags:
      dev: [--save-dev]
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on adding new package managers.

## Testing

```bash
go test ./...
```

Tests verify command construction without executing real CLIs. Each test compares the built command array against expected output.

## Related projects

This library is part of a toolkit for building dependency-aware tools:

- [manifests](https://github.com/git-pkgs/manifests) - Parse lockfiles and manifests to read dependency state
- [vers](https://github.com/git-pkgs/vers) - Version range parsing and comparison
- [git-pkgs](https://github.com/git-pkgs/git-pkgs) - CLI for tracking dependency history

The **manifests** library reads state (parse lockfiles, extract dependency trees) while **managers** runs CLI commands that modify state. The CLIs themselves update the lockfiles. Together they cover the full lifecycle:

```go
// Read current state
deps := manifests.Parse("Gemfile.lock")

// Check what's outdated
cmd, _ := managers.BuildCommand("bundler", "outdated", input)
runner.Run(ctx, cmd, opts)

// Apply update (CLI modifies Gemfile.lock)
cmd, _ = managers.BuildCommand("bundler", "update", managers.CommandInput{
    Args: map[string]string{"package": "rails"},
})
runner.Run(ctx, cmd, opts)

// Read new state
deps = manifests.Parse("Gemfile.lock")
```
