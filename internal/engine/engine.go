package engine

import (
	"strings"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/diagnostics"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/directives"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/jsondoc"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/plan"
)

type Evaluation struct {
	Lines []string
}

func Validate(root *jsondoc.Node, parsedPlan *plan.Plan) error {
	_, err := evaluate(root, parsedPlan)
	return err
}

func Render(root *jsondoc.Node, parsedPlan *plan.Plan) (string, error) {
	evaluation, err := evaluate(root, parsedPlan)
	if err != nil {
		return "", err
	}

	return strings.Join(evaluation.Lines, "\n"), nil
}

func evaluate(root *jsondoc.Node, parsedPlan *plan.Plan) (*Evaluation, error) {
	if parsedPlan.Version != 1 {
		return nil, diagnostics.New("unsupported_version", -1, "", "plan version %d is not supported", parsedPlan.Version)
	}

	consumed := make(map[string]struct{})
	lines := make([]string, 0)

	for index, directive := range parsedPlan.Directives {
		result, err := directives.Execute(root, index, directive)
		if err != nil {
			return nil, err
		}

		lines = append(lines, result.Lines...)
		for _, path := range result.Consumed {
			consumed[path] = struct{}{}
		}
	}

	for _, path := range root.LeafPaths(nil) {
		if _, ok := consumed[path]; !ok {
			return nil, diagnostics.New(
				"missing_coverage",
				-1,
				path,
				"plan does not cover JSON path %q",
				path,
			)
		}
	}

	return &Evaluation{Lines: lines}, nil
}
