package managers

import (
	"reflect"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/git-pkgs/managers/definitions"
)

func TestAPMCommands(t *testing.T) {
	tr := loadTranslator(t)
	tests := []struct {
		name      string
		operation string
		input     CommandInput
		want      []string
	}{
		{"init", "init", CommandInput{}, []string{"apm", "init", "--yes"}},
		{"install", "install", CommandInput{}, []string{"apm", "install"}},
		{"frozen install", "install", CommandInput{Flags: map[string]any{"frozen": true}}, []string{"apm", "install", "--frozen"}},
		{"add", "add", CommandInput{Args: map[string]string{"package": "owner/repo"}}, []string{"apm", "install", "owner/repo"}},
		{"add with ref", "add", CommandInput{Args: map[string]string{"package": "owner/repo", "version": "v1.2.0"}}, []string{"apm", "install", "owner/repo#v1.2.0"}},
		{"add dev", "add", CommandInput{
			Args:  map[string]string{"package": "owner/repo", "version": "main"},
			Flags: map[string]any{"dev": true},
		}, []string{"apm", "install", "owner/repo#main", "--dev"}},
		{"remove", "remove", CommandInput{Args: map[string]string{"package": "owner/repo"}}, []string{"apm", "uninstall", "owner/repo"}},
		{"list", "list", CommandInput{}, []string{"apm", "deps", "list"}},
		{"outdated", "outdated", CommandInput{}, []string{"apm", "outdated"}},
		{"update all", "update", CommandInput{}, []string{"apm", "update", "--yes"}},
		{"update one", "update", CommandInput{Args: map[string]string{"package": "owner/repo"}}, []string{"apm", "update", "owner/repo", "--yes"}},
		{"resolve", "resolve", CommandInput{}, []string{"apm", "deps", "tree"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tr.BuildCommand("apm", tt.operation, tt.input)
			if err != nil {
				t.Fatalf("BuildCommand failed: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPMAddRequiresPackage(t *testing.T) {
	if _, err := loadTranslator(t).BuildCommand("apm", "add", CommandInput{}); err == nil {
		t.Error("expected an error without a package")
	}
}

func TestValidateAPMPackage(t *testing.T) {
	for _, name := range []string{"owner/repo", "gitlab.com/group/sub/repo", "git@github.com:owner/repo.git", "../local", "plugin@marketplace"} {
		if err := ValidatePackageName("apm_package", name); err != nil {
			t.Errorf("%q rejected: %v", name, err)
		}
	}
	for _, name := range []string{"", "-", "--force", "-g"} {
		if err := ValidatePackageName("apm_package", name); err == nil {
			t.Errorf("%q accepted", name)
		}
	}
}

func TestAPMRejectsFlagPackages(t *testing.T) {
	tr := loadTranslator(t)
	for _, op := range []string{"add", "remove", "update"} {
		if _, err := tr.BuildCommand("apm", op, CommandInput{Args: map[string]string{"package": "--force"}}); err == nil {
			t.Errorf("%s accepted a flag as the package", op)
		}
	}
	for _, pkg := range []string{"owner/repo", "https://github.com/owner/repo.git", "./skills/local", "plugin@marketplace"} {
		if _, err := tr.BuildCommand("apm", "add", CommandInput{Args: map[string]string{"package": pkg}}); err != nil {
			t.Errorf("add rejected %q: %v", pkg, err)
		}
	}
}

func TestAPMCapabilities(t *testing.T) {
	defs, err := definitions.LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	i := slices.IndexFunc(defs, func(d *definitions.Definition) bool { return d.Name == "apm" })
	if i < 0 {
		t.Fatal("apm definition not embedded")
	}
	m := &GenericManager{def: defs[i]}
	for _, c := range []Capability{CapInit, CapInstall, CapInstallFrozen, CapAdd, CapAddDev, CapRemove, CapUpdate, CapList, CapOutdated, CapResolve} {
		if !m.Supports(c) {
			t.Errorf("apm should support %s", c)
		}
	}
	// No APM dependency command has JSON output.
	for _, c := range []Capability{CapJSONOutput, CapPath, CapVendor} {
		if m.Supports(c) {
			t.Errorf("apm should not support %s", c)
		}
	}
}

func TestAPMDetectFS(t *testing.T) {
	detector := newEmbeddedDetector(t)
	tests := []struct {
		name    string
		project fstest.MapFS
		want    FileDetection
	}{
		{"manifest only", fstest.MapFS{
			"apm.yml": {Data: []byte("name: app\n")},
		}, FileDetection{Manager: "apm", Ecosystem: "apm", Directory: ".", Manifest: "apm.yml"}},
		{"current lockfile", fstest.MapFS{
			"apm.yml":       {Data: []byte("name: app\n")},
			"apm.lock.yaml": {Data: []byte("lockfile_version: \"2\"\n")},
		}, FileDetection{Manager: "apm", Ecosystem: "apm", Directory: ".", Manifest: "apm.yml", Lockfile: "apm.lock.yaml"}},
		{"legacy lockfile", fstest.MapFS{
			"apm.yml":  {Data: []byte("name: app\n")},
			"apm.lock": {Data: []byte("lockfile_version: \"1\"\n")},
		}, FileDetection{Manager: "apm", Ecosystem: "apm", Directory: ".", Manifest: "apm.yml", Lockfile: "apm.lock"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, policy := range []*FileDetectionPolicy{nil, GitPkgsFileDetectionPolicy()} {
				got, err := detector.DetectFS(tt.project, ".", FileDetectOptions{Policy: policy})
				if err != nil {
					t.Fatalf("DetectFS: %v", err)
				}
				if got != tt.want {
					t.Errorf("DetectFS (policy %v) = %#v, want %#v", policy != nil, got, tt.want)
				}
			}
		})
	}
}

func TestAPMDetectAllFSWithPython(t *testing.T) {
	detector := newEmbeddedDetector(t)
	project := fstest.MapFS{
		"pyproject.toml": {Data: []byte("[project]\nname = 'app'\n")},
		"apm.yml":        {Data: []byte("name: app\n")},
	}
	want := []FileDetection{
		{Manager: "uv", Ecosystem: "pypi", Directory: ".", Manifest: "pyproject.toml"},
		{Manager: "apm", Ecosystem: "apm", Directory: ".", Manifest: "apm.yml"},
	}
	for _, policy := range []*FileDetectionPolicy{nil, GitPkgsFileDetectionPolicy()} {
		got, err := detector.DetectAllFS(project, ".", FileDetectOptions{Policy: policy})
		if err != nil {
			t.Fatalf("DetectAllFS: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DetectAllFS (policy %v) = %#v, want %#v", policy != nil, got, want)
		}
	}
}
