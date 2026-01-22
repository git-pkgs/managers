package managers

import (
	"reflect"
	"testing"

	"github.com/git-pkgs/managers/definitions"
)

func loadTranslator(t *testing.T) *Translator {
	t.Helper()
	defs, err := definitions.LoadEmbedded()
	if err != nil {
		t.Fatalf("failed to load definitions: %v", err)
	}

	translator := NewTranslator()
	for _, def := range defs {
		translator.Register(def)
	}
	return translator
}

// --- npm tests ---

func TestNpmInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNpmInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "ci"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNpmAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "add", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "install", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNpmAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "add", CommandInput{
		Args:  map[string]string{"package": "lodash"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "install", "lodash", "--save-dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNpmAddVersion(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "add", CommandInput{
		Args: map[string]string{
			"package": "lodash",
			"version": "4.17.21",
		},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "install", "lodash@4.17.21"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNpmList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "list", "--json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNpmRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "remove", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "uninstall", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNpmOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "outdated", "--json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNpmUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "update", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "update", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- bundler tests ---

func TestBundlerInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "install", "--frozen"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "add", CommandInput{
		Args: map[string]string{"package": "rails"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "add", "rails"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "add", CommandInput{
		Args:  map[string]string{"package": "rspec"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "add", "rspec", "--group", "development"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerAddVersion(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "add", CommandInput{
		Args: map[string]string{
			"package": "rails",
			"version": "~> 7.0",
		},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "add", "rails", "--version", "~> 7.0"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "remove", CommandInput{
		Args: map[string]string{"package": "rails"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "remove", "rails"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "list", "--format=json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "outdated", "--parseable"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "update", CommandInput{
		Args: map[string]string{"package": "rails"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "update", "rails"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- cargo tests ---

func TestCargoInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cargo", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "fetch"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCargoInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cargo", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "fetch", "--frozen"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCargoAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cargo", "add", CommandInput{
		Args: map[string]string{"package": "serde"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "add", "serde"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCargoAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cargo", "add", CommandInput{
		Args:  map[string]string{"package": "tokio-test"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "add", "tokio-test", "--dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCargoRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cargo", "remove", CommandInput{
		Args: map[string]string{"package": "serde"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "remove", "serde"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCargoList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cargo", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "tree"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCargoUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cargo", "update", CommandInput{
		Args: map[string]string{"package": "serde"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "update", "serde"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- gomod tests ---

func TestGomodInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gomod", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"go", "mod", "download"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGomodAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gomod", "add", CommandInput{
		Args: map[string]string{"package": "github.com/pkg/errors"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"go", "get", "github.com/pkg/errors"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGomodAddChain(t *testing.T) {
	tr := loadTranslator(t)
	cmds, err := tr.BuildCommands("gomod", "add", CommandInput{
		Args: map[string]string{"package": "github.com/pkg/errors"},
	})
	if err != nil {
		t.Fatalf("BuildCommands failed: %v", err)
	}
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	expected1 := []string{"go", "get", "github.com/pkg/errors"}
	expected2 := []string{"go", "mod", "tidy"}
	if !reflect.DeepEqual(cmds[0], expected1) {
		t.Errorf("cmd[0]: got %v, want %v", cmds[0], expected1)
	}
	if !reflect.DeepEqual(cmds[1], expected2) {
		t.Errorf("cmd[1]: got %v, want %v", cmds[1], expected2)
	}
}

func TestGomodRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gomod", "remove", CommandInput{
		Args: map[string]string{"package": "github.com/pkg/errors"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"go", "get", "github.com/pkg/errors@none"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGomodList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gomod", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"go", "list", "-m", "-json", "all"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGomodOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gomod", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"go", "list", "-m", "-u", "-json", "all"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGomodUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gomod", "update", CommandInput{
		Args: map[string]string{"package": "github.com/pkg/errors"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"go", "get", "-u", "github.com/pkg/errors"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- error cases ---

func TestMissingRequiredPackage(t *testing.T) {
	tr := loadTranslator(t)
	_, err := tr.BuildCommand("npm", "add", CommandInput{})
	if err == nil {
		t.Error("expected error for missing package, got nil")
	}
}

func TestUnknownManager(t *testing.T) {
	tr := NewTranslator()
	_, err := tr.BuildCommand("unknown", "install", CommandInput{})
	if err == nil {
		t.Error("expected error for unknown manager, got nil")
	}
}

func TestUnsupportedOperation(t *testing.T) {
	tr := loadTranslator(t)
	_, err := tr.BuildCommand("npm", "unknown_operation", CommandInput{})
	if err != ErrUnsupportedOperation {
		t.Errorf("expected ErrUnsupportedOperation, got %v", err)
	}
}

// --- pnpm tests ---

func TestPnpmInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPnpmInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "install", "--frozen-lockfile"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPnpmAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "add", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "add", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPnpmAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "add", CommandInput{
		Args:  map[string]string{"package": "lodash"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "add", "lodash", "--save-dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPnpmRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "remove", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "remove", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPnpmList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "list", "--json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPnpmOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "outdated", "--json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPnpmUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "update", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "update", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- uv tests ---

func TestUvInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "sync"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestUvInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "sync", "--frozen"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestUvAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "add", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "add", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestUvAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "add", CommandInput{
		Args:  map[string]string{"package": "pytest"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "add", "pytest", "--dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestUvRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "remove", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "remove", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestUvList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "tree"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestUvOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "tree", "--outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestUvUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "update", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "sync", "--upgrade"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- yarn tests ---

func TestYarnInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("yarn", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"yarn", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestYarnInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("yarn", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"yarn", "install", "--frozen-lockfile"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestYarnAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("yarn", "add", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"yarn", "add", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestYarnAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("yarn", "add", CommandInput{
		Args:  map[string]string{"package": "lodash"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"yarn", "add", "lodash", "--dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestYarnRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("yarn", "remove", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"yarn", "remove", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestYarnList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("yarn", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"yarn", "list", "--json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestYarnOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("yarn", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"yarn", "outdated", "--json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestYarnUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("yarn", "update", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	// yarn uses "upgrade" not "update"
	expected := []string{"yarn", "upgrade", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- composer tests ---

func TestComposerInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerInstallProduction(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "install", CommandInput{
		Flags: map[string]any{"production": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "install", "--no-dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "add", CommandInput{
		Args: map[string]string{"package": "monolog/monolog"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "require", "monolog/monolog"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "add", CommandInput{
		Args:  map[string]string{"package": "phpunit/phpunit"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "require", "phpunit/phpunit", "--dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "remove", CommandInput{
		Args: map[string]string{"package": "monolog/monolog"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "remove", "monolog/monolog"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "show"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerOutdatedJson(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "outdated", CommandInput{
		Flags: map[string]any{"json": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "outdated", "--format=json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "update", CommandInput{
		Args: map[string]string{"package": "monolog/monolog"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "update", "monolog/monolog"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- poetry tests ---

func TestPoetryInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPoetryInstallProduction(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "install", CommandInput{
		Flags: map[string]any{"production": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "install", "--only", "main"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPoetryAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "add", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "add", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPoetryAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "add", CommandInput{
		Args:  map[string]string{"package": "pytest"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "add", "pytest", "--group", "dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPoetryRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "remove", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "remove", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPoetryList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "show"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPoetryOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "show", "--outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPoetryUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "update", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "update", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- mix tests ---

func TestMixInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("mix", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mix", "deps.get"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMixInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("mix", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mix", "deps.get", "--check-locked"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMixList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("mix", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mix", "deps"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMixOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("mix", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mix", "hex.outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMixUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("mix", "update", CommandInput{
		Args: map[string]string{"package": "phoenix"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mix", "deps.update", "phoenix"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMixUpdateAll(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("mix", "update", CommandInput{
		Flags: map[string]any{"all": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mix", "deps.update", "--all"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- pub tests ---

func TestPubInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pub", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dart", "pub", "get"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPubInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pub", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dart", "pub", "get", "--enforce-lockfile"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPubAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pub", "add", CommandInput{
		Args: map[string]string{"package": "http"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dart", "pub", "add", "http"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPubRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pub", "remove", CommandInput{
		Args: map[string]string{"package": "http"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dart", "pub", "remove", "http"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPubList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pub", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dart", "pub", "deps"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPubOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pub", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dart", "pub", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPubOutdatedJson(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pub", "outdated", CommandInput{
		Flags: map[string]any{"json": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dart", "pub", "outdated", "--json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPubUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pub", "update", CommandInput{
		Args: map[string]string{"package": "http"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dart", "pub", "upgrade", "http"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- cocoapods tests ---

func TestCocoapodsInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cocoapods", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pod", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCocoapodsInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cocoapods", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pod", "install", "--deployment"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCocoapodsOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cocoapods", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pod", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCocoapodsUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cocoapods", "update", CommandInput{
		Args: map[string]string{"package": "Alamofire"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pod", "update", "Alamofire"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- bun tests ---

func TestBunInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "install", "--frozen-lockfile"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunInstallProduction(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "install", CommandInput{
		Flags: map[string]any{"production": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "install", "--production"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "add", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "add", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "add", CommandInput{
		Args:  map[string]string{"package": "typescript"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "add", "typescript", "--dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "remove", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "remove", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "pm", "list"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "update", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "update", "lodash"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunUpdateLatest(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "update", CommandInput{
		Flags: map[string]any{"latest": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "update", "--latest"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- extra args tests ---

func TestExtraArgs(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "install", CommandInput{
		Extra: []string{"--legacy-peer-deps"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "install", "--legacy-peer-deps"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestExtraArgsWithFlags(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "add", CommandInput{
		Args:  map[string]string{"package": "lodash"},
		Flags: map[string]any{"dev": true},
		Extra: []string{"--legacy-peer-deps", "--verbose"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "install", "lodash", "--save-dev", "--legacy-peer-deps", "--verbose"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestExtraArgsBundler(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "install", CommandInput{
		Extra: []string{"--jobs=4", "--retry=3"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "install", "--jobs=4", "--retry=3"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestExtraArgsCargo(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cargo", "add", CommandInput{
		Args:  map[string]string{"package": "serde"},
		Extra: []string{"--features", "derive"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "add", "serde", "--features", "derive"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestExtraArgsGomod(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gomod", "add", CommandInput{
		Args:  map[string]string{"package": "github.com/pkg/errors"},
		Extra: []string{"-v"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"go", "get", "github.com/pkg/errors", "-v"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- maven tests ---

func TestMavenInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("maven", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mvn", "dependency:resolve"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMavenList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("maven", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mvn", "dependency:list"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMavenOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("maven", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mvn", "versions:display-dependency-updates"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMavenUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("maven", "update", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mvn", "versions:use-latest-releases", "-DgenerateBackupPoms=false"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- gradle tests ---

func TestGradleInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gradle", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gradle", "dependencies", "--write-locks"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGradleList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gradle", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gradle", "dependencies"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- nuget tests ---

func TestNugetInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nuget", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dotnet", "restore"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNugetInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nuget", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dotnet", "restore", "--locked-mode"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNugetAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nuget", "add", CommandInput{
		Args: map[string]string{"package": "Newtonsoft.Json"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dotnet", "add", "package", "Newtonsoft.Json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNugetAddVersion(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nuget", "add", CommandInput{
		Args:  map[string]string{"package": "Newtonsoft.Json"},
		Flags: map[string]any{"version": "13.0.1"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dotnet", "add", "package", "Newtonsoft.Json", "--version", "13.0.1"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNugetRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nuget", "remove", CommandInput{
		Args: map[string]string{"package": "Newtonsoft.Json"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dotnet", "remove", "package", "Newtonsoft.Json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNugetList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nuget", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dotnet", "list", "package"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNugetOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nuget", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"dotnet", "list", "package", "--outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- swift tests ---

func TestSwiftInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("swift", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"swift", "package", "resolve"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestSwiftAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("swift", "add", CommandInput{
		Args: map[string]string{"package": "https://github.com/apple/swift-argument-parser"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"swift", "package", "add-dependency", "https://github.com/apple/swift-argument-parser"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestSwiftList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("swift", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"swift", "package", "show-dependencies"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestSwiftOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("swift", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"swift", "package", "update", "--dry-run"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestSwiftUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("swift", "update", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"swift", "package", "update"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- deno tests ---

func TestDenoInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("deno", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"deno", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestDenoInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("deno", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"deno", "install", "--frozen"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestDenoAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("deno", "add", CommandInput{
		Args: map[string]string{"package": "@std/path"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"deno", "add", "@std/path"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestDenoAddDev(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("deno", "add", CommandInput{
		Args:  map[string]string{"package": "@std/testing"},
		Flags: map[string]any{"dev": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"deno", "add", "@std/testing", "--dev"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestDenoRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("deno", "remove", CommandInput{
		Args: map[string]string{"package": "@std/path"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"deno", "remove", "@std/path"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestDenoOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("deno", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"deno", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestDenoUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("deno", "update", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"deno", "outdated", "--update"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- conda tests ---

func TestCondaInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conda", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conda", "install", "--yes"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCondaInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conda", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conda", "install", "--yes", "--freeze-installed"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCondaAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conda", "add", CommandInput{
		Args: map[string]string{"package": "numpy"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conda", "install", "--yes", "numpy"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCondaRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conda", "remove", CommandInput{
		Args: map[string]string{"package": "numpy"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conda", "remove", "--yes", "numpy"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCondaList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conda", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conda", "list", "--json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCondaOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conda", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conda", "search", "--outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCondaUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conda", "update", CommandInput{
		Args: map[string]string{"package": "numpy"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conda", "update", "--yes", "numpy"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- pip tests ---

func TestPipInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pip", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pip", "install", "-r", "requirements.txt"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPipAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pip", "add", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pip", "install", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPipRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pip", "remove", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pip", "uninstall", "--yes", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPipList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pip", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pip", "list", "--format=json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPipOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pip", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pip", "list", "--outdated", "--format=json"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPipUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pip", "update", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pip", "install", "--upgrade", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- gem tests ---

func TestGemInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gem", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gem", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGemAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gem", "add", CommandInput{
		Args: map[string]string{"package": "nokogiri"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gem", "install", "nokogiri"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGemAddVersion(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gem", "add", CommandInput{
		Args:  map[string]string{"package": "nokogiri"},
		Flags: map[string]any{"version": "1.15.0"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gem", "install", "nokogiri", "--version", "1.15.0"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGemRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gem", "remove", CommandInput{
		Args: map[string]string{"package": "nokogiri"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gem", "uninstall", "nokogiri"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGemList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gem", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gem", "list", "--local"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGemOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gem", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gem", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGemUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gem", "update", CommandInput{
		Args: map[string]string{"package": "nokogiri"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gem", "update", "nokogiri"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- brew tests ---

func TestBrewInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("brew", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"brew", "bundle", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBrewAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("brew", "add", CommandInput{
		Args: map[string]string{"package": "jq"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"brew", "install", "jq"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBrewAddCask(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("brew", "add", CommandInput{
		Args:  map[string]string{"package": "visual-studio-code"},
		Flags: map[string]any{"cask": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"brew", "install", "visual-studio-code", "--cask"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBrewRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("brew", "remove", CommandInput{
		Args: map[string]string{"package": "jq"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"brew", "uninstall", "jq"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBrewList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("brew", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"brew", "list"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBrewOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("brew", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"brew", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBrewUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("brew", "update", CommandInput{
		Args: map[string]string{"package": "jq"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"brew", "upgrade", "jq"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- helm tests ---

func TestHelmInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("helm", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"helm", "dependency", "build"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestHelmAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("helm", "add", CommandInput{
		Args: map[string]string{
			"package": "bitnami",
			"url":     "https://charts.bitnami.com/bitnami",
		},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"helm", "repo", "add", "bitnami", "https://charts.bitnami.com/bitnami"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestHelmRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("helm", "remove", CommandInput{
		Args: map[string]string{"package": "bitnami"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"helm", "repo", "remove", "bitnami"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestHelmList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("helm", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"helm", "dependency", "list"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestHelmUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("helm", "update", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"helm", "dependency", "update"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- sbt tests ---

func TestSbtInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("sbt", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"sbt", "update"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestSbtList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("sbt", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"sbt", "dependencyTree"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestSbtOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("sbt", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"sbt", "dependencyUpdates"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- rebar3 tests ---

func TestRebar3Install(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("rebar3", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"rebar3", "get-deps"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestRebar3List(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("rebar3", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"rebar3", "deps"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestRebar3Outdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("rebar3", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"rebar3", "hex", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestRebar3Update(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("rebar3", "update", CommandInput{
		Args: map[string]string{"package": "cowboy"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"rebar3", "upgrade", "cowboy"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- cabal tests ---

func TestCabalInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cabal", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cabal", "build", "--only-dependencies"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCabalList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cabal", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cabal", "list", "--installed"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCabalOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cabal", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cabal", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCabalUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cabal", "update", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cabal", "update"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- stack tests ---

func TestStackInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("stack", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"stack", "build", "--only-dependencies"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestStackList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("stack", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"stack", "ls", "dependencies"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestStackOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("stack", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"stack", "ls", "dependencies", "--outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestStackUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("stack", "update", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"stack", "update"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- opam tests ---

func TestOpamInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("opam", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"opam", "install", "--deps-only", "."}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestOpamAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("opam", "add", CommandInput{
		Args: map[string]string{"package": "lwt"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"opam", "install", "lwt"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestOpamRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("opam", "remove", CommandInput{
		Args: map[string]string{"package": "lwt"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"opam", "remove", "lwt"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestOpamList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("opam", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"opam", "list", "--installed"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestOpamOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("opam", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"opam", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestOpamUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("opam", "update", CommandInput{
		Args: map[string]string{"package": "lwt"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"opam", "upgrade", "lwt"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- vcpkg tests ---

func TestVcpkgInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("vcpkg", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"vcpkg", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestVcpkgAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("vcpkg", "add", CommandInput{
		Args: map[string]string{"package": "fmt"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"vcpkg", "add", "port", "fmt"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestVcpkgRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("vcpkg", "remove", CommandInput{
		Args: map[string]string{"package": "fmt"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"vcpkg", "remove", "fmt"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestVcpkgList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("vcpkg", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"vcpkg", "list"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestVcpkgOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("vcpkg", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"vcpkg", "update"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestVcpkgUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("vcpkg", "update", CommandInput{
		Args: map[string]string{"package": "fmt"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"vcpkg", "upgrade", "fmt"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- conan tests ---

func TestConanInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conan", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conan", "install", "."}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestConanAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conan", "add", CommandInput{
		Args: map[string]string{"package": "boost/1.82.0"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conan", "install", "--requires", "boost/1.82.0"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestConanRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conan", "remove", CommandInput{
		Args: map[string]string{"package": "boost/1.82.0"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conan", "remove", "boost/1.82.0"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestConanList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conan", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conan", "list", "*:*"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestConanUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conan", "update", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conan", "install", ".", "--update"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- luarocks tests ---

func TestLuarocksInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("luarocks", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"luarocks", "install", "--deps-only", "."}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestLuarocksAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("luarocks", "add", CommandInput{
		Args: map[string]string{"package": "luasocket"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"luarocks", "install", "luasocket"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestLuarocksRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("luarocks", "remove", CommandInput{
		Args: map[string]string{"package": "luasocket"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"luarocks", "remove", "luasocket"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestLuarocksList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("luarocks", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"luarocks", "list"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestLuarocksOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("luarocks", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"luarocks", "list", "--outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestLuarocksUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("luarocks", "update", CommandInput{
		Args: map[string]string{"package": "luasocket"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"luarocks", "install", "luasocket"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- shards tests ---

func TestShardsInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("shards", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"shards", "install"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestShardsInstallFrozen(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("shards", "install", CommandInput{
		Flags: map[string]any{"frozen": true},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"shards", "install", "--frozen"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestShardsList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("shards", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"shards", "list"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestShardsOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("shards", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"shards", "outdated"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestShardsUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("shards", "update", CommandInput{
		Args: map[string]string{"package": "kemal"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"shards", "update", "kemal"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- nimble tests ---

func TestNimbleInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nimble", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"nimble", "install", "--depsOnly"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNimbleAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nimble", "add", CommandInput{
		Args: map[string]string{"package": "jester"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"nimble", "install", "jester"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNimbleRemove(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nimble", "remove", CommandInput{
		Args: map[string]string{"package": "jester"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"nimble", "uninstall", "jester"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNimbleList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nimble", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"nimble", "list", "--installed"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNimbleUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nimble", "update", CommandInput{
		Args: map[string]string{"package": "jester"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"nimble", "install", "jester"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- lein tests ---

func TestLeinInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("lein", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"lein", "deps"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestLeinList(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("lein", "list", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"lein", "deps", ":tree"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestLeinOutdated(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("lein", "outdated", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"lein", "ancient"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- cpanm tests ---

func TestCpanmInstall(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cpanm", "install", CommandInput{})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cpanm", "--installdeps", "."}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCpanmAdd(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cpanm", "add", CommandInput{
		Args: map[string]string{"package": "Moose"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cpanm", "--install", "Moose"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCpanmUpdate(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("cpanm", "update", CommandInput{
		Args: map[string]string{"package": "Moose"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cpanm", "--install", "Moose"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

// --- path command tests ---

func TestNpmPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("npm", "path", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"npm", "ls", "lodash", "--parseable"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPnpmPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pnpm", "path", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pnpm", "list", "lodash", "--parseable", "--depth", "0"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBundlerPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bundler", "path", CommandInput{
		Args: map[string]string{"package": "rails"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bundle", "info", "rails", "--path"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPipPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("pip", "path", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"pip", "show", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestUvPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("uv", "path", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"uv", "pip", "show", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGomodPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gomod", "path", CommandInput{
		Args: map[string]string{"package": "github.com/stretchr/testify"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"go", "list", "-m", "-json", "github.com/stretchr/testify"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCargoPath(t *testing.T) {
	tr := loadTranslator(t)
	// Cargo path uses extraction_only for package, so it shouldn't appear in command
	cmd, err := tr.BuildCommand("cargo", "path", CommandInput{
		Args: map[string]string{"package": "serde"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"cargo", "metadata", "--format-version", "1"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestComposerPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("composer", "path", CommandInput{
		Args: map[string]string{"package": "monolog/monolog"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"composer", "show", "monolog/monolog", "--path"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBrewPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("brew", "path", CommandInput{
		Args: map[string]string{"package": "git"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"brew", "--prefix", "git"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestGemPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("gem", "path", CommandInput{
		Args: map[string]string{"package": "rails"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"gem", "which", "rails"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestDenoPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("deno", "path", CommandInput{
		Args: map[string]string{"package": "https://deno.land/std/http/server.ts"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"deno", "info", "https://deno.land/std/http/server.ts"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestNimblePath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("nimble", "path", CommandInput{
		Args: map[string]string{"package": "jester"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"nimble", "path", "jester"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestOpamPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("opam", "path", CommandInput{
		Args: map[string]string{"package": "lwt"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"opam", "var", "lwt:lib"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestLuarocksPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("luarocks", "path", CommandInput{
		Args: map[string]string{"package": "luasocket"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"luarocks", "show", "luasocket"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestConanPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conan", "path", CommandInput{
		Args: map[string]string{"package": "zlib/1.2.13"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conan", "cache", "path", "zlib/1.2.13"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestPoetryPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("poetry", "path", CommandInput{
		Args: map[string]string{"package": "requests"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"poetry", "run", "pip", "show", "requests"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestCondaPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("conda", "path", CommandInput{
		Args: map[string]string{"package": "numpy"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"conda", "run", "pip", "show", "numpy"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestYarnPath(t *testing.T) {
	tr := loadTranslator(t)
	// yarn path uses extraction_only, so package not in command
	cmd, err := tr.BuildCommand("yarn", "path", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"yarn", "list", "--depth=0"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestBunPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("bun", "path", CommandInput{
		Args: map[string]string{"package": "lodash"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"bun", "pm", "ls"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestMixPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("mix", "path", CommandInput{
		Args: map[string]string{"package": "phoenix"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"mix", "deps"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestShardsPath(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("shards", "path", CommandInput{
		Args: map[string]string{"package": "kemal"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"shards", "list"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}

func TestRebar3Path(t *testing.T) {
	tr := loadTranslator(t)
	cmd, err := tr.BuildCommand("rebar3", "path", CommandInput{
		Args: map[string]string{"package": "cowboy"},
	})
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	expected := []string{"rebar3", "deps"}
	if !reflect.DeepEqual(cmd, expected) {
		t.Errorf("got %v, want %v", cmd, expected)
	}
}
