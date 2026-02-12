# managers

A Go library that wraps package manager CLIs behind a common interface. Part of the [git-pkgs](https://github.com/git-pkgs) project.

## What it does

Translates generic operations (install, add, remove, list, outdated, update, vendor, resolve) into the correct CLI commands for each package manager. Define what you want to do once, and the library figures out the right command for npm, bundler, cargo, go, or any other supported manager.

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

**Source exploration** - Open the source code of any installed dependency in your editor. The `path` operation returns the filesystem location regardless of whether it's in node_modules, site-packages, or a cargo registry.

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

Most managers support: install, add, remove, list, outdated, update, resolve. Some also support vendor and path. Some managers (maven, gradle, sbt, lein) have limited CLI support for add/remove operations.

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

### Policies

PolicyRunner wraps a Runner and applies checks before commands execute. Use this to enforce security policies, license compliance, or package blocklists.

```go
// Create a policy runner that wraps the real executor
runner := managers.NewPolicyRunner(
    managers.NewExecRunner(),
    managers.WithPolicyMode(managers.PolicyEnforce),
)

// Add policies
runner.AddPolicy(managers.PackageBlocklistPolicy{
    Blocked: map[string]string{
        "event-stream": "compromised in 2018",
    },
})

// Commands are checked before execution
result, err := runner.Run(ctx, "/path/to/project", "npm", "install", "event-stream")
// Returns ErrPolicyViolation
```

The Policy interface:

```go
type Policy interface {
    Name() string
    Check(ctx context.Context, op *PolicyOperation) (*PolicyResult, error)
}
```

PolicyOperation contains the manager name, operation, packages, flags, and the full command. PolicyResult indicates whether to allow or deny, with an optional reason and warnings.

Three modes control enforcement:
- `PolicyEnforce` - block operations that fail checks
- `PolicyWarn` - log warnings but allow operations to proceed
- `PolicyDisabled` - skip all policy checks

Built-in policies include AllowAllPolicy, DenyAllPolicy, and PackageBlocklistPolicy. Implement the Policy interface for custom checks like vulnerability scanning or license validation.

## Operations

| Operation | Description |
|-----------|-------------|
| `install` | Install dependencies from lockfile |
| `add` | Add a new dependency |
| `remove` | Remove a dependency |
| `list` | List installed packages |
| `outdated` | Show packages with available updates |
| `update` | Update dependencies |
| `path` | Get filesystem path to installed package |
| `vendor` | Copy dependencies into the project directory |
| `resolve` | Produce dependency graph output from the local CLI |

### Common flags

| Flag | Description |
|------|-------------|
| `dev` | Add as development dependency |
| `frozen` | Fail if lockfile would change (CI mode) |
| `json` | Output in JSON format (where supported) |

### Getting package paths

The `path` operation returns the filesystem path to an installed package, useful for source exploration or editor integration:

```go
manager, _ := managers.Detect("/path/to/project")
result, _ := manager.Path(ctx, "lodash")
fmt.Println(result.Path) // "/path/to/project/node_modules/lodash"
```

The library handles extracting clean paths from various output formats (JSON, line-based, regex patterns). For managers with predictable locations (yarn, mix, shards), paths are computed from templates.

**Managers with path support:** npm, pnpm, yarn, bun, bundler, gem, pip, uv, poetry, conda, gomod, cargo, composer, brew, deno, nimble, opam, luarocks, conan, mix, shards, rebar3

### Vendoring dependencies

The `vendor` operation copies dependencies into the project directory for offline builds or source inspection:

```go
manager, _ := managers.Detect("/path/to/project")
result, _ := manager.Vendor(ctx)
```

**Managers with vendor support:** gomod, cargo, bundler, pip, rebar3

### Resolving dependency graphs

The `resolve` operation runs the package manager's dependency graph command and returns raw output. Some managers produce JSON (npm, cargo, pip), others produce text trees (go, maven, poetry). Parsing and normalization is left to the caller.

```go
manager, _ := managers.Detect("/path/to/project")
result, _ := manager.Resolve(ctx)
fmt.Println(result.Stdout) // raw CLI output (JSON tree, text tree, etc.)
```

**Managers with resolve support:** npm, pnpm, yarn, bun, bundler, cargo, gomod, pip, uv, poetry, conda, composer, maven, gradle, lein, swift, deno, stack, pub, mix, rebar3, nuget, conan, helm

### Escape hatch

For manager-specific flags not covered by the common interface, use `Extra`:

```go
cmd, _ := translator.BuildCommand("npm", "install", managers.CommandInput{
    Flags: map[string]any{"frozen": true},
    Extra: []string{"--legacy-peer-deps"},
})
// Result: ["npm", "install", "--ci", "--legacy-peer-deps"]
```

## Configuration files

This library builds and executes CLI commands. It doesn't read or modify package manager configuration files. When commands run, they inherit the environment and respect native config files:

- npm/yarn/pnpm: `.npmrc`, `~/.npmrc`
- pip/poetry/uv: `pip.conf`, `.pypirc`, `pyproject.toml`
- bundler: `~/.bundle/config`, `.bundle/config`
- cargo: `~/.cargo/config.toml`, `.cargo/config.toml`
- composer: `auth.json`, `config.json`
- go: `GOPROXY`, `GOPRIVATE` environment variables

Private registries, proxy servers (like Artifactory), scoped registries, and credentials all work as configured for the underlying tool. The library just builds the right command; the CLI handles authentication and registry resolution.

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
