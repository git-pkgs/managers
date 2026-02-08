package managers

import (
	"context"
	"errors"
	"testing"

	"github.com/git-pkgs/managers/definitions"
)

func newTestManager(def *definitions.Definition, runner *MockRunner) *GenericManager {
	translator := NewTranslator()
	translator.Register(def)
	return &GenericManager{
		def:        def,
		dir:        "/test/project",
		translator: translator,
		runner:     runner,
	}
}

func TestGenericManager_Path_Raw(t *testing.T) {
	def := &definitions.Definition{
		Name:   "testpkg",
		Binary: "testpkg",
		Commands: map[string]definitions.Command{
			"path": {
				Base: []string{"show", "--path"},
				Args: map[string]definitions.Arg{
					"package": {Position: 0, Required: true},
				},
			},
		},
		Capabilities: []string{"path"},
	}

	runner := NewMockRunner()
	runner.Results = []*Result{{
		ExitCode: 0,
		Stdout:   "/usr/local/lib/testpkg/lodash\n",
	}}

	mgr := newTestManager(def, runner)
	result, err := mgr.Path(context.Background(), "lodash")
	if err != nil {
		t.Fatalf("Path failed: %v", err)
	}

	if result.Path != "/usr/local/lib/testpkg/lodash" {
		t.Errorf("got path %q, want %q", result.Path, "/usr/local/lib/testpkg/lodash")
	}

	if len(runner.Captured) != 1 {
		t.Fatalf("expected 1 command, got %d", len(runner.Captured))
	}
	expected := []string{"testpkg", "show", "--path", "lodash"}
	if !slicesEqual(runner.Captured[0], expected) {
		t.Errorf("got command %v, want %v", runner.Captured[0], expected)
	}
}

func TestGenericManager_Path_JSON(t *testing.T) {
	def := &definitions.Definition{
		Name:   "gomod",
		Binary: "go",
		Commands: map[string]definitions.Command{
			"path": {
				Base: []string{"list", "-m", "-json"},
				Args: map[string]definitions.Arg{
					"package": {Position: 0, Required: true},
				},
				Extract: &definitions.Extract{
					Type:  "json",
					Field: "Dir",
				},
			},
		},
		Capabilities: []string{"path"},
	}

	runner := NewMockRunner()
	runner.Results = []*Result{{
		ExitCode: 0,
		Stdout:   `{"Path": "github.com/pkg/errors", "Dir": "/home/user/go/pkg/mod/github.com/pkg/errors@v0.9.1"}`,
	}}

	mgr := newTestManager(def, runner)
	result, err := mgr.Path(context.Background(), "github.com/pkg/errors")
	if err != nil {
		t.Fatalf("Path failed: %v", err)
	}

	expected := "/home/user/go/pkg/mod/github.com/pkg/errors@v0.9.1"
	if result.Path != expected {
		t.Errorf("got path %q, want %q", result.Path, expected)
	}
}

func TestGenericManager_Path_LinePrefix(t *testing.T) {
	def := &definitions.Definition{
		Name:   "pip",
		Binary: "pip",
		Commands: map[string]definitions.Command{
			"path": {
				Base: []string{"show"},
				Args: map[string]definitions.Arg{
					"package": {Position: 0, Required: true},
				},
				Extract: &definitions.Extract{
					Type:   "line_prefix",
					Prefix: "Location: ",
				},
			},
		},
		Capabilities: []string{"path"},
	}

	runner := NewMockRunner()
	runner.Results = []*Result{{
		ExitCode: 0,
		Stdout: `Name: requests
Version: 2.28.1
Summary: Python HTTP for Humans.
Location: /usr/local/lib/python3.9/site-packages
Requires: certifi, charset-normalizer`,
	}}

	mgr := newTestManager(def, runner)
	result, err := mgr.Path(context.Background(), "requests")
	if err != nil {
		t.Fatalf("Path failed: %v", err)
	}

	expected := "/usr/local/lib/python3.9/site-packages"
	if result.Path != expected {
		t.Errorf("got path %q, want %q", result.Path, expected)
	}
}

func TestGenericManager_Path_Template(t *testing.T) {
	def := &definitions.Definition{
		Name:   "yarn",
		Binary: "yarn",
		Commands: map[string]definitions.Command{
			"path": {
				Base: []string{"why"},
				Args: map[string]definitions.Arg{
					"package": {Position: 0, Required: true, ExtractionOnly: true},
				},
				Extract: &definitions.Extract{
					Type:    "template",
					Pattern: "node_modules/{package}",
				},
			},
		},
		Capabilities: []string{"path"},
	}

	runner := NewMockRunner()
	runner.Results = []*Result{{
		ExitCode: 0,
		Stdout:   "whatever output, ignored for template",
	}}

	mgr := newTestManager(def, runner)
	result, err := mgr.Path(context.Background(), "lodash")
	if err != nil {
		t.Fatalf("Path failed: %v", err)
	}

	if result.Path != "node_modules/lodash" {
		t.Errorf("got path %q, want %q", result.Path, "node_modules/lodash")
	}

	// extraction_only means package arg is not passed to command
	expected := []string{"yarn", "why"}
	if !slicesEqual(runner.Captured[0], expected) {
		t.Errorf("got command %v, want %v", runner.Captured[0], expected)
	}
}

