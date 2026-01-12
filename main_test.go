package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// TestConvertDirectives runs data-driven tests for the directive-based convert functionality.
// Each subdirectory in testdata/convert-directives/ represents a test case with:
//   - instance.json: the JSON instance to convert
//   - schema.json: the JSON Schema
//   - plan.json: the directive-based plan file
//   - expected.md: the expected Markdown output
func TestConvertDirectives(t *testing.T) {
	testdataDir := "testdata/convert-directives"

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testName := entry.Name()
		testDir := filepath.Join(testdataDir, testName)

		t.Run(testName, func(t *testing.T) {
			// Read test files
			instanceBytes, err := os.ReadFile(filepath.Join(testDir, "instance.json"))
			if err != nil {
				t.Fatalf("failed to read instance.json: %v", err)
			}

			schemaBytes, err := os.ReadFile(filepath.Join(testDir, "schema.json"))
			if err != nil {
				t.Fatalf("failed to read schema.json: %v", err)
			}

			planBytes, err := os.ReadFile(filepath.Join(testDir, "plan.json"))
			if err != nil {
				t.Fatalf("failed to read plan.json: %v", err)
			}

			expectedBytes, err := os.ReadFile(filepath.Join(testDir, "expected.md"))
			if err != nil {
				t.Fatalf("failed to read expected.md: %v", err)
			}

			// Run conversion
			actual, err := runConversion(instanceBytes, schemaBytes, planBytes)
			if err != nil {
				t.Fatalf("conversion failed: %v", err)
			}

			expected := strings.TrimSpace(string(expectedBytes))
			actual = strings.TrimSpace(actual)

			if actual != expected {
				t.Errorf("output mismatch\n\nExpected:\n%s\n\nActual:\n%s\n\nDiff:\n%s",
					expected, actual, diff(expected, actual))
			}
		})
	}
}

// runConversion performs the JSON to Markdown conversion for testing
func runConversion(instanceBytes, schemaBytes, planBytes []byte) (string, error) {
	// Parse JSON instance
	var jsonInstance interface{}
	if err := jsonUnmarshal(instanceBytes, &jsonInstance); err != nil {
		return "", &inputError{"invalid JSON instance: " + err.Error()}
	}

	// Parse schema
	var schema map[string]interface{}
	if err := jsonUnmarshal(schemaBytes, &schema); err != nil {
		return "", &inputError{"invalid JSON schema: " + err.Error()}
	}

	// Parse plan
	var plan Plan
	if err := jsonUnmarshal(planBytes, &plan); err != nil {
		return "", &inputError{"invalid plan JSON: " + err.Error()}
	}

	// Create config
	config := &ConvertConfig{
		JSONInstance: jsonInstance,
		JSONBytes:    instanceBytes,
		Schema:       schema,
		SchemaBytes:  schemaBytes,
		Plan:         &plan,
		PlanBytes:    planBytes,
		Verbose:      false,
	}

	// Convert to Markdown
	markdown := convertToMarkdown(config)
	return markdown, nil
}

// jsonUnmarshal is a helper to unmarshal JSON
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

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

