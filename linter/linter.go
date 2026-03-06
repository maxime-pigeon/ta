// Package linter runs external Linters and converts their output into
// Comments. Each Linter is loaded from a TOML configuration file that
// specifies the CLI command, file extensions, and JSON field mappings.
package linter

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Lint is a single finding produced by a Linter.
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
		Body: fmt.Sprintf("%s: %s", strings.ToUpper(l.Severity[:1])+l.Severity[1:], l.Message),
	}
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

// Mapping describes how to extract fields from a Linter's JSON output.
type Mapping struct {
	FilePath     string `toml:"file_path"`
	Lints        string `toml:"lints"`
	Rule         string `toml:"rule"`
	Message      string `toml:"message"`
	Line         string `toml:"line"`
	Column       string `toml:"column"`
	Severity     string `toml:"severity"`
	SeverityType string `toml:"severity_type"` // "int" or "string"
}

// Linter describes a single linter.
type Linter struct {
	Name       string   `toml:"name"`
	Extensions []string `toml:"extensions"`
	Command    []string `toml:"command"`
	ConfigFile string   `toml:"config_file"`
	Stderr     bool     `toml:"stderr"`
	Mapping    Mapping  `toml:"mapping"`
}

// filterFiles returns only the files whose extension matches one in
// l.Extensions.
func (l Linter) filterFiles(files []string) []string {
	var result []string
	for _, f := range files {
		for _, ext := range l.Extensions {
			if strings.HasSuffix(f, ext) {
				result = append(result, f)
				break
			}
		}
	}
	return result
}

// parseLints converts the raw JSON output from a Linter into a flat slice of
// Lints using the field mappings in l.
func (l Linter) parseLints(data []byte) ([]Lint, error) {
	var fileReports []map[string]any
	if err := json.Unmarshal(data, &fileReports); err != nil {
		return nil, fmt.Errorf("parsing %s output: %w", l.Name, err)
	}

	var lints []Lint
	for _, fileReport := range fileReports {
		filePath, _ := fileReport[l.Mapping.FilePath].(string)
		rawLints, _ := fileReport[l.Mapping.Lints].([]any)
		found, err := l.lintsFrom(filePath, rawLints)
		if err != nil {
			return nil, err
		}
		lints = append(lints, found...)
	}
	return lints, nil
}

// lintsFrom converts the raw lints from a single file report into Lints.
func (l Linter) lintsFrom(filePath string, rawLints []any) ([]Lint, error) {
	var lints []Lint
	for _, raw := range rawLints {
		rawLint, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		rule, _ := rawLint[l.Mapping.Rule].(string)
		message, _ := rawLint[l.Mapping.Message].(string)
		line, ok := toInt(rawLint[l.Mapping.Line])
		if !ok || line == 0 {
			return nil, fmt.Errorf("%s: could not parse line number for finding %q in %s (raw value: %v)", l.Name, message, filePath, rawLint[l.Mapping.Line])
		}
		column, _ := toInt(rawLint[l.Mapping.Column])
		lints = append(lints, Lint{
			FilePath: filePath,
			Line:     line,
			Column:   column,
			Rule:     rule,
			Message:  message,
			Severity: parseSeverity(l.Mapping.SeverityType, rawLint[l.Mapping.Severity]),
		})
	}
	return lints, nil
}

// run invokes the Linter's CLI and parses its output. A non-zero exit code is
// treated as expected (it signals that issues were found); only failures to
// launch the process are returned as errors.
func (l Linter) run(files []string, dir string) ([]Lint, error) {
	args := l.Command[1:]
	if l.ConfigFile != "" {
		args = append(args, "--config", filepath.Join(dir, l.ConfigFile))
	}
	args = append(args, files...)
	cmd := exec.Command(l.Command[0], args...)
	cmd.Dir = dir
	var out []byte
	var err error
	if l.Stderr {
		out, err = cmd.CombinedOutput()
	} else {
		out, err = cmd.Output()
	}
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("running %s: %w", l.Name, err)
		}
	}
	return l.parseLints(out)
}

// RunAll runs each configured Linter against the matching subset of files and
// returns all Lints collected.
func RunAll(linters []Linter, files []string, dir string) ([]Lint, error) {
	var lints []Lint
	for _, l := range linters {
		matched := l.filterFiles(files)
		if len(matched) == 0 {
			continue
		}
		found, err := l.run(matched, dir)
		if err != nil {
			return nil, fmt.Errorf("error running %s: %w", l.Name, err)
		}
		lints = append(lints, found...)
	}
	return lints, nil
}

// toInt converts a JSON-unmarshaled number (float64) or a native int to int.
// JSON numbers always unmarshal as float64, so a direct type assertion to int
// would fail.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	}
	return 0, false
}

// parseSeverity normalizes a linter's severity value to "warning" or "error".
// Linters encode severity differently: as an int (1=warning, 2=error) or as a
// string ("warning"/"error"), controlled by severityType.
func parseSeverity(severityType string, v any) string {
	switch severityType {
	case "int":
		n, _ := toInt(v)
		if n == 2 {
			return "error"
		}
		return "warning"
	case "string":
		if s, _ := v.(string); s == "error" {
			return "error"
		}
		return "warning"
	}
	return "warning"
}
