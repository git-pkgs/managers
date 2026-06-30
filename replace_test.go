package managers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/git-pkgs/managers/definitions"
)

func newReplaceTestManager(t *testing.T, name, dir string, runner Runner) *GenericManager {
	t.Helper()
	defs, err := definitions.LoadEmbedded()
	if err != nil {
		t.Fatalf("load definitions: %v", err)
	}
	tr := NewTranslator()
	var def *definitions.Definition
	for _, d := range defs {
		tr.Register(d)
		if d.Name == name {
			def = d
		}
	}
	if def == nil {
		t.Fatalf("definition %q not found", name)
	}
	return &GenericManager{def: def, dir: dir, translator: tr, runner: runner}
}

func TestReplaceOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		opts    ReplaceOptions
		wantErr string
	}{
		{"requires mode", ReplaceOptions{}, "specify one of"},
		{"rejects path and git", ReplaceOptions{Path: "../x", Git: "https://example.test/x"}, "specify only one"},
		{"rejects path and drop", ReplaceOptions{Path: "../x", Drop: true}, "specify only one"},
		{"ref requires git", ReplaceOptions{Path: "../x", Ref: "main"}, "ref can only be used with a git replacement"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.opts.mode()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestReplaceGoPath(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "gomod", "/test/dependent", runner)

	result, err := mgr.Replace(context.Background(), "example.test/lib", ReplaceOptions{Path: "../lib"})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if !result.Success() {
		t.Fatalf("result = %+v, want success", result)
	}
	want := []string{"go", "mod", "edit", "-replace", "example.test/lib=../lib"}
	if !reflect.DeepEqual(runner.LastCaptured(), want) {
		t.Errorf("got %v, want %v", runner.LastCaptured(), want)
	}
}

func TestReplaceGoGit(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "gomod", "/test/dependent", runner)

	_, err := mgr.Replace(context.Background(), "example.test/lib", ReplaceOptions{
		Git: "https://github.com/fork/lib.git",
		Ref: "v1.2.3",
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	want := []string{"go", "mod", "edit", "-replace", "example.test/lib=github.com/fork/lib@v1.2.3"}
	if !reflect.DeepEqual(runner.LastCaptured(), want) {
		t.Errorf("got %v, want %v", runner.LastCaptured(), want)
	}
}

func TestReplaceGoGitRequiresVersionRef(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "gomod", "/test/dependent", runner)

	_, err := mgr.Replace(context.Background(), "example.test/lib", ReplaceOptions{
		Git: "https://github.com/fork/lib",
		Ref: "feature-branch",
	})
	if err == nil || !strings.Contains(err.Error(), "Go module version") {
		t.Fatalf("error = %v, want Go module version error", err)
	}

	_, err = mgr.Replace(context.Background(), "example.test/lib", ReplaceOptions{
		Git: "https://github.com/fork/lib",
	})
	if err == nil || !strings.Contains(err.Error(), "requires Ref") {
		t.Fatalf("error = %v, want requires Ref error", err)
	}
}

func TestReplaceGoDrop(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "gomod", "/test/dependent", runner)

	_, err := mgr.Replace(context.Background(), "example.test/lib", ReplaceOptions{Drop: true})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	want := []string{"go", "mod", "edit", "-dropreplace", "example.test/lib"}
	if !reflect.DeepEqual(runner.LastCaptured(), want) {
		t.Errorf("got %v, want %v", runner.LastCaptured(), want)
	}
}

