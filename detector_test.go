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

func TestDetectorDetectFSSkipsDanglingFileCheckSymlink(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/package-lock.json", []byte(`{"lockfileVersion":3}`), 0o600); err != nil {
		t.Fatalf("write package-lock.json: %v", err)
	}
	if err := os.Symlink("missing", dir+"/pyproject.toml"); err != nil {
		t.Fatalf("symlink pyproject.toml: %v", err)
	}

	got, err := newEmbeddedDetector(t).DetectFS(os.DirFS(dir), ".", FileDetectOptions{})
	if err != nil {
		t.Fatalf("DetectFS: %v", err)
	}
	if got.Manager != "npm" {
		t.Fatalf("manager = %q, want %q", got.Manager, "npm")
	}
}

func TestDetectorDetectFSSkipsDanglingNestedManifestSymlink(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/package-lock.json", []byte(`{"lockfileVersion":3}`), 0o600); err != nil {
		t.Fatalf("write package-lock.json: %v", err)
	}
	if err := os.Mkdir(dir+"/project", 0o700); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.Symlink("missing", dir+"/project/build.properties"); err != nil {
		t.Fatalf("symlink project/build.properties: %v", err)
	}

	got, err := newEmbeddedDetector(t).DetectFS(os.DirFS(dir), ".", FileDetectOptions{})
	if err != nil {
		t.Fatalf("DetectFS: %v", err)
	}
	if got.Manager != "npm" {
		t.Fatalf("manager = %q, want %q", got.Manager, "npm")
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
			Manager:   "npm",
			Ecosystem: "npm",
			Directory: ".",
			Manifest:  "package.json",
			Lockfile:  "package-lock.json",
		},
		{
			Manager:   "gomod",
			Ecosystem: "go",
			Directory: ".",
			Manifest:  "go.mod",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetectAllFS result = %#v, want %#v", got, want)
	}
}

