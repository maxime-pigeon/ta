// Package review parses JSON output from linters and converts findings into
// Comments. It auto-detects the format used by ESLint, stylelint, and
// html-validate without requiring any configuration.
package review

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Lint is a single finding produced by a linter.
type Lint struct {
	FilePath string
	Line     int
	Column   int
	Rule     string
	Message  string
	Severity string
}

// ToComment converts a Lint into a Comment.
func (l Lint) ToComment() Comment {
	return Comment{
		Path: l.FilePath,
		Line: l.Line,
		Body: formatBody(l.Severity, l.Message),
	}
}

// formatBody formats a severity and message into a comment body string.
func formatBody(severity, message string) string {
	return fmt.Sprintf("%s: %s", strings.ToUpper(severity[:1])+severity[1:], message)
}

// Comment is a student-facing piece of feedback derived from a Lint.
type Comment struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Body string `json:"body"`
}

// ToComments converts a slice of Lints into student-facing Comments.
func ToComments(lints []Lint) []Comment {
	comments := make([]Comment, len(lints))
	for i, l := range lints {
		comments[i] = l.ToComment()
	}
	return comments
}

// Parse auto-detects the linter output format and returns all lints.
// It handles ESLint (filePath/messages/ruleId), stylelint (source/warnings/rule/text),
// and html-validate (filePath/messages/ruleId) formats.
func Parse(data []byte) ([]Lint, error) {
	var fileReports []map[string]any
	if err := json.Unmarshal(data, &fileReports); err != nil {
		return nil, fmt.Errorf("parsing linter output: %w", err)
	}

	var lints []Lint
	for _, report := range fileReports {
		filePath := findStr(report, "filePath", "source")
		rawLints := findArr(report, "messages", "warnings")
		for _, raw := range rawLints {
			m, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			line, ok := toInt(m["line"])
			if !ok || line == 0 {
				continue
			}
			column, _ := toInt(m["column"])
			lints = append(lints, Lint{
				FilePath: filePath,
				Line:     line,
				Column:   column,
				Rule:     findStr(m, "ruleId", "rule"),
				Message:  findStr(m, "message", "text"),
				Severity: parseSeverity(m["severity"]),
			})
		}
	}
	return lints, nil
}

// findStr returns the string value of the first matching key found in m.
func findStr(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			return v
		}
	}
	return ""
}

// findArr returns the []any value of the first matching key found in m.
func findArr(m map[string]any, keys ...string) []any {
	for _, k := range keys {
		if v, ok := m[k].([]any); ok {
			return v
		}
	}
	return nil
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

// parseSeverity normalizes a severity value to "warning" or "error" by
// auto-detecting whether it is an integer (2=error) or a string ("error").
func parseSeverity(v any) string {
	switch val := v.(type) {
	case float64:
		if int(val) == 2 {
			return "error"
		}
		return "warning"
	case string:
		if val == "error" {
			return "error"
		}
		return "warning"
	}
	return "warning"
}