func TestGenericManager_Path_JSONArray(t *testing.T) {
	def := &definitions.Definition{
		Name:   "cargo",
		Binary: "cargo",
		Commands: map[string]definitions.Command{
			"path": {
				Base: []string{"metadata", "--format-version", "1"},
				Args: map[string]definitions.Arg{
					"package": {Position: 0, Required: true, ExtractionOnly: true},
				},
				Extract: &definitions.Extract{
					Type:          "json_array",
					ArrayField:    "packages",
					MatchField:    "name",
					ExtractField:  "manifest_path",
					StripFilename: true,
				},
			},
		},
		Capabilities: []string{"path"},
	}

	runner := NewMockRunner()
	runner.Results = []*Result{{
		ExitCode: 0,
		Stdout: `{
			"packages": [
				{"name": "serde", "manifest_path": "/home/user/.cargo/registry/src/serde-1.0.0/Cargo.toml"},
				{"name": "tokio", "manifest_path": "/home/user/.cargo/registry/src/tokio-1.0.0/Cargo.toml"}
			]
		}`,
	}}

	mgr := newTestManager(def, runner)
	result, err := mgr.Path(context.Background(), "serde")
	if err != nil {
		t.Fatalf("Path failed: %v", err)
	}

	expected := "/home/user/.cargo/registry/src/serde-1.0.0"
	if result.Path != expected {
		t.Errorf("got path %q, want %q", result.Path, expected)
	}
}

func TestGenericManager_Path_RunnerError(t *testing.T) {
	def := &definitions.Definition{
		Name:   "testpkg",
		Binary: "testpkg",
		Commands: map[string]definitions.Command{
			"path": {
				Base: []string{"path"},
				Args: map[string]definitions.Arg{
					"package": {Position: 0, Required: true},
				},
			},
		},
		Capabilities: []string{"path"},
	}

	runner := NewMockRunner()
	runner.Errors = []error{errors.New("command not found")}

	mgr := newTestManager(def, runner)
	_, err := mgr.Path(context.Background(), "lodash")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGenericManager_Path_ExtractionError(t *testing.T) {
	def := &definitions.Definition{
		Name:   "testpkg",
		Binary: "testpkg",
		Commands: map[string]definitions.Command{
			"path": {
				Base: []string{"path"},
				Args: map[string]definitions.Arg{
					"package": {Position: 0, Required: true},
				},
				Extract: &definitions.Extract{
					Type:   "line_prefix",
					Prefix: "Location: ",
				},
			},
		},
		Capabilities: []string{"path"},
	}

	runner := NewMockRunner()
	runner.Results = []*Result{{
		ExitCode: 0,
		Stdout:   "no location line here",
	}}

	mgr := newTestManager(def, runner)
	result, err := mgr.Path(context.Background(), "lodash")
	if err == nil {
		t.Error("expected extraction error, got nil")
	}
	// Result should still be returned even on extraction error
	if result == nil || result.Result == nil {
		t.Error("expected result to be returned even on extraction error")
	}
}

func TestGenericManager_Path_NoPathCommand(t *testing.T) {
	def := &definitions.Definition{
		Name:   "testpkg",
		Binary: "testpkg",
		Commands: map[string]definitions.Command{
			"install": {
				Base: []string{"install"},
			},
		},
		Capabilities: []string{"install"},
	}

	runner := NewMockRunner()
	mgr := newTestManager(def, runner)
	_, err := mgr.Path(context.Background(), "lodash")
	if err == nil {
		t.Error("expected error for missing path command, got nil")
	}
}

func TestGenericManager_Vendor(t *testing.T) {
	def := &definitions.Definition{
		Name:   "gomod",
		Binary: "go",
		Commands: map[string]definitions.Command{
			"vendor": {
				Base: []string{"mod", "vendor"},
			},
		},
		Capabilities: []string{"vendor"},
	}

	runner := NewMockRunner()
	runner.Results = []*Result{{
		ExitCode: 0,
		Stdout:   "",
	}}

	mgr := newTestManager(def, runner)
	result, err := mgr.Vendor(context.Background())
	if err != nil {
		t.Fatalf("Vendor failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("got exit code %d, want 0", result.ExitCode)
	}

	if len(runner.Captured) != 1 {
		t.Fatalf("expected 1 command, got %d", len(runner.Captured))
	}
	expected := []string{"go", "mod", "vendor"}
	if !slicesEqual(runner.Captured[0], expected) {
		t.Errorf("got command %v, want %v", runner.Captured[0], expected)
	}
}

func TestGenericManager_Vendor_NoCommand(t *testing.T) {
	def := &definitions.Definition{
		Name:   "testpkg",
		Binary: "testpkg",
		Commands: map[string]definitions.Command{
			"install": {
				Base: []string{"install"},
			},
		},
		Capabilities: []string{"install"},
	}

	runner := NewMockRunner()
	mgr := newTestManager(def, runner)
	_, err := mgr.Vendor(context.Background())
	if err == nil {
		t.Error("expected error for missing vendor command, got nil")
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
