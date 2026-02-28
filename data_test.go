package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/diagnostics"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/engine"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/jsondoc"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/plan"
)

func TestFixtures(t *testing.T) {
	entries, err := os.ReadDir("tests")
	if err != nil {
		t.Fatalf("read tests directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		caseDir := filepath.Join("tests", entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			runFixtureCase(t, caseDir)
		})
	}
}

func runFixtureCase(t *testing.T, caseDir string) {
	t.Helper()

	inputBytes := mustReadFile(t, filepath.Join(caseDir, "input.json"))
	root := mustParseJSON(t, inputBytes)

	expectedPlanBytes := mustReadFile(t, filepath.Join(caseDir, "plan.json"))
	expectedPlan := mustParsePlan(t, expectedPlanBytes)

	generatedPlan, err := plan.Generate(root)
	if err != nil {
		t.Fatalf("generate plan: %v", err)
	}

	generatedPlanBytes := mustMarshalPlan(t, generatedPlan)
	if normalizeFixtureText(string(generatedPlanBytes)) != normalizeFixtureText(string(expectedPlanBytes)) {
		t.Fatalf("generated plan mismatch\nexpected:\n%s\nactual:\n%s", string(expectedPlanBytes), string(generatedPlanBytes))
	}

	if err := engine.Validate(root, expectedPlan); err != nil {
		t.Fatalf("validate baseline plan: %v", err)
	}

	rendered, err := engine.Render(root, expectedPlan)
	if err != nil {
		t.Fatalf("render baseline plan: %v", err)
	}

	expectedOutput := mustReadFile(t, filepath.Join(caseDir, "output.md"))
	if normalizeFixtureText(rendered) != normalizeFixtureText(string(expectedOutput)) {
		t.Fatalf("baseline markdown mismatch\nexpected:\n%s\nactual:\n%s", string(expectedOutput), rendered)
	}

	runValidPlans(t, caseDir, root)
	runInvalidPlans(t, caseDir, root)
}

func runValidPlans(t *testing.T, caseDir string, root *jsondoc.Node) {
	t.Helper()

	validDir := filepath.Join(caseDir, "valid-plans")
	entries, err := os.ReadDir(validDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("read valid plans: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".json")
		t.Run("valid/"+name, func(t *testing.T) {
			planBytes := mustReadFile(t, filepath.Join(validDir, entry.Name()))
			parsedPlan := mustParsePlan(t, planBytes)

			if err := engine.Validate(root, parsedPlan); err != nil {
				t.Fatalf("validate valid plan: %v", err)
			}

			rendered, err := engine.Render(root, parsedPlan)
			if err != nil {
				t.Fatalf("render valid plan: %v", err)
			}

			expectedPath := filepath.Join(validDir, name+".md")
			expectedOutput := mustReadFile(t, expectedPath)
			if normalizeFixtureText(rendered) != normalizeFixtureText(string(expectedOutput)) {
				t.Fatalf("valid plan markdown mismatch\nexpected:\n%s\nactual:\n%s", string(expectedOutput), rendered)
			}
		})
	}
}

func runInvalidPlans(t *testing.T, caseDir string, root *jsondoc.Node) {
	t.Helper()

	invalidDir := filepath.Join(caseDir, "invalid-plans")
	entries, err := os.ReadDir(invalidDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("read invalid plans: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".json")
		t.Run("invalid/"+name, func(t *testing.T) {
			planBytes := mustReadFile(t, filepath.Join(invalidDir, entry.Name()))
			parsedPlan := mustParsePlan(t, planBytes)

			err := engine.Validate(root, parsedPlan)
			if err == nil {
				t.Fatalf("expected validation error")
			}

			expectedError := strings.TrimSpace(string(mustReadFile(t, filepath.Join(invalidDir, name+".error"))))
			actualError := strings.TrimSpace(diagnostics.Format(err))
			if expectedError != actualError {
				t.Fatalf("invalid plan error mismatch\nexpected:\n%s\nactual:\n%s", expectedError, actualError)
			}

			if _, renderErr := engine.Render(root, parsedPlan); renderErr == nil {
				t.Fatalf("expected render error")
			}
		})
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return data
}

func normalizeFixtureText(value string) string {
	return strings.TrimRight(value, "\n")
}

func mustParseJSON(t *testing.T, data []byte) *jsondoc.Node {
	t.Helper()

	root, err := jsondoc.Parse(data)
	if err != nil {
		t.Fatalf("parse input json: %v", err)
	}

	return root
}

func mustParsePlan(t *testing.T, data []byte) *plan.Plan {
	t.Helper()

	parsedPlan, err := plan.Parse(data)
	if err != nil {
		t.Fatalf("parse plan: %v", err)
	}

	return parsedPlan
}

func mustMarshalPlan(t *testing.T, parsedPlan *plan.Plan) []byte {
	t.Helper()

	data, err := plan.Marshal(*parsedPlan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}

	return data
}
