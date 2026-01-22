package managers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/git-pkgs/managers/definitions"
)

func ExtractPath(output string, extract *definitions.Extract, pkg string) (string, error) {
	if extract == nil || extract.Type == "" || extract.Type == "raw" {
		return strings.TrimSpace(output), nil
	}

	var result string
	var err error

	switch extract.Type {
	case "json":
		result, err = extractJSON(output, extract.Field)
	case "line_prefix":
		result, err = extractLinePrefix(output, extract.Prefix)
	case "regex":
		result, err = extractRegex(output, extract.Pattern)
	case "json_array":
		result, err = extractJSONArray(output, extract.ArrayField, extract.MatchField, extract.ExtractField, pkg)
	case "template":
		result, err = extractTemplate(extract.Pattern, pkg)
	default:
		return "", fmt.Errorf("unknown extract type: %s", extract.Type)
	}

	if err != nil {
		return "", err
	}

	if extract.StripFilename {
		result = filepath.Dir(result)
	}

	return result, nil
}

func extractJSON(output string, field string) (string, error) {
	if field == "" {
		return "", fmt.Errorf("json extraction requires field name")
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	value, ok := data[field]
	if !ok {
		return "", fmt.Errorf("field %q not found in JSON", field)
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("field %q is not a string", field)
	}

	return str, nil
}

func extractLinePrefix(output string, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("line_prefix extraction requires prefix")
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), nil
		}
	}

	return "", fmt.Errorf("no line found with prefix %q", prefix)
}

func extractRegex(output string, pattern string) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("regex extraction requires pattern")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return "", fmt.Errorf("pattern did not match or no capture group found")
	}

	return strings.TrimSpace(matches[1]), nil
}

func extractTemplate(pattern, pkg string) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("template extraction requires pattern")
	}
	if pkg == "" {
		return "", fmt.Errorf("template extraction requires package name")
	}
	return strings.ReplaceAll(pattern, "{package}", pkg), nil
}

func extractJSONArray(output, arrayField, matchField, extractField, pkg string) (string, error) {
	if arrayField == "" || matchField == "" || extractField == "" {
		return "", fmt.Errorf("json_array extraction requires array_field, match_field, and extract_field")
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	arr, ok := data[arrayField].([]any)
	if !ok {
		return "", fmt.Errorf("field %q is not an array", arrayField)
	}

	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name, ok := obj[matchField].(string)
		if !ok || name != pkg {
			continue
		}

		value, ok := obj[extractField].(string)
		if !ok {
			return "", fmt.Errorf("field %q is not a string in matched element", extractField)
		}

		return strings.TrimSpace(value), nil
	}

	return "", fmt.Errorf("no element found with %s=%q", matchField, pkg)
}
