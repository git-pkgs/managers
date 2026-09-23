package managers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestComposerResolve(t *testing.T) {
	runner := NewMockRunner()
	detector := NewDetector(NewTranslator(), runner)
	detector.Register(embeddedDef(t, "composer"))
	mgr, err := detector.Detect(t.TempDir(), DetectOptions{Manager: "composer"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{"composer", "show", "--tree", "--format=json"}
	if len(runner.Captured) != 1 || !slicesEqual(runner.Captured[0], want) {
		t.Errorf("commands = %v, want [%v]", runner.Captured, want)
	}
}

func TestComposerDetectVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires a POSIX shell")
	}
	def := embeddedDef(t, "composer")
	def.Binary = filepath.Join(t.TempDir(), "composer")
	script := "#!/bin/sh\n[ \"$#\" = 1 ] && [ \"$1\" = --version ] || exit 1\nprintf '%s\\n' \"$COMPOSER_TEST_OUTPUT\"\n"
	if err := os.WriteFile(def.Binary, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	detector := NewDetector(NewTranslator(), NewMockRunner())
	tests := []struct {
		output string
		want   string
	}{
		{"Composer version 2.10.3 2026-08-27 13:34:23", "2.10.3"},
		{"Composer version 2.10.0-RC2 2026-05-20 11:00:00", "2.10.0-RC2"},
		{"Composer version 2.11.999-dev+source", "2.11.999-dev+source"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Setenv("COMPOSER_TEST_OUTPUT", tt.output)
			version, err := detector.DetectVersion(def)
			if err != nil {
				t.Fatal(err)
			}
			if version != tt.want {
				t.Errorf("version = %q, want %q", version, tt.want)
			}
		})
	}
}
