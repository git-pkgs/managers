package managers

import (
	"testing"

	"github.com/git-pkgs/managers/definitions"
)

func TestExtractPath_Raw(t *testing.T) {
	output := "  /path/to/package  \n"
	result, err := ExtractPath(output, nil, "")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	if result != "/path/to/package" {
		t.Errorf("got %q, want %q", result, "/path/to/package")
	}
}

func TestExtractPath_RawExplicit(t *testing.T) {
	output := "/path/to/package\n"
	result, err := ExtractPath(output, &definitions.Extract{Type: "raw"}, "")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	if result != "/path/to/package" {
		t.Errorf("got %q, want %q", result, "/path/to/package")
	}
}

func TestExtractPath_JSON(t *testing.T) {
	output := `{"Dir": "/home/user/go/pkg/mod/example.com@v1.0.0", "Path": "example.com"}`
	result, err := ExtractPath(output, &definitions.Extract{
		Type:  "json",
		Field: "Dir",
	}, "")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "/home/user/go/pkg/mod/example.com@v1.0.0"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_JSON_MissingField(t *testing.T) {
	output := `{"Path": "example.com"}`
	_, err := ExtractPath(output, &definitions.Extract{
		Type:  "json",
		Field: "Dir",
	}, "")
	if err == nil {
		t.Error("expected error for missing field, got nil")
	}
}

func TestExtractPath_LinePrefix(t *testing.T) {
	output := `Name: requests
Version: 2.28.1
Summary: Python HTTP for Humans.
Location: /usr/local/lib/python3.9/site-packages
Requires: certifi, charset-normalizer`
	result, err := ExtractPath(output, &definitions.Extract{
		Type:   "line_prefix",
		Prefix: "Location: ",
	}, "")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "/usr/local/lib/python3.9/site-packages"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_LinePrefix_NotFound(t *testing.T) {
	output := `Name: requests
Version: 2.28.1`
	_, err := ExtractPath(output, &definitions.Extract{
		Type:   "line_prefix",
		Prefix: "Location: ",
	}, "")
	if err == nil {
		t.Error("expected error for missing prefix, got nil")
	}
}

func TestExtractPath_Regex(t *testing.T) {
	output := `Package path: /var/lib/gems/3.0.0/gems/rails-7.0.0`
	result, err := ExtractPath(output, &definitions.Extract{
		Type:    "regex",
		Pattern: `Package path: (.+)`,
	}, "")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "/var/lib/gems/3.0.0/gems/rails-7.0.0"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_Regex_NoMatch(t *testing.T) {
	output := `some unrelated output`
	_, err := ExtractPath(output, &definitions.Extract{
		Type:    "regex",
		Pattern: `Package path: (.+)`,
	}, "")
	if err == nil {
		t.Error("expected error for no match, got nil")
	}
}

func TestExtractPath_JSONArray(t *testing.T) {
	output := `{
		"packages": [
			{"name": "serde", "manifest_path": "/home/user/.cargo/registry/src/serde-1.0.0/Cargo.toml"},
			{"name": "tokio", "manifest_path": "/home/user/.cargo/registry/src/tokio-1.0.0/Cargo.toml"}
		]
	}`
	result, err := ExtractPath(output, &definitions.Extract{
		Type:         "json_array",
		ArrayField:   "packages",
		MatchField:   "name",
		ExtractField: "manifest_path",
	}, "serde")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "/home/user/.cargo/registry/src/serde-1.0.0/Cargo.toml"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_JSONArray_NotFound(t *testing.T) {
	output := `{
		"packages": [
			{"name": "serde", "manifest_path": "/path/to/serde/Cargo.toml"}
		]
	}`
	_, err := ExtractPath(output, &definitions.Extract{
		Type:         "json_array",
		ArrayField:   "packages",
		MatchField:   "name",
		ExtractField: "manifest_path",
	}, "tokio")
	if err == nil {
		t.Error("expected error for package not found, got nil")
	}
}

func TestExtractPath_JSONArray_StripFilename(t *testing.T) {
	output := `{
		"packages": [
			{"name": "serde", "manifest_path": "/home/user/.cargo/registry/src/serde-1.0.0/Cargo.toml"}
		]
	}`
	result, err := ExtractPath(output, &definitions.Extract{
		Type:          "json_array",
		ArrayField:    "packages",
		MatchField:    "name",
		ExtractField:  "manifest_path",
		StripFilename: true,
	}, "serde")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "/home/user/.cargo/registry/src/serde-1.0.0"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_UnknownType(t *testing.T) {
	_, err := ExtractPath("output", &definitions.Extract{Type: "invalid"}, "")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestExtractPath_GemWhich(t *testing.T) {
	// Simulates gem which output
	output := `/var/lib/gems/3.0.0/gems/rails-7.0.0/lib/rails.rb`
	result, err := ExtractPath(output, &definitions.Extract{
		Type:    "regex",
		Pattern: `^(.+/gems/[^/]+)`,
	}, "")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "/var/lib/gems/3.0.0/gems/rails-7.0.0"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_LuarocksShow(t *testing.T) {
	// Simulates luarocks show output
	output := `luasocket 3.1.0-1 - Network support for the Lua language

Installed in:   /usr/local/lib/luarocks/rocks-5.4
...`
	result, err := ExtractPath(output, &definitions.Extract{
		Type:    "regex",
		Pattern: `Installed in:\s+(\S+)`,
	}, "")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "/usr/local/lib/luarocks/rocks-5.4"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_DenoInfo(t *testing.T) {
	// Simulates deno info output
	output := `local: /Users/user/.cache/deno/deps/https/deno.land/abc123
type: TypeScript
...`
	result, err := ExtractPath(output, &definitions.Extract{
		Type:   "line_prefix",
		Prefix: "local: ",
	}, "")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "/Users/user/.cache/deno/deps/https/deno.land/abc123"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_Template(t *testing.T) {
	// Template extraction computes path from pattern
	result, err := ExtractPath("ignored output", &definitions.Extract{
		Type:    "template",
		Pattern: "node_modules/{package}",
	}, "lodash")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "node_modules/lodash"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_Template_ScopedPackage(t *testing.T) {
	// Template with scoped npm package
	result, err := ExtractPath("", &definitions.Extract{
		Type:    "template",
		Pattern: "node_modules/{package}",
	}, "@types/node")
	if err != nil {
		t.Fatalf("ExtractPath failed: %v", err)
	}
	expected := "node_modules/@types/node"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractPath_Template_MissingPackage(t *testing.T) {
	_, err := ExtractPath("", &definitions.Extract{
		Type:    "template",
		Pattern: "deps/{package}",
	}, "")
	if err == nil {
		t.Error("expected error for missing package, got nil")
	}
}
