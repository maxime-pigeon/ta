// Package review parses JSON output from linters and converts findings into
// Comments. It auto-detects the format used by ESLint, stylelint, and
// html-validate without requiring any configuration.
package review

import (
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
	capitalized := strings.ToUpper(severity[:1]) + severity[1:]
	return fmt.Sprintf("%s: %s", capitalized, message)
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
