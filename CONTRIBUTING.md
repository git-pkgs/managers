# Contributing

## Adding a new package manager

Adding support for a new package manager involves three steps: capturing reference material, writing a definition file, and adding tests.

### 1. Capture CLI help

First, capture the help output for reference. This serves as documentation and helps verify commands later.

```bash
mkdir -p references/mymanager
cd references/mymanager

mymanager --version > version.txt
mymanager --help > help.txt
mymanager install --help > install.txt
mymanager add --help > add.txt
# ... capture help for each command you'll support
```

### 2. Write the definition file

Create `definitions/mymanager.yaml`. Here's the structure:

```yaml
name: mymanager
ecosystem: myecosystem  # npm, pypi, cargo, gem, etc.
binary: mymanager       # the CLI binary name
version: ">=1.0.0"      # minimum supported version
status: current
min_tested: "1.0.0"
max_tested: "2.0.0"

detection:
  lockfiles:
    - mymanager.lock
  manifests:
    - mymanager.toml
  priority: 10  # higher = preferred when multiple managers detected

version_detection:
  command: [--version]
  pattern: '(\d+\.\d+\.\d+)'

commands:
  install:
    base: [install]
    flags:
      frozen: [--frozen-lockfile]
      production: [--prod]
    exit_codes:
      0: success
      1: error

  add:
    base: [add]
    args:
      package: {position: 0, required: true, validate: package_name}
      version: {suffix: "@"}  # for package@version syntax
    flags:
      dev: [--dev]
      optional: [--optional]
    exit_codes:
      0: success
      1: error

  remove:
    base: [remove]
    args:
      package: {position: 0, required: true}
    exit_codes:
      0: success
      1: error

  list:
    base: [list]
    flags:
      json: [--json]
    default_flags: [--json]
    exit_codes:
      0: success
      1: error

  outdated:
    base: [outdated]
    flags:
      json: [--json]
    default_flags: [--json]
    exit_codes:
      0: success
      1: outdated

  update:
    base: [update]
    args:
      package: {position: 0, required: false}
    exit_codes:
      0: success
      1: error

capabilities:
  - install
  - install_frozen
  - add
  - add_dev
  - remove
  - update
  - list
  - outdated
  - json_output
```

### Schema reference

**Args:**

| Field | Description |
|-------|-------------|
| `position` | Positional order (0-indexed) |
| `required` | Whether the arg must be provided |
| `validate` | Validator name (npm_package, gem_name, etc.) |
| `flag` | Use a flag instead of positional (`--version VALUE`) |
| `suffix` | Append to previous arg (`@` for `pkg@version`) |
| `fixed_suffix` | Always append this value (`@none` for Go remove) |

**Flags:**

Flags can be simple arrays or complex structures:

```yaml
# Simple: just add these strings
dev: [--save-dev]

# With value: include field reference
workspace: [--workspace, {value: workspace}]

# With join: for --flag=value syntax
group: [--group, {value: group_name, join: "="}]
```

**Command chaining:**

Some operations need multiple commands:

```yaml
add:
  base: [get]
  args:
    package: {position: 0, required: true}
  then:
    - base: [mod, tidy]  # runs after main command
```

### 3. Add tests

Add tests to `translator_test.go`:

```go
func TestMymanagerInstall(t *testing.T) {
    tr := loadTranslator(t)
    cmd, err := tr.BuildCommand("mymanager", "install", CommandInput{})
    if err != nil {
        t.Fatalf("BuildCommand failed: %v", err)
    }
    expected := []string{"mymanager", "install"}
    if !reflect.DeepEqual(cmd, expected) {
        t.Errorf("got %v, want %v", cmd, expected)
    }
}

func TestMymanagerAdd(t *testing.T) {
    tr := loadTranslator(t)
    cmd, err := tr.BuildCommand("mymanager", "add", CommandInput{
        Args: map[string]string{"package": "some-package"},
    })
    if err != nil {
        t.Fatalf("BuildCommand failed: %v", err)
    }
    expected := []string{"mymanager", "add", "some-package"}
    if !reflect.DeepEqual(cmd, expected) {
        t.Errorf("got %v, want %v", cmd, expected)
    }
}

func TestMymanagerAddDev(t *testing.T) {
    tr := loadTranslator(t)
    cmd, err := tr.BuildCommand("mymanager", "add", CommandInput{
        Args:  map[string]string{"package": "some-package"},
        Flags: map[string]any{"dev": true},
    })
    if err != nil {
        t.Fatalf("BuildCommand failed: %v", err)
    }
    expected := []string{"mymanager", "add", "some-package", "--dev"}
    if !reflect.DeepEqual(cmd, expected) {
        t.Errorf("got %v, want %v", cmd, expected)
    }
}
```

Run tests:

```bash
go test ./... -v
```

## Package managers to add

Tier 1 (high priority):

| Manager | Ecosystem | Notes |
|---------|-----------|-------|
| yarn | npm | Classic (v1) and Berry (v2+) have different commands |
| bun | npm | Fast npm-compatible runtime |
| poetry | pypi | Popular Python project manager |
| pip | pypi | Basic Python installer |
| pdm | pypi | PEP 582 Python manager |
| composer | packagist | PHP |

Tier 2:

| Manager | Ecosystem | Notes |
|---------|-----------|-------|
| gem | rubygems | Ruby (bundler preferred for projects) |
| dotnet | nuget | .NET CLI |
| maven | maven | Java (XML-based, may need special handling) |
| gradle | maven | Java/Kotlin build tool |
| pub | pub.dev | Dart/Flutter |
| mix | hex | Elixir |
| deno | deno.land | Deno JavaScript runtime |

Tier 3:

| Manager | Ecosystem | Notes |
|---------|-----------|-------|
| cabal | hackage | Haskell |
| stack | hackage | Haskell (alternative to cabal) |
| opam | opam | OCaml |
| leiningen | clojars | Clojure |
| rebar3 | hex | Erlang |
| swift | swiftpm | Swift Package Manager |
| conan | conan | C/C++ |
| vcpkg | vcpkg | C/C++ (Microsoft) |
| helm | helm | Kubernetes charts |
| terraform | terraform | Infrastructure modules |

## Version testing with Docker

To test against specific CLI versions, we use Docker. Each package manager can have version-specific Dockerfiles:

```
docker/
  npm/
    Dockerfile.7
    Dockerfile.10
  bundler/
    Dockerfile.2.0
    Dockerfile.2.5
```

Example Dockerfile:

```dockerfile
FROM node:20
# npm comes with node, specific version can be installed:
RUN npm install -g npm@10.2.0
WORKDIR /app
```

Run tests against a specific version:

```bash
docker build -t managers-npm-10 -f docker/npm/Dockerfile.10 .
docker run -v $(pwd):/app managers-npm-10 go test ./...
```

## Code style

- Run `go fmt` before committing
- Tests should verify command construction, not execute real CLIs
- Keep definitions minimal - only include flags that are commonly used
- Document any non-obvious command mappings in YAML comments