// TestPlanValidation runs data-driven tests for plan schema validation.
// Files in testdata/plan-validation/valid/ should pass validation.
// Files in testdata/plan-validation/invalid/ should fail validation.
func TestPlanValidation(t *testing.T) {
	// Test valid plans
	t.Run("valid", func(t *testing.T) {
		validDir := "testdata/plan-validation/valid"
		entries, err := os.ReadDir(validDir)
		if err != nil {
			t.Fatalf("failed to read valid plans directory: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}

			testName := strings.TrimSuffix(entry.Name(), ".json")
			planPath := filepath.Join(validDir, entry.Name())

			t.Run(testName, func(t *testing.T) {
				planBytes, err := os.ReadFile(planPath)
				if err != nil {
					t.Fatalf("failed to read plan file: %v", err)
				}

				err = validatePlanBytes(planBytes)
				if err != nil {
					t.Errorf("expected valid plan, got error: %v", err)
				}
			})
		}
	})

	// Test invalid plans
	t.Run("invalid", func(t *testing.T) {
		invalidDir := "testdata/plan-validation/invalid"
		entries, err := os.ReadDir(invalidDir)
		if err != nil {
			t.Fatalf("failed to read invalid plans directory: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}

			testName := strings.TrimSuffix(entry.Name(), ".json")
			planPath := filepath.Join(invalidDir, entry.Name())

			t.Run(testName, func(t *testing.T) {
				planBytes, err := os.ReadFile(planPath)
				if err != nil {
					t.Fatalf("failed to read plan file: %v", err)
				}

				err = validatePlanBytes(planBytes)
				if err == nil {
					t.Errorf("expected invalid plan to fail validation, but it passed")
				}
			})
		}
	})
}

// validatePlanBytes validates a plan JSON against the plan schema
func validatePlanBytes(planBytes []byte) error {
	var planObj interface{}
	if err := json.Unmarshal(planBytes, &planObj); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	if err := compiler.AddResource(planSchemaValidationURL, bytes.NewReader([]byte(planSchemaJSON))); err != nil {
		return fmt.Errorf("failed to load plan schema: %v", err)
	}

	compiledSchema, err := compiler.Compile(planSchemaValidationURL)
	if err != nil {
		return fmt.Errorf("failed to compile plan schema: %v", err)
	}

	if err := compiledSchema.Validate(planObj); err != nil {
		return fmt.Errorf("plan validation failed: %v", err)
	}

	return nil
}

// TestSchemaDigest tests the schema digest generation
func TestSchemaDigest(t *testing.T) {
	testCases := []struct {
		name          string
		schema        string
		expectedPaths int
		expectedRoot  string
	}{
		{
			name: "simple-object",
			schema: `{
				"type": "object",
				"title": "Test",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "integer"}
				}
			}`,
			expectedPaths: 3, // root + name + age
			expectedRoot:  "object",
		},
		{
			name: "nested-object",
			schema: `{
				"type": "object",
				"properties": {
					"person": {
						"type": "object",
						"properties": {
							"name": {"type": "string"}
						}
					}
				}
			}`,
			expectedPaths: 3, // root + person + person/name
			expectedRoot:  "object",
		},
		{
			name: "array-of-objects",
			schema: `{
				"type": "object",
				"properties": {
					"items": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"id": {"type": "string"}
							}
						}
					}
				}
			}`,
			expectedPaths: 4, // root + items + items[*] + items[*]/id
			expectedRoot:  "object",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var schema map[string]interface{}
			if err := json.Unmarshal([]byte(tc.schema), &schema); err != nil {
				t.Fatalf("failed to parse schema: %v", err)
			}

			digest := generateSchemaDigest(schema)

			if digest.RootType != tc.expectedRoot {
				t.Errorf("expected root type %q, got %q", tc.expectedRoot, digest.RootType)
			}

			if len(digest.PathIndex) != tc.expectedPaths {
				t.Errorf("expected %d paths, got %d", tc.expectedPaths, len(digest.PathIndex))
				for i, p := range digest.PathIndex {
					t.Logf("  path[%d]: %s (type: %s)", i, p.Path, p.Type)
				}
			}
		})
	}
}

// TestSchemaFingerprint tests the deterministic fingerprint generation
func TestSchemaFingerprint(t *testing.T) {
	// Same schema with different whitespace/ordering should produce same fingerprint
	schema1 := `{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"integer"}}}`
	schema2 := `{
		"properties": {
			"b": {"type": "integer"},
			"a": {"type": "string"}
		},
		"type": "object"
	}`

	fp1 := calculateFingerprint([]byte(schema1))
	fp2 := calculateFingerprint([]byte(schema2))

	if fp1 != fp2 {
		t.Errorf("expected same fingerprint for equivalent schemas\n  schema1: %s\n  schema2: %s", fp1, fp2)
	}

	// Different schema should produce different fingerprint
	schema3 := `{"type":"object","properties":{"c":{"type":"string"}}}`
	fp3 := calculateFingerprint([]byte(schema3))

	if fp1 == fp3 {
		t.Errorf("expected different fingerprint for different schemas")
	}
}
