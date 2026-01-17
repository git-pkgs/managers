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
