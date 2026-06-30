package managers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	hexSHARE       = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)
	bareTOMLKeyRE  = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	gemSourceKeys  = map[string]bool{"path": true, "git": true, "github": true, "gist": true, "bitbucket": true, "branch": true, "tag": true, "ref": true}
	tomlSectionEnd = func(line string) bool { return tomlSectionName(line) != "" }
)

const manifestPerm os.FileMode = 0o644 // -rw-r--r--, matches what package managers write

func editTOMLReplaceEntry(path, section, pkg string, mode replaceMode, opts ReplaceOptions) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filepath.Base(path), err)
	}
	entry := tomlReplaceEntry(mode, opts)
	updated := updateTOMLSectionEntry(string(content), section, pkg, entry, mode == replaceModeDrop)
	return os.WriteFile(path, []byte(updated), manifestPerm)
}

func tomlReplaceEntry(mode replaceMode, opts ReplaceOptions) string {
	switch mode {
	case replaceModePath:
		return fmt.Sprintf("{ path = %s }", quoteTOMLString(opts.Path))
	case replaceModeGit:
		parts := []string{fmt.Sprintf("git = %s", quoteTOMLString(opts.Git))}
		if opts.Ref != "" {
			parts = append(parts, fmt.Sprintf("%s = %s", tomlGitRefKey(opts.Ref), quoteTOMLString(opts.Ref)))
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	default:
		return ""
	}
}

func tomlGitRefKey(ref string) string {
	if hexSHARE.MatchString(ref) {
		return "rev"
	}
	return "branch"
}

// updateTOMLSectionEntry sets or removes "key = entry" inside a single
// TOML table header. The section is created if it doesn't exist; if drop
// removes the last entry the section header is removed too. This is a
// line-based editor that preserves surrounding content; it does not
// support multi-line values for the target key.
func updateTOMLSectionEntry(content, section, key, entry string, drop bool) string {
	lines := strings.SplitAfter(content, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}

	target := tomlSectionName(section)
	start, end := -1, len(lines)
	for i, line := range lines {
		name := tomlSectionName(strings.TrimSpace(line))
		if name == target {
			start = i
			continue
		}
		if start >= 0 && i > start && tomlSectionEnd(strings.TrimSpace(line)) {
			end = i
			break
		}
	}

	keyName := tomlKey(key)

	if start == -1 {
		if drop {
			return content
		}
		prefix := content
		if prefix != "" && !strings.HasSuffix(prefix, "\n") {
			prefix += "\n"
		}
		if prefix != "" && !strings.HasSuffix(prefix, "\n\n") {
			prefix += "\n"
		}
		return prefix + section + "\n" + keyName + " = " + entry + "\n"
	}

	before := append([]string{}, lines[:start]...)
	header := lines[start : start+1]
	body := append([]string{}, lines[start+1:end]...)
	after := append([]string{}, lines[end:]...)

	body = filterTOMLKey(body, keyName)
	if !drop {
		body = append(body, keyName+" = "+entry+"\n")
	}
	if drop && !tomlSectionHasEntries(body) {
		return strings.Join(append(before, after...), "")
	}
	return strings.Join(append(append(append(before, header...), body...), after...), "")
}

func filterTOMLKey(lines []string, key string) []string {
	out := lines[:0]
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, key+"=") {
			continue
		}
		out = append(out, line)
	}
	return out
}

func tomlSectionName(line string) string {
	line = strings.TrimSpace(stripTOMLComment(line))
	if !strings.HasPrefix(line, "[") {
		return ""
	}
	end := strings.Index(line, "]")
	if end == -1 {
		return ""
	}
	return line[:end+1]
}

func stripTOMLComment(line string) string {
	if idx := strings.Index(line, "#"); idx >= 0 {
		return line[:idx]
	}
	return line
}

func tomlSectionHasEntries(lines []string) bool {
	for _, line := range lines {
		if strings.TrimSpace(stripTOMLComment(line)) != "" {
			return true
		}
	}
	return false
}

func tomlKey(key string) string {
	if bareTOMLKeyRE.MatchString(key) {
		return key
	}
	return quoteTOMLString(key)
}