func TestReplaceNPMFamily(t *testing.T) {
	tests := []struct {
		manager string
		opts    ReplaceOptions
		want    []string
	}{
		{"npm", ReplaceOptions{Path: "../lodash"}, []string{"npm", "install", "lodash@file:../lodash"}},
		{"pnpm", ReplaceOptions{Path: "../lodash"}, []string{"pnpm", "add", "lodash@file:../lodash"}},
		{"yarn", ReplaceOptions{Path: "../lodash"}, []string{"yarn", "add", "lodash@file:../lodash"}},
		{"bun", ReplaceOptions{Path: "../lodash"}, []string{"bun", "add", "lodash@file:../lodash"}},
		{"npm", ReplaceOptions{Git: "https://github.com/fork/lodash", Ref: "fix"}, []string{"npm", "install", "lodash@git+https://github.com/fork/lodash#fix"}},
		{"pnpm", ReplaceOptions{Git: "git@github.com:fork/lodash.git"}, []string{"pnpm", "add", "lodash@git@github.com:fork/lodash.git"}},
	}
	for _, tt := range tests {
		t.Run(tt.manager+"/"+tt.want[len(tt.want)-1], func(t *testing.T) {
			runner := NewMockRunner()
			mgr := newReplaceTestManager(t, tt.manager, "/test/dependent", runner)
			_, err := mgr.Replace(context.Background(), "lodash", tt.opts)
			if err != nil {
				t.Fatalf("Replace: %v", err)
			}
			if !reflect.DeepEqual(runner.LastCaptured(), tt.want) {
				t.Errorf("got %v, want %v", runner.LastCaptured(), tt.want)
			}
		})
	}
}

