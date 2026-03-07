package review

import (
	"testing"
)

func TestFindStr(t *testing.T) {
	m := map[string]any{
		"a": "alpha",
		"b": "beta",
		"c": 42,
	}

	t.Run("first key found", func(t *testing.T) {
		v, ok := findStr(m, "a", "b")
		if !ok || v != "alpha" {
			t.Errorf("got (%q, %v), want (\"alpha\", true)", v, ok)
		}
	})

	t.Run("second key found", func(t *testing.T) {
		v, ok := findStr(m, "missing", "b")
		if !ok || v != "beta" {
			t.Errorf("got (%q, %v), want (\"beta\", true)", v, ok)
		}
	})

	t.Run("no key found", func(t *testing.T) {
		v, ok := findStr(m, "missing", "also-missing")
		if ok || v != "" {
			t.Errorf("got (%q, %v), want (\"\", false)", v, ok)
		}
	})

	t.Run("key present but wrong type", func(t *testing.T) {
		v, ok := findStr(m, "c")
		if ok || v != "" {
			t.Errorf("got (%q, %v), want (\"\", false)", v, ok)
		}
	})
}

func TestFindArr(t *testing.T) {
	arr1 := []any{"x"}
	arr2 := []any{"y"}
	m := map[string]any{
		"a": arr1,
		"b": arr2,
		"c": "not-an-array",
	}

	t.Run("first key found", func(t *testing.T) {
		v, ok := findArr(m, "a", "b")
		if !ok || len(v) != 1 || v[0] != "x" {
			t.Errorf("got (%v, %v), want ([x], true)", v, ok)
		}
	})

	t.Run("second key found", func(t *testing.T) {
		v, ok := findArr(m, "missing", "b")
		if !ok || len(v) != 1 || v[0] != "y" {
			t.Errorf("got (%v, %v), want ([y], true)", v, ok)
		}
	})

	t.Run("no key found", func(t *testing.T) {
		v, ok := findArr(m, "missing", "also-missing")
		if ok || v != nil {
			t.Errorf("got (%v, %v), want (nil, false)", v, ok)
		}
	})

	t.Run("key present but wrong type", func(t *testing.T) {
		v, ok := findArr(m, "c")
		if ok || v != nil {
			t.Errorf("got (%v, %v), want (nil, false)", v, ok)
		}
	})
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		want   int
		wantOk bool
	}{
		{"float64", float64(5), 5, true},
		{"float64 zero", float64(0), 0, true},
		{"int", int(3), 3, true},
		{"string", "3", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toInt(tc.input)
			if got != tc.want || ok != tc.wantOk {
				t.Errorf(
					"toInt(%v) = (%d, %v), want (%d, %v)",
					tc.input, got, ok, tc.want, tc.wantOk,
				)
			}
		})
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{"float64 1 → warning", float64(1), "warning", false},
		{"float64 2 → error", float64(2), "error", false},
		{"float64 unknown", float64(99), "", true},
		{"string warning", "warning", "warning", false},
		{"string error", "error", "error", false},
		{"string unknown", "critical", "", true},
		{"nil", nil, "", true},
		{"int type", int(1), "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSeverity(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf(
					"parseSeverity(%v) error = %v, wantErr %v",
					tc.input, err, tc.wantErr,
				)
			}
			if got != tc.want {
				t.Errorf(
					"parseSeverity(%v) = %q, want %q",
					tc.input, got, tc.want,
				)
			}
		})
	}
}

