package managers

import (
	"testing"

	"github.com/git-pkgs/managers/definitions"
)

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
