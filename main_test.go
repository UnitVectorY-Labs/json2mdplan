package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// diff returns a simple line-by-line diff
func diff(expected, actual string) string {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	var result strings.Builder
	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := 0; i < maxLines; i++ {
		var expLine, actLine string
		if i < len(expectedLines) {
			expLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actLine = actualLines[i]
		}

		if expLine != actLine {
			result.WriteString(fmt.Sprintf("Line %d:\n", i+1))
			result.WriteString("  expected: ")
			result.WriteString(fmt.Sprintf("%q", expLine))
			result.WriteString("\n")
			result.WriteString("  actual:   ")
			result.WriteString(fmt.Sprintf("%q", actLine))
			result.WriteString("\n")
		}
	}

	return result.String()
}

func mustUnmarshal(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
}

func TestKeyToLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_name", "User Name"},
		{"first_name", "First Name"},
		{"firstName", "First Name"},
		{"URL", "URL"},
		{"status", "Status"},
		{"", ""},
		{"camelCase", "Camel Case"},
		{"snake_case_key", "Snake Case Key"},
		{"API_KEY", "API_KEY"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := keyToLabel(tt.input)
			if got != tt.expected {
				t.Errorf("keyToLabel(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGenerateTemplate(t *testing.T) {
	input := `{
		"title": "My Project",
		"status": "active",
		"tags": ["go", "cli"],
		"metadata": {
			"version": "1.0"
		},
		"contributors": [
			{"name": "Alice", "role": "lead"},
			{"name": "Bob", "role": "dev"}
		],
		"phases": [
			{"name": "Phase 1", "tasks": [{"id": 1}]}
		]
	}`

	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	tmpl := generateTemplate(data)

	if tmpl.Version != "1" {
		t.Errorf("expected version '1', got %q", tmpl.Version)
	}
	if tmpl.Template == nil {
		t.Fatal("expected non-nil template")
	}
	if tmpl.Template.Render != "inline" {
		t.Errorf("expected root render 'inline', got %q", tmpl.Template.Render)
	}

	// tags should be bullet_list
	if tags, ok := tmpl.Template.Properties["tags"]; ok {
		if tags.Render != "bullet_list" {
			t.Errorf("expected tags render 'bullet_list', got %q", tags.Render)
		}
	} else {
		t.Error("expected 'tags' property in template")
	}

	// contributors (flat objects) should be table
	if contrib, ok := tmpl.Template.Properties["contributors"]; ok {
		if contrib.Render != "table" {
			t.Errorf("expected contributors render 'table', got %q", contrib.Render)
		}
	} else {
		t.Error("expected 'contributors' property in template")
	}

	// phases (complex objects) should be sections
	if phases, ok := tmpl.Template.Properties["phases"]; ok {
		if phases.Render != "sections" {
			t.Errorf("expected phases render 'sections', got %q", phases.Render)
		}
		if phases.Items != nil && phases.Items.TitleKey != "name" {
			t.Errorf("expected phases title_key 'name', got %q", phases.Items.TitleKey)
		}
	} else {
		t.Error("expected 'phases' property in template")
	}
}

func TestConvertSimpleObject(t *testing.T) {
	input := `{"title": "Hello", "status": "active"}`
	tmplJSON := `{
		"version": "1",
		"template": {
			"render": "inline",
			"order": ["title", "status"],
			"properties": {
				"title": {"render": "labeled_value", "label": "Title"},
				"status": {"render": "labeled_value", "label": "Status"}
			}
		}
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	var tmplFile TemplateFile
	mustUnmarshal(t, []byte(tmplJSON), &tmplFile)

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	expected := "- **Title**: Hello\n- **Status**: active\n"
	if result != expected {
		t.Errorf("output mismatch\nExpected:\n%q\nActual:\n%q\nDiff:\n%s", expected, result, diff(expected, result))
	}
}

func TestConvertTable(t *testing.T) {
	input := `[
		{"name": "Alice", "role": "lead"},
		{"name": "Bob", "role": "dev"}
	]`
	tmplJSON := `{
		"version": "1",
		"template": {
			"render": "table",
			"items": {
				"order": ["name", "role"],
				"properties": {
					"name": {"label": "Name"},
					"role": {"label": "Role"}
				}
			}
		}
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	var tmplFile TemplateFile
	mustUnmarshal(t, []byte(tmplJSON), &tmplFile)

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	expected := "| Name | Role |\n| --- | --- |\n| Alice | lead |\n| Bob | dev |\n"
	if result != expected {
		t.Errorf("output mismatch\nExpected:\n%q\nActual:\n%q\nDiff:\n%s", expected, result, diff(expected, result))
	}
}

func TestConvertSections(t *testing.T) {
	input := `{
		"phases": [
			{"name": "Phase 1", "status": "done"},
			{"name": "Phase 2", "status": "pending"}
		]
	}`
	tmplJSON := `{
		"version": "1",
		"template": {
			"render": "inline",
			"order": ["phases"],
			"properties": {
				"phases": {
					"render": "sections",
					"title": "Phases",
					"items": {
						"render": "inline",
						"title_key": "name",
						"order": ["name", "status"],
						"properties": {
							"name": {"render": "labeled_value", "label": "Name"},
							"status": {"render": "labeled_value", "label": "Status"}
						}
					}
				}
			}
		}
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	var tmplFile TemplateFile
	mustUnmarshal(t, []byte(tmplJSON), &tmplFile)

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	// Phases has a title so sub-items are at level+1
	if !strings.Contains(result, "# Phases") {
		t.Error("expected '# Phases' heading")
	}
	if !strings.Contains(result, "## Phase 1") {
		t.Error("expected '## Phase 1' heading")
	}
	if !strings.Contains(result, "## Phase 2") {
		t.Error("expected '## Phase 2' heading")
	}
	// name should be skipped (title_key)
	if strings.Contains(result, "**Name**") {
		t.Error("title_key 'name' should not appear as labeled value")
	}
	if !strings.Contains(result, "**Status**: done") {
		t.Error("expected status labeled value for Phase 1")
	}
}

func TestConvertBulletList(t *testing.T) {
	input := `{"tags": ["go", "cli", "tool"]}`
	tmplJSON := `{
		"version": "1",
		"template": {
			"render": "inline",
			"properties": {
				"tags": {
					"render": "bullet_list",
					"title": "Tags"
				}
			}
		}
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	var tmplFile TemplateFile
	mustUnmarshal(t, []byte(tmplJSON), &tmplFile)

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	if !strings.Contains(result, "## Tags") {
		t.Error("expected '## Tags' heading")
	}
	if !strings.Contains(result, "- go\n") {
		t.Error("expected '- go' bullet")
	}
}

func TestConvertHidden(t *testing.T) {
	input := `{"visible": "yes", "secret": "no"}`
	tmplJSON := `{
		"version": "1",
		"template": {
			"render": "inline",
			"properties": {
				"visible": {"render": "labeled_value", "label": "Visible"},
				"secret": {"render": "hidden"}
			}
		}
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	var tmplFile TemplateFile
	mustUnmarshal(t, []byte(tmplJSON), &tmplFile)

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	if !strings.Contains(result, "Visible") {
		t.Error("expected 'Visible' in output")
	}
	if strings.Contains(result, "secret") || strings.Contains(result, "no") {
		t.Error("hidden field should not appear in output")
	}
}

func TestConvertText(t *testing.T) {
	input := `{"description": "A great project"}`
	tmplJSON := `{
		"version": "1",
		"template": {
			"render": "inline",
			"properties": {
				"description": {"render": "text"}
			}
		}
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	var tmplFile TemplateFile
	mustUnmarshal(t, []byte(tmplJSON), &tmplFile)

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	if !strings.Contains(result, "A great project\n") {
		t.Errorf("expected text paragraph, got %q", result)
	}
}

func TestConvertHeading(t *testing.T) {
	input := `{"title": "My Document"}`
	tmplJSON := `{
		"version": "1",
		"template": {
			"render": "inline",
			"properties": {
				"title": {"render": "heading"}
			}
		}
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	var tmplFile TemplateFile
	mustUnmarshal(t, []byte(tmplJSON), &tmplFile)

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	if !strings.Contains(result, "## My Document\n") {
		t.Errorf("expected heading, got %q", result)
	}
}

func TestRunCLI(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{"no mode", []string{}, 2},
		{"conflicting modes", []string{"--generate", "--convert"}, 2},
		{"convert without template", []string{"--convert"}, 2},
		{"version", []string{"--version"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, strings.NewReader(""), &stdout, &stderr)
			if code != tt.wantCode {
				t.Errorf("expected exit code %d, got %d (stderr: %s)", tt.wantCode, code, stderr.String())
			}
		})
	}
}

func TestFormatScalar(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{float64(42), "42"},
		{float64(3.14), "3.14"},
		{true, "true"},
		{false, "false"},
		{nil, ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			got := formatScalar(tt.input)
			if got != tt.expected {
				t.Errorf("formatScalar(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNullValuesSkipped(t *testing.T) {
	input := `{"name": "Alice", "email": null}`
	tmplJSON := `{
		"version": "1",
		"template": {
			"render": "inline",
			"order": ["name", "email"],
			"properties": {
				"name": {"render": "labeled_value", "label": "Name"},
				"email": {"render": "labeled_value", "label": "Email"}
			}
		}
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	var tmplFile TemplateFile
	mustUnmarshal(t, []byte(tmplJSON), &tmplFile)

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	if strings.Contains(result, "Email") {
		t.Error("null values should be skipped")
	}
	if !strings.Contains(result, "**Name**: Alice") {
		t.Error("expected Name labeled value")
	}
}

func TestEscapeTableCell(t *testing.T) {
	if escapeTableCell("a|b") != "a\\|b" {
		t.Error("expected pipe to be escaped")
	}
	if escapeTableCell("normal") != "normal" {
		t.Error("expected no change for normal text")
	}
}

func TestGenerateRoundTrip(t *testing.T) {
	// Generate a template from data, then use it to convert the same data
	input := `{
		"title": "Test",
		"items": [
			{"name": "A", "value": 1},
			{"name": "B", "value": 2}
		]
	}`

	var data interface{}
	mustUnmarshal(t, []byte(input), &data)

	tmpl := generateTemplate(data)

	// Verify we can marshal/unmarshal the template
	tmplBytes, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatalf("failed to marshal template: %v", err)
	}

	var tmplFile TemplateFile
	if err := json.Unmarshal(tmplBytes, &tmplFile); err != nil {
		t.Fatalf("failed to unmarshal template: %v", err)
	}

	conv := newConverter(false)
	result := conv.convert(data, tmplFile.Template)

	if result == "" {
		t.Error("expected non-empty output from round trip")
	}
	if !strings.Contains(result, "Test") {
		t.Error("expected 'Test' in output")
	}
}