func TestReplaceNPMDropUnsupported(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "npm", "/test/dependent", runner)
	_, err := mgr.Replace(context.Background(), "lodash", ReplaceOptions{Drop: true})
	if !errors.Is(err, ErrUnsupportedOperation) {
		t.Fatalf("error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestReplaceComposerPath(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "composer", "/test/dependent", runner)

	_, err := mgr.Replace(context.Background(), "vendor/pkg", ReplaceOptions{Path: "../pkg"})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	want := []string{"composer", "config", "repositories.git-pkgs-vendor-pkg", `{"type":"path","url":"../pkg"}`}
	if !reflect.DeepEqual(runner.LastCaptured(), want) {
		t.Errorf("got %v, want %v", runner.LastCaptured(), want)
	}
}

func TestReplaceComposerGitRefRunsRequire(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "composer", "/test/dependent", runner)

	_, err := mgr.Replace(context.Background(), "vendor/pkg", ReplaceOptions{
		Git:   "https://github.com/fork/pkg",
		Ref:   "feature",
		Extra: []string{"--no-update"},
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if len(runner.Captured) != 2 {
		t.Fatalf("captured %d commands, want 2: %v", len(runner.Captured), runner.Captured)
	}
	wantConfig := []string{"composer", "config", "repositories.git-pkgs-vendor-pkg", `{"type":"vcs","url":"https://github.com/fork/pkg"}`}
	if !reflect.DeepEqual(runner.Captured[0], wantConfig) {
		t.Errorf("config: got %v, want %v", runner.Captured[0], wantConfig)
	}
	wantRequire := []string{"composer", "require", "vendor/pkg:dev-feature", "--no-update"}
	if !reflect.DeepEqual(runner.Captured[1], wantRequire) {
		t.Errorf("require: got %v, want %v", runner.Captured[1], wantRequire)
	}
}

func TestReplaceComposerDrop(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "composer", "/test/dependent", runner)

	_, err := mgr.Replace(context.Background(), "vendor/pkg", ReplaceOptions{Drop: true})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	want := []string{"composer", "config", "--unset", "repositories.git-pkgs-vendor-pkg"}
	if !reflect.DeepEqual(runner.LastCaptured(), want) {
		t.Errorf("got %v, want %v", runner.LastCaptured(), want)
	}
}

func TestReplaceUnsupportedManager(t *testing.T) {
	runner := NewMockRunner()
	mgr := newReplaceTestManager(t, "pip", "/test/dependent", runner)
	_, err := mgr.Replace(context.Background(), "requests", ReplaceOptions{Path: "../requests"})
	if !errors.Is(err, ErrUnsupportedOperation) {
		t.Fatalf("error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestReplaceCargoManifest(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "Cargo.toml")
	writeFile(t, manifest, "[package]\nname = \"app\"\n")

	mgr := newReplaceTestManager(t, "cargo", dir, NewMockRunner())

	_, err := mgr.Replace(context.Background(), "serde", ReplaceOptions{Path: "../serde"})
	if err != nil {
		t.Fatalf("Replace path: %v", err)
	}
	got := readFile(t, manifest)
	if !strings.Contains(got, "[patch.crates-io]\nserde = { path = \"../serde\" }\n") {
		t.Fatalf("Cargo.toml missing patch entry:\n%s", got)
	}

	_, err = mgr.Replace(context.Background(), "serde", ReplaceOptions{Git: "https://github.com/fork/serde", Ref: "main"})
	if err != nil {
		t.Fatalf("Replace git: %v", err)
	}
	got = readFile(t, manifest)
	if !strings.Contains(got, `serde = { git = "https://github.com/fork/serde", branch = "main" }`) {
		t.Fatalf("Cargo.toml missing git patch entry:\n%s", got)
	}
	if strings.Contains(got, `path = "../serde"`) {
		t.Fatalf("Cargo.toml retained old path entry:\n%s", got)
	}

	_, err = mgr.Replace(context.Background(), "serde", ReplaceOptions{Drop: true})
	if err != nil {
		t.Fatalf("Replace drop: %v", err)
	}
	got = readFile(t, manifest)
	if strings.Contains(got, "serde") || strings.Contains(got, "[patch.crates-io]") {
		t.Fatalf("Cargo.toml still contains patch after drop:\n%s", got)
	}
}

func TestReplaceCargoManifestSHARev(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "Cargo.toml")
	writeFile(t, manifest, "[package]\nname = \"app\"\n")

	mgr := newReplaceTestManager(t, "cargo", dir, NewMockRunner())
	_, err := mgr.Replace(context.Background(), "serde", ReplaceOptions{Git: "https://github.com/fork/serde", Ref: "abc1234"})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	got := readFile(t, manifest)
	if !strings.Contains(got, `rev = "abc1234"`) {
		t.Fatalf("Cargo.toml should use rev for SHA:\n%s", got)
	}
}

func TestReplaceUVManifest(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "pyproject.toml")
	writeFile(t, manifest, "[project]\nname = \"app\"\n\n[tool.uv]\n")

	mgr := newReplaceTestManager(t, "uv", dir, NewMockRunner())

	_, err := mgr.Replace(context.Background(), "demo-pkg", ReplaceOptions{Path: "../demo"})
	if err != nil {
		t.Fatalf("Replace path: %v", err)
	}
	got := readFile(t, manifest)
	if !strings.Contains(got, "[tool.uv.sources]\ndemo-pkg = { path = \"../demo\" }\n") {
		t.Fatalf("pyproject.toml missing uv source:\n%s", got)
	}

	_, err = mgr.Replace(context.Background(), "demo-pkg", ReplaceOptions{Drop: true})
	if err != nil {
		t.Fatalf("Replace drop: %v", err)
	}
	got = readFile(t, manifest)
	if strings.Contains(got, "demo-pkg") || strings.Contains(got, "[tool.uv.sources]") {
		t.Fatalf("pyproject.toml still contains uv source:\n%s", got)
	}
	if !strings.Contains(got, "[tool.uv]") {
		t.Fatalf("pyproject.toml lost unrelated section:\n%s", got)
	}
}

func TestReplaceGemfile(t *testing.T) {
	dir := t.TempDir()
	gemfile := filepath.Join(dir, "Gemfile")
	writeFile(t, gemfile, `source "https://rubygems.org"`+"\n"+`gem "rails", "~> 7.0", require: false # pinned`+"\n")

	mgr := newReplaceTestManager(t, "bundler", dir, NewMockRunner())

	_, err := mgr.Replace(context.Background(), "rails", ReplaceOptions{Git: "https://github.com/fork/rails", Ref: "main"})
	if err != nil {
		t.Fatalf("Replace git: %v", err)
	}
	got := readFile(t, gemfile)
	want := `gem "rails", "~> 7.0", require: false, git: "https://github.com/fork/rails", branch: "main" # pinned`
	if !strings.Contains(got, want) {
		t.Fatalf("Gemfile = %q, want containing %q", got, want)
	}

	_, err = mgr.Replace(context.Background(), "rails", ReplaceOptions{Drop: true})
	if err != nil {
		t.Fatalf("Replace drop: %v", err)
	}
	got = readFile(t, gemfile)
	if !strings.Contains(got, `gem "rails", "~> 7.0", require: false # pinned`) {
		t.Fatalf("Gemfile did not restore original line after drop:\n%s", got)
	}
	if strings.Contains(got, "git:") || strings.Contains(got, "branch:") {
		t.Fatalf("Gemfile still has source args after drop:\n%s", got)
	}
}

func TestReplaceGemfileFiltersExistingSourceArgs(t *testing.T) {
	line := `gem "rails", "~> 7.0", github: "rails/rails", :branch => "main", require: false` + "\n"
	got, err := updateGemfileLine(line, "rails", replaceModeGit, ReplaceOptions{Git: "https://github.com/fork/rails", Ref: "feature"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "github:") || strings.Contains(got, ":branch =>") {
		t.Fatalf("retained old source args: %s", got)
	}
	want := `gem "rails", "~> 7.0", require: false, git: "https://github.com/fork/rails", branch: "feature"`
	if !strings.Contains(got, want) {
		t.Fatalf("got %q, want containing %q", got, want)
	}
}

func TestReplaceGemfileAppendsMissingGem(t *testing.T) {
	dir := t.TempDir()
	gemfile := filepath.Join(dir, "Gemfile")
	writeFile(t, gemfile, `source "https://rubygems.org"`+"\n")

	mgr := newReplaceTestManager(t, "bundler", dir, NewMockRunner())
	_, err := mgr.Replace(context.Background(), "rake", ReplaceOptions{Path: "../rake"})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	got := readFile(t, gemfile)
	if !strings.Contains(got, `gem "rake", path: "../rake"`) {
		t.Fatalf("Gemfile missing appended gem:\n%s", got)
	}
}

func TestUpdateTOMLSectionEntry(t *testing.T) {
	in := "[patch.crates-io] # local overrides\nfoo = { path = \"../foo\" }\n\n[other]\nx = 1\n"

	out := updateTOMLSectionEntry(in, "[patch.crates-io]", "bar", `{ path = "../bar" }`, false)
	if !strings.Contains(out, "[patch.crates-io] # local overrides") {
		t.Fatalf("lost commented header:\n%s", out)
	}
	if !strings.Contains(out, "foo = { path = \"../foo\" }") {
		t.Fatalf("lost existing entry:\n%s", out)
	}
	if !strings.Contains(out, "bar = { path = \"../bar\" }") {
		t.Fatalf("missing new entry:\n%s", out)
	}
	if !strings.Contains(out, "[other]\nx = 1") {
		t.Fatalf("lost trailing section:\n%s", out)
	}

	out = updateTOMLSectionEntry(out, "[patch.crates-io]", "foo", "", true)
	if strings.Contains(out, "foo =") {
		t.Fatalf("foo not removed:\n%s", out)
	}
	if !strings.Contains(out, "bar =") {
		t.Fatalf("bar should remain:\n%s", out)
	}
}

func TestNormalizeGoReplaceTarget(t *testing.T) {
	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{"https://github.com/foo/bar.git", "github.com/foo/bar", false},
		{"git@github.com:foo/bar.git", "github.com/foo/bar", false},
		{"github.com/foo/bar", "github.com/foo/bar", false},
		{"ftp://github.com/foo/bar", "", true},
		{"git+https://github.com/foo/bar", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := normalizeGoReplaceTarget(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("normalizeGoReplaceTarget(%q) = %q, want error", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizeGoReplaceTarget(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("normalizeGoReplaceTarget(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