func TestDetectorDetectAllFSPreservesSelectionPolicy(t *testing.T) {
	detector := newEmbeddedDetector(t)
	tests := []struct {
		name    string
		project fstest.MapFS
		want    []FileDetection
	}{
		{
			name: "ecosystem order and names",
			project: fstest.MapFS{
				"package-lock.json": {Data: []byte(`{"lockfileVersion":3}`)},
				"Gemfile.lock":      {Data: []byte("GEM\n")},
				"Cargo.toml":        {Data: []byte("[package]\n")},
				"go.mod":            {Data: []byte("module example.com/app\n")},
				"go.sum":            {Data: []byte{}},
			},
			want: []FileDetection{
				{Manager: "npm", Ecosystem: "npm", Directory: ".", Lockfile: "package-lock.json"},
				{Manager: "bundler", Ecosystem: "rubygems", Directory: ".", Lockfile: "Gemfile.lock"},
				{Manager: "cargo", Ecosystem: "cargo", Directory: ".", Manifest: "Cargo.toml"},
				{Manager: "gomod", Ecosystem: "go", Directory: ".", Manifest: "go.mod"},
			},
		},
		{
			name: "manifest defaults",
			project: fstest.MapFS{
				"package.json":   {Data: []byte(`{"name":"app"}`)},
				"pyproject.toml": {Data: []byte("[project]\nname = 'app'\n")},
				"DESCRIPTION":    {Data: []byte("Package: app\n")},
			},
			want: []FileDetection{
				{Manager: "npm", Ecosystem: "npm", Directory: ".", Manifest: "package.json"},
				{Manager: "uv", Ecosystem: "pypi", Directory: ".", Manifest: "pyproject.toml"},
				{Manager: "renv", Ecosystem: "cran", Directory: ".", Manifest: "DESCRIPTION"},
			},
		},
		{
			name: "lockfiles bypass file checks and use priority",
			project: fstest.MapFS{
				"yarn.lock":            {Data: []byte("__metadata:\n")},
				".yarnrc.yml":          {Data: []byte("nodeLinker: node-modules\n")},
				"uv.lock":              {Data: []byte("version = 1\n")},
				"poetry.lock":          {Data: []byte("package = []\n")},
				"cabal.project.freeze": {Data: []byte("constraints: any.base\n")},
				"stack.yaml.lock":      {Data: []byte("packages: []\n")},
			},
			want: []FileDetection{
				{Manager: "yarn", Ecosystem: "npm", Directory: ".", Lockfile: "yarn.lock"},
				{Manager: "uv", Ecosystem: "pypi", Directory: ".", Lockfile: "uv.lock"},
				{Manager: "cabal", Ecosystem: "hackage", Directory: ".", Lockfile: "cabal.project.freeze"},
			},
		},
		{
			name: "detection-only definitions",
			project: fstest.MapFS{
				"Manifest.toml":     {Data: []byte("manifest_format = '2.0'\n")},
				"Cartfile.resolved": {Data: []byte("github example/repo 1.0.0\n")},
				"elm.json":          {Data: []byte(`{"type":"application"}`)},
				"dub.json":          {Data: []byte(`{"name":"app"}`)},
				"haxelib.json":      {Data: []byte(`{"name":"app"}`)},
				"flake.lock":        {Data: []byte(`{"version":7}`)},
			},
			want: []FileDetection{
				{Manager: "pkg", Ecosystem: "julia", Directory: ".", Lockfile: "Manifest.toml"},
				{Manager: "carthage", Ecosystem: "carthage", Directory: ".", Lockfile: "Cartfile.resolved"},
				{Manager: "elm", Ecosystem: "elm", Directory: ".", Manifest: "elm.json"},
				{Manager: "dub", Ecosystem: "dub", Directory: ".", Manifest: "dub.json"},
				{Manager: "haxelib", Ecosystem: "haxelib", Directory: ".", Manifest: "haxelib.json"},
				{Manager: "nix", Ecosystem: "nix", Directory: ".", Lockfile: "flake.lock"},
			},
		},
		{
			name: "nuget and manifest-only vcpkg",
			project: fstest.MapFS{
				"packages.config":     {Data: []byte("<packages />")},
				"project.assets.json": {Data: []byte(`{"version":3}`)},
				"vcpkg.json":          {Data: []byte(`{"name":"app"}`)},
			},
			want: []FileDetection{
				{Manager: "nuget", Ecosystem: "nuget", Directory: ".", Manifest: "packages.config", Lockfile: "project.assets.json"},
				{Manager: "vcpkg", Ecosystem: "vcpkg", Directory: ".", Manifest: "vcpkg.json"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := detector.DetectAllFS(test.project, ".", FileDetectOptions{
				OnConflict: ConflictUseFirst,
				Policy:     GitPkgsFileDetectionPolicy(),
			})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("DetectAllFS = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestGitPkgsFileDetectionPolicyKeepsExpansionsSeparate(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"bun.lockb":           {Data: []byte{}},
		"npm-shrinkwrap.json": {Data: []byte(`{"lockfileVersion":3}`)},
		"deno.jsonc":          {Data: []byte("{}")},
		"Chart.lock":          {Data: []byte("dependencies: []\n")},
		"Brewfile":            {Data: []byte("brew 'git'\n")},
	}

	got, err := detector.DetectAllFS(project, ".", FileDetectOptions{
		OnConflict: ConflictUseFirst,
		Policy:     GitPkgsFileDetectionPolicy(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("DetectAllFS = %#v, want no detections", got)
	}
}

func TestGitPkgsFileDetectionPolicyMatchesCabalManifest(t *testing.T) {
	detector := newEmbeddedDetector(t)
	got, err := detector.DetectAllFS(
		fstest.MapFS{"app.cabal": {Data: []byte("name: app\n")}},
		".",
		FileDetectOptions{Policy: GitPkgsFileDetectionPolicy()},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []FileDetection{{
		Manager:   "cabal",
		Ecosystem: "hackage",
		Directory: ".",
		Manifest:  "app.cabal",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetectAllFS = %#v, want %#v", got, want)
	}
}

func TestDetectorEcosystem(t *testing.T) {
	detector := newEmbeddedDetector(t)
	for _, test := range []struct {
		manager   string
		ecosystem string
	}{
		{manager: "bundler", ecosystem: "rubygems"},
		{manager: "gomod", ecosystem: "go"},
		{manager: "pkg", ecosystem: "julia"},
	} {
		got, ok := detector.Ecosystem(test.manager)
		if !ok || got != test.ecosystem {
			t.Fatalf("Ecosystem(%q) = %q, %t; want %q, true", test.manager, got, ok, test.ecosystem)
		}
	}
	if got, ok := detector.Ecosystem("missing"); ok || got != "" {
		t.Fatalf("Ecosystem(%q) = %q, %t; want empty result", "missing", got, ok)
	}
}

func TestDetectorDetectAllFSRejectsGoSumAlone(t *testing.T) {
	detector := newEmbeddedDetector(t)
	got, err := detector.DetectAllFS(fstest.MapFS{"go.sum": {Data: []byte{}}}, ".", FileDetectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("DetectAllFS = %#v, want no detections", got)
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

func TestDetectorAffectedDirectories(t *testing.T) {
	detector := newEmbeddedDetector(t)
	tests := []struct {
		path string
		want []string
	}{
		{path: "package.json", want: []string{"."}},
		{path: "apps/web/package.json", want: []string{"apps/web"}},
		{path: "example.gemspec", want: []string{"."}},
		{path: "service/project/build.properties", want: []string{"service"}},
		{path: "apps/web/.yarnrc.yml", want: []string{"apps/web"}},
		{path: "apps/web/main.go"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			got, err := detector.AffectedDirectories(test.path)
			if err != nil {
				t.Fatalf("AffectedDirectories: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("AffectedDirectories = %v, want %v", got, test.want)
			}
		})
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
