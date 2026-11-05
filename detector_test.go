package managers

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/git-pkgs/managers/definitions"
)

func TestDetectorDetectFS(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"services/web/package.json":      {Data: []byte(`{"name":"web"}`)},
		"services/web/package-lock.json": {Data: []byte(`{"lockfileVersion":3}`)},
	}

	got, err := detector.DetectFS(project, "services/web", FileDetectOptions{})
	if err != nil {
		t.Fatalf("DetectFS: %v", err)
	}

	want := FileDetection{
		Manager:   "npm",
		Ecosystem: "npm",
		Directory: "services/web",
		Manifest:  "package.json",
		Lockfile:  "package-lock.json",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetectFS result = %#v, want %#v", got, want)
	}
}

func TestDetectorDetectFSUsesDefaultForManifestOnly(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"package.json": {Data: []byte(`{"name":"web"}`)},
	}

	got, err := detector.DetectFS(project, ".", FileDetectOptions{})
	if err != nil {
		t.Fatalf("DetectFS: %v", err)
	}
	if got.Manager != "npm" {
		t.Fatalf("manager = %q, want %q", got.Manager, "npm")
	}
}

func TestDetectorDetectFSMatchesManifestPatterns(t *testing.T) {
	detector := NewDetector(NewTranslator(), NewMockRunner())
	detector.Register(&definitions.Definition{
		Name:      "gem",
		Ecosystem: "rubygems",
		Detection: definitions.Detection{Manifests: []string{"*.gemspec"}},
	})

	got, err := detector.DetectFS(fstest.MapFS{
		"example.gemspec": {Data: []byte("Gem::Specification.new")},
	}, ".", FileDetectOptions{})
	if err != nil {
		t.Fatalf("DetectFS: %v", err)
	}
	if got.Manifest != "example.gemspec" {
		t.Fatalf("manifest = %q, want %q", got.Manifest, "example.gemspec")
	}
}

func TestDetectorDetectFSChecksFileContents(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"pyproject.toml": {Data: []byte("[tool.poetry]\nname = 'example'\n")},
	}

	got, err := detector.DetectFS(project, ".", FileDetectOptions{})
	if err != nil {
		t.Fatalf("DetectFS: %v", err)
	}
	if got.Manager != "poetry" {
		t.Fatalf("manager = %q, want %q", got.Manager, "poetry")
	}
}

func TestDetectorDetectFSReportsConflictingLockfiles(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"package.json":      {Data: []byte(`{"name":"example"}`)},
		"package-lock.json": {Data: []byte(`{"lockfileVersion":3}`)},
		"pnpm-lock.yaml":    {Data: []byte("lockfileVersion: '9.0'")},
	}

	_, err := detector.DetectFS(project, ".", FileDetectOptions{})
	var conflict ErrConflictingLockfiles
	if !errors.As(err, &conflict) {
		t.Fatalf("error = %v, want ErrConflictingLockfiles", err)
	}
	if want := []string{"pnpm-lock.yaml", "package-lock.json"}; !reflect.DeepEqual(conflict.Lockfiles, want) {
		t.Fatalf("lockfiles = %v, want %v", conflict.Lockfiles, want)
	}
}

func TestDetectorDetectFSUsesPriorityWhenConflictAllowed(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"package.json":      {Data: []byte(`{"name":"example"}`)},
		"package-lock.json": {Data: []byte(`{"lockfileVersion":3}`)},
		"pnpm-lock.yaml":    {Data: []byte("lockfileVersion: '9.0'")},
	}

	got, err := detector.DetectFS(project, ".", FileDetectOptions{OnConflict: ConflictUseFirst})
	if err != nil {
		t.Fatalf("DetectFS: %v", err)
	}
	if got.Manager != "pnpm" || got.Lockfile != "pnpm-lock.yaml" {
		t.Fatalf("DetectFS result = %#v, want pnpm lockfile", got)
	}
}

func TestDetectorDetectAllFSReturnsEachEcosystem(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"go.mod":            {Data: []byte("module example.com/project")},
		"go.sum":            {Data: []byte("example.com/dependency v1.0.0 h1:hash")},
		"package.json":      {Data: []byte(`{"name":"example"}`)},
		"package-lock.json": {Data: []byte(`{"lockfileVersion":3}`)},
	}

	got, err := detector.DetectAllFS(project, ".", FileDetectOptions{})
	if err != nil {
		t.Fatalf("DetectAllFS: %v", err)
	}

	want := []FileDetection{
		{
			Manager:   "gomod",
			Ecosystem: "golang",
			Directory: ".",
			Manifest:  "go.mod",
			Lockfile:  "go.sum",
		},
		{
			Manager:   "npm",
			Ecosystem: "npm",
			Directory: ".",
			Manifest:  "package.json",
			Lockfile:  "package-lock.json",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetectAllFS result = %#v, want %#v", got, want)
	}
}

func TestDetectorDetectAllFSUsesDefaultForManifestOnly(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"package.json": {Data: []byte(`{"name":"example"}`)},
	}

	got, err := detector.DetectAllFS(project, ".", FileDetectOptions{})
	if err != nil {
		t.Fatalf("DetectAllFS: %v", err)
	}
	if len(got) != 1 || got[0].Manager != "npm" {
		t.Fatalf("DetectAllFS result = %#v, want npm", got)
	}
}

func TestDetectorDetectUsesFilesystemDetection(t *testing.T) {
	detector := newEmbeddedDetector(t)
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/package.json", []byte(`{"name":"example"}`), 0o600); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(dir+"/package-lock.json", []byte(`{"lockfileVersion":3}`), 0o600); err != nil {
		t.Fatalf("write package-lock.json: %v", err)
	}

	manager, err := detector.Detect(dir, DetectOptions{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if manager.Name() != "npm" {
		t.Fatalf("manager = %q, want %q", manager.Name(), "npm")
	}
}

func TestDetectExplicit_RequireCLI(t *testing.T) {
	translator := NewTranslator()
	runner := NewMockRunner()
	detector := NewDetector(translator, runner)

	def := &definitions.Definition{
		Name:   "fakepkg",
		Binary: "binary-that-does-not-exist",
	}
	detector.Register(def)

	t.Run("require CLI fails when binary missing", func(t *testing.T) {
		_, err := detector.Detect("/tmp", DetectOptions{
			Manager:    "fakepkg",
			RequireCLI: true,
		})
		if err == nil {
			t.Fatal("expected error when binary not found")
		}
		var cliErr ErrCLINotFound
		if !isErrCLINotFound(err, &cliErr) {
			t.Fatalf("expected ErrCLINotFound, got: %v", err)
		}
	})

	t.Run("no CLI required succeeds without binary", func(t *testing.T) {
		mgr, err := detector.Detect("/tmp", DetectOptions{
			Manager:    "fakepkg",
			RequireCLI: false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mgr == nil {
			t.Fatal("expected manager, got nil")
		}
	})
}

func isErrCLINotFound(err error, target *ErrCLINotFound) bool {
	e, ok := err.(ErrCLINotFound)
	if ok {
		*target = e
	}
	return ok
}

func newEmbeddedDetector(t *testing.T) *Detector {
	t.Helper()

	defs, err := definitions.LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded: %v", err)
	}

	detector := NewDetector(NewTranslator(), NewMockRunner())
	for _, def := range defs {
		detector.Register(def)
	}
	return detector
}
