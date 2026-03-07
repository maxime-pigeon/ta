package review

import (
	"testing"
)

func TestFormatBody(t *testing.T) {
	tests := []struct {
		severity string
		message  string
		want     string
	}{
		{"error", "something went wrong", "Error: something went wrong"},
		{"warning", "be careful", "Warning: be careful"},
	}

	for _, tc := range tests {
		t.Run(tc.severity, func(t *testing.T) {
			got := formatBody(tc.severity, tc.message)
			if got != tc.want {
				t.Errorf(
					"formatBody(%q, %q) = %q, want %q",
					tc.severity, tc.message, got, tc.want,
				)
			}
		})
	}
}

func TestToComment(t *testing.T) {
	l := Lint{
		FilePath: "src/index.js",
		Line:     5,
		Column:   2,
		Rule:     "no-var",
		Message:  "Unexpected var.",
		Severity: "error",
	}
	got := l.ToComment()
	want := Comment{
		Path: "src/index.js",
		Line: 5,
		Body: "Error: Unexpected var.",
	}
	if got != want {
		t.Errorf("ToComment() = %+v, want %+v", got, want)
	}
}

func TestToComments(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		got := ToComments([]Lint{})
		if len(got) != 0 {
			t.Errorf("got %d comments, want 0", len(got))
		}
	})

	t.Run("preserves order", func(t *testing.T) {
		lints := []Lint{
			{
				FilePath: "a.js",
				Line:     1,
				Severity: "error",
				Message:  "first",
			},
			{
				FilePath: "b.js",
				Line:     2,
				Severity: "warning",
				Message:  "second",
			},
		}
		got := ToComments(lints)
		if len(got) != 2 {
			t.Fatalf("got %d comments, want 2", len(got))
		}
		if got[0].Path != "a.js" || got[0].Body != "Error: first" {
			t.Errorf("got[0] = %+v, unexpected", got[0])
		}
		if got[1].Path != "b.js" || got[1].Body != "Warning: second" {
			t.Errorf("got[1] = %+v, unexpected", got[1])
		}
	})
}