func TestParseRawLint(t *testing.T) {
	t.Run("ESLint lint", func(t *testing.T) {
		raw := map[string]any{
			"message":  "no-unused-vars message",
			"ruleId":   "no-unused-vars",
			"line":     float64(10),
			"column":   float64(3),
			"severity": float64(2),
		}
		got, err := parseRawLint("foo.js", raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := Lint{
			FilePath: "foo.js",
			Line:     10,
			Column:   3,
			Rule:     "no-unused-vars",
			Message:  "no-unused-vars message",
			Severity: "error",
		}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("stylelint lint", func(t *testing.T) {
		raw := map[string]any{
			"text":     "color-no-invalid-hex message",
			"rule":     "color-no-invalid-hex",
			"line":     float64(5),
			"column":   float64(1),
			"severity": "warning",
		}
		got, err := parseRawLint("foo.css", raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := Lint{
			FilePath: "foo.css",
			Line:     5,
			Column:   1,
			Rule:     "color-no-invalid-hex",
			Message:  "color-no-invalid-hex message",
			Severity: "warning",
		}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("not a map", func(t *testing.T) {
		_, err := parseRawLint("foo.js", "not-a-map")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("missing message and text", func(t *testing.T) {
		raw := map[string]any{
			"ruleId":   "rule",
			"line":     float64(1),
			"severity": float64(2),
		}
		_, err := parseRawLint("foo.js", raw)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("missing ruleId and rule", func(t *testing.T) {
		raw := map[string]any{
			"message":  "msg",
			"line":     float64(1),
			"severity": float64(2),
		}
		_, err := parseRawLint("foo.js", raw)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("missing line", func(t *testing.T) {
		raw := map[string]any{
			"message":  "msg",
			"ruleId":   "rule",
			"severity": float64(2),
		}
		_, err := parseRawLint("foo.js", raw)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("line zero", func(t *testing.T) {
		raw := map[string]any{
			"message":  "msg",
			"ruleId":   "rule",
			"line":     float64(0),
			"severity": float64(2),
		}
		_, err := parseRawLint("foo.js", raw)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("unknown severity", func(t *testing.T) {
		raw := map[string]any{
			"message":  "msg",
			"ruleId":   "rule",
			"line":     float64(1),
			"severity": float64(99),
		}
		_, err := parseRawLint("foo.js", raw)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("column optional", func(t *testing.T) {
		raw := map[string]any{
			"message":  "msg",
			"ruleId":   "rule",
			"line":     float64(1),
			"severity": float64(1),
		}
		got, err := parseRawLint("foo.js", raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Column != 0 {
			t.Errorf("got column %d, want 0", got.Column)
		}
	})
}

func TestParseFileReport(t *testing.T) {
	t.Run("ESLint file report", func(t *testing.T) {
		report := map[string]any{
			"filePath": "foo.js",
			"messages": []any{
				map[string]any{
					"message":  "msg",
					"ruleId":   "rule",
					"line":     float64(1),
					"severity": float64(2),
				},
			},
		}
		lints, err := parseFileReport(report)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lints) != 1 {
			t.Fatalf("got %d lints, want 1", len(lints))
		}
		if lints[0].FilePath != "foo.js" {
			t.Errorf(
				"got filePath %q, want \"foo.js\"",
				lints[0].FilePath,
			)
		}
	})

	t.Run("stylelint file report", func(t *testing.T) {
		report := map[string]any{
			"source": "foo.css",
			"warnings": []any{
				map[string]any{
					"text":     "msg",
					"rule":     "rule",
					"line":     float64(2),
					"severity": "error",
				},
			},
		}
		lints, err := parseFileReport(report)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lints) != 1 {
			t.Fatalf("got %d lints, want 1", len(lints))
		}
		if lints[0].FilePath != "foo.css" {
			t.Errorf(
				"got filePath %q, want \"foo.css\"",
				lints[0].FilePath,
			)
		}
	})

	t.Run("missing filePath and source", func(t *testing.T) {
		report := map[string]any{
			"messages": []any{},
		}
		_, err := parseFileReport(report)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("missing messages and warnings", func(t *testing.T) {
		report := map[string]any{
			"filePath": "foo.js",
		}
		_, err := parseFileReport(report)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("empty messages", func(t *testing.T) {
		report := map[string]any{
			"filePath": "foo.js",
			"messages": []any{},
		}
		lints, err := parseFileReport(report)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lints) != 0 {
			t.Errorf("got %d lints, want 0", len(lints))
		}
	})
}

var eslintJSON = []byte(`[
  {
    "filePath": "src/index.js",
    "messages": [
      {
        "message": "Unexpected var.",
        "ruleId": "no-var",
        "line": 3,
        "column": 1,
        "severity": 2
      }
    ]
  }
]`)

var stylelintJSON = []byte(`[
  {
    "source": "src/main.css",
    "warnings": [
      {
        "text": "Unexpected invalid hex color",
        "rule": "color-no-invalid-hex",
        "line": 7,
        "column": 3,
        "severity": "warning"
      }
    ]
  }
]`)

func TestParse(t *testing.T) {
	t.Run("ESLint format", func(t *testing.T) {
		lints, err := Parse(eslintJSON)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lints) != 1 {
			t.Fatalf("got %d lints, want 1", len(lints))
		}
		want := Lint{
			FilePath: "src/index.js",
			Line:     3,
			Column:   1,
			Rule:     "no-var",
			Message:  "Unexpected var.",
			Severity: "error",
		}
		if lints[0] != want {
			t.Errorf("got %+v, want %+v", lints[0], want)
		}
	})

	t.Run("stylelint format", func(t *testing.T) {
		lints, err := Parse(stylelintJSON)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lints) != 1 {
			t.Fatalf("got %d lints, want 1", len(lints))
		}
		want := Lint{
			FilePath: "src/main.css",
			Line:     7,
			Column:   3,
			Rule:     "color-no-invalid-hex",
			Message:  "Unexpected invalid hex color",
			Severity: "warning",
		}
		if lints[0] != want {
			t.Errorf("got %+v, want %+v", lints[0], want)
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		data := []byte(`[
      {"filePath":"a.js","messages":[
        {"message":"m1","ruleId":"r1","line":1,"severity":1}
      ]},
      {"filePath":"b.js","messages":[
        {"message":"m2","ruleId":"r2","line":2,"severity":2}
      ]}
    ]`)
		lints, err := Parse(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lints) != 2 {
			t.Fatalf("got %d lints, want 2", len(lints))
		}
	})

	t.Run("empty array", func(t *testing.T) {
		lints, err := Parse([]byte(`[]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lints) != 0 {
			t.Errorf("got %d lints, want 0", len(lints))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := Parse([]byte(`not json`))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("missing filePath", func(t *testing.T) {
		_, err := Parse([]byte(
			`[{"messages":[]}]`,
		))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
