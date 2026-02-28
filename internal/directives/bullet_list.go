package directives

import (
	"strconv"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/diagnostics"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/jsondoc"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/plan"
)

type bulletListHandler struct{}

func (bulletListHandler) Execute(root *jsondoc.Node, directiveIndex int, directive plan.Directive) (*Result, error) {
	if len(directive.Fields) > 0 {
		return nil, unexpectedPlanShape(directiveIndex, directive.Path, directive.Op, "fields are not supported")
	}

	target, absolutePath, err := requirePath(root, directiveIndex, directive.Path, jsondoc.Array, directive.Op)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0, len(target.Array))
	consumed := make([]string, 0, len(target.Array))

	for index, item := range target.Array {
		if !item.IsScalar() {
			return nil, diagnostics.New(
				"non_scalar_item",
				directiveIndex,
				directive.Path,
				"directive %q requires all array items at path %q to be scalar values",
				directive.Op,
				displayPath(directive.Path),
			)
		}

		value, err := item.FormatScalar()
		if err != nil {
			return nil, err
		}

		lines = append(lines, formatBullet(value))
		consumed = append(consumed, absolutePath+"/"+strconv.Itoa(index))
	}

	return &Result{
		Lines:    lines,
		Consumed: consumed,
	}, nil
}
