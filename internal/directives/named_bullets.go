package directives

import (
	"fmt"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/jsondoc"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/plan"
)

type namedBulletsHandler struct{}

func (namedBulletsHandler) Execute(root *jsondoc.Node, directiveIndex int, directive plan.Directive) (*Result, error) {
	if len(directive.Fields) == 0 {
		return nil, unexpectedPlanShape(directiveIndex, directive.Path, directive.Op, "fields must not be empty")
	}

	target, targetPath, err := requirePath(root, directiveIndex, directive.Path, jsondoc.Object, directive.Op)
	if err != nil {
		return nil, err
	}
	targetTokens, err := jsondoc.PointerTokens(targetPath)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0, len(directive.Fields))
	consumed := make([]string, 0, len(directive.Fields))

	for _, field := range directive.Fields {
		if field.Path == "" || field.Path == "." {
			return nil, unexpectedPlanShape(directiveIndex, directive.Path, directive.Op, "field paths must not be empty")
		}

		node, absolutePath, err := jsondoc.Resolve(root, target, targetTokens, field.Path)
		if err != nil {
			return nil, missingFieldError(directiveIndex, field.Path)
		}
		if !node.IsScalar() {
			return nil, nonScalarFieldError(directiveIndex, field.Path)
		}

		value, err := node.FormatScalar()
		if err != nil {
			return nil, err
		}

		lines = append(lines, formatBullet(fmt.Sprintf("**%s:** %s", field.Label, value)))
		consumed = append(consumed, absolutePath)
	}

	return &Result{
		Lines:    lines,
		Consumed: consumed,
	}, nil
}