func quoteTOMLString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func editGemfileReplaceEntry(path, pkg string, mode replaceMode, opts ReplaceOptions) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading Gemfile: %w", err)
	}

	lines := strings.SplitAfter(string(content), "\n")
	lineRE := regexp.MustCompile(`^\s*gem\s+["']` + regexp.QuoteMeta(pkg) + `["']`)
	found := false
	for i, line := range lines {
		if !lineRE.MatchString(strings.TrimRight(line, "\n")) {
			continue
		}
		updated, err := updateGemfileLine(line, pkg, mode, opts)
		if err != nil {
			return err
		}
		lines[i] = updated
		found = true
		break
	}
	if !found {
		if mode == replaceModeDrop {
			return fmt.Errorf("gem %q not found in Gemfile", pkg)
		}
		lines = append(lines, fmt.Sprintf("gem %q%s\n", pkg, gemReplacementSuffix(mode, opts)))
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "")), manifestPerm)
}

func updateGemfileLine(line, pkg string, mode replaceMode, opts ReplaceOptions) (string, error) {
	newline := ""
	if strings.HasSuffix(line, "\n") {
		newline = "\n"
		line = strings.TrimSuffix(line, "\n")
	}

	re := regexp.MustCompile(`^(\s*)gem\s+["']` + regexp.QuoteMeta(pkg) + `["'](.*)$`)
	m := re.FindStringSubmatch(line)
	if m == nil {
		return "", fmt.Errorf("gem %q not found in Gemfile line", pkg)
	}
	indent, tail := m[1], m[2]

	args, comment := parseGemfileArgs(tail)
	args = filterGemReplacementArgs(args)

	switch mode {
	case replaceModePath:
		args = append(args, fmt.Sprintf("path: %q", opts.Path))
	case replaceModeGit:
		args = append(args, fmt.Sprintf("git: %q", opts.Git))
		if opts.Ref != "" {
			args = append(args, fmt.Sprintf("branch: %q", opts.Ref))
		}
	case replaceModeDrop:
	}

	return fmt.Sprintf("%sgem %q%s%s", indent, pkg, formatGemfileArgs(args, comment), newline), nil
}

func gemReplacementSuffix(mode replaceMode, opts ReplaceOptions) string {
	switch mode {
	case replaceModePath:
		return fmt.Sprintf(", path: %q", opts.Path)
	case replaceModeGit:
		s := fmt.Sprintf(", git: %q", opts.Git)
		if opts.Ref != "" {
			s += fmt.Sprintf(", branch: %q", opts.Ref)
		}
		return s
	default:
		return ""
	}
}

func parseGemfileArgs(tail string) ([]string, string) {
	argsText, comment := splitRubyLineComment(tail)
	parts := splitRubyArgs(argsText)
	args := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			args = append(args, part)
		}
	}
	return args, comment
}

func splitRubyLineComment(s string) (string, string) {
	var quote rune
	escaped := false
	for i, r := range s {
		if escaped {
			escaped = false
			continue
		}
		if r == '\\' && quote != 0 {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		if r == '"' || r == '\'' {
			quote = r
			continue
		}
		if r == '#' {
			return strings.TrimRight(s[:i], " \t"), s[i:]
		}
	}
	return s, ""
}

func splitRubyArgs(s string) []string {
	var parts []string
	var quote rune
	escaped := false
	start := 0
	for i, r := range s {
		if escaped {
			escaped = false
			continue
		}
		if r == '\\' && quote != 0 {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		if r == '"' || r == '\'' {
			quote = r
			continue
		}
		if r == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func filterGemReplacementArgs(args []string) []string {
	out := args[:0]
	for _, arg := range args {
		if gemSourceKeys[gemArgKey(arg)] {
			continue
		}
		out = append(out, arg)
	}
	return out
}

func gemArgKey(arg string) string {
	key := strings.TrimSpace(arg)
	key = strings.TrimPrefix(key, ":")
	if before, _, ok := strings.Cut(key, "=>"); ok {
		return strings.TrimSpace(before)
	}
	if before, _, ok := strings.Cut(key, ":"); ok {
		return strings.TrimSpace(before)
	}
	return key
}

func formatGemfileArgs(args []string, comment string) string {
	var tail string
	if len(args) > 0 {
		tail = ", " + strings.Join(args, ", ")
	}
	if comment != "" {
		if tail != "" {
			tail += " "
		}
		tail += comment
	}
	return tail
}
