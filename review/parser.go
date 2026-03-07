package review

import (
	"encoding/json"
	"fmt"
)

// Parse auto-detects the linter output format and returns all lints.
// It handles ESLint (filePath/messages/ruleId), stylelint
// (source/warnings/rule/text), and html-validate formats.
func Parse(data []byte) ([]Lint, error) {
	var fileReports []map[string]any
	if err := json.Unmarshal(data, &fileReports); err != nil {
		return nil, fmt.Errorf("parsing linter output: %w", err)
	}

	var lints []Lint
	for _, fileReport := range fileReports {
		fileLints, err := parseFileReport(fileReport)
		if err != nil {
			return nil, err
		}
		lints = append(lints, fileLints...)
	}
	return lints, nil
}

// parseFileReport extracts all lints from a single file report map.
func parseFileReport(fileReport map[string]any) ([]Lint, error) {
	filePath, ok := findStr(fileReport, "filePath", "source")
	if !ok {
		return nil, fmt.Errorf(
			"parsing file report: missing filePath or source",
		)
	}
	rawLints, ok := findArr(fileReport, "messages", "warnings")
	if !ok {
		return nil, fmt.Errorf(
			"parsing file report %q: missing messages or warnings",
			filePath,
		)
	}
	var lints []Lint
	for _, raw := range rawLints {
		lint, err := parseRawLint(filePath, raw)
		if err != nil {
			return nil, err
		}
		lints = append(lints, lint)
	}
	return lints, nil
}

// parseRawLint extracts a single Lint from a raw lint map belonging to
// the given filePath.
func parseRawLint(filePath string, raw any) (Lint, error) {
	m, ok := raw.(map[string]any)
	if !ok {
		return Lint{}, fmt.Errorf(
			"parsing lint in %q: unexpected type %T",
			filePath, raw,
		)
	}
	message, ok := findStr(m, "message", "text")
	if !ok {
		return Lint{}, fmt.Errorf(
			"parsing lint in %q: missing message or text",
			filePath,
		)
	}
	rule, ok := findStr(m, "ruleId", "rule")
	if !ok {
		return Lint{}, fmt.Errorf(
			"parsing lint %q in %q: missing ruleId or rule",
			message, filePath,
		)
	}
	line, ok := toInt(m["line"])
	if !ok || line == 0 {
		return Lint{}, fmt.Errorf(
			"parsing lint %q in %q: missing or invalid line number",
			message, filePath,
		)
	}
	severity, err := parseSeverity(m["severity"])
	if err != nil {
		return Lint{}, fmt.Errorf(
			"parsing lint %q in %q: %w",
			message, filePath, err,
		)
	}
	column, _ := toInt(m["column"])
	return Lint{
		FilePath: filePath,
		Line:     line,
		Column:   column,
		Rule:     rule,
		Message:  message,
		Severity: severity,
	}, nil
}

// findStr returns the string value of the first matching key found in m,
// and a bool indicating whether a match was found.
func findStr(m map[string]any, keys ...string) (string, bool) {
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			return v, true
		}
	}
	return "", false
}

// findArr returns the []any value of the first matching key found in m,
// and a bool indicating whether a match was found.
func findArr(m map[string]any, keys ...string) ([]any, bool) {
	for _, k := range keys {
		if v, ok := m[k].([]any); ok {
			return v, true
		}
	}
	return nil, false
}

// toInt converts a JSON-unmarshaled number (float64) or a native int to int.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	}
	return 0, false
}

// parseSeverity normalizes a severity value to "warning" or "error".
// It returns an error for unrecognized values.
func parseSeverity(v any) (string, error) {
	switch val := v.(type) {
	case float64:
		switch int(val) {
		case 1:
			return "warning", nil
		case 2:
			return "error", nil
		default:
			return "", fmt.Errorf("unknown severity %v", val)
		}
	case string:
		switch val {
		case "warning":
			return "warning", nil
		case "error":
			return "error", nil
		default:
			return "", fmt.Errorf("unknown severity %q", val)
		}
	}
	return "", fmt.Errorf("unknown severity type %T", v)
}
