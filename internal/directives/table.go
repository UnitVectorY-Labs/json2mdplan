package directives

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/jsondoc"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/plan"
)

type tableHandler struct{}

func (tableHandler) Execute(root *jsondoc.Node, directiveIndex int, directive plan.Directive) (*Result, error) {
	if len(directive.Fields) == 0 {
		return nil, unexpectedPlanShape(directiveIndex, directive.Path, directive.Op, "fields must not be empty")
	}

	target, absolutePath, err := requirePath(root, directiveIndex, directive.Path, jsondoc.Array, directive.Op)
	if err != nil {
		return nil, err
	}

	// Build header row
	headers := make([]string, 0, len(directive.Fields))
	for _, field := range directive.Fields {
		headers = append(headers, field.Label)
	}

	lines := make([]string, 0, len(target.Array)+2)
	lines = append(lines, "| "+strings.Join(headers, " | ")+" |")

	separators := make([]string, 0, len(directive.Fields))
	for range directive.Fields {
		separators = append(separators, "---")
	}
	lines = append(lines, "| "+strings.Join(separators, " | ")+" |")

	consumed := make([]string, 0)

	for rowIndex, item := range target.Array {
		if item.Kind != jsondoc.Object {
			return nil, unexpectedPlanShape(directiveIndex, directive.Path, directive.Op, "all array items must be objects")
		}

		rowPrefix := absolutePath + "/" + strconv.Itoa(rowIndex)

		cells := make([]string, 0, len(directive.Fields))
		for _, field := range directive.Fields {
			if field.Path == "" || field.Path == "." {
				return nil, unexpectedPlanShape(directiveIndex, directive.Path, directive.Op, "field paths must not be empty")
			}

			rowTokens, err := jsondoc.PointerTokens(rowPrefix)
			if err != nil {
				return nil, err
			}

			node, fieldAbsPath, err := jsondoc.Resolve(root, item, rowTokens, field.Path)
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

			cells = append(cells, value)
			consumed = append(consumed, fieldAbsPath)
		}

		lines = append(lines, fmt.Sprintf("| %s |", strings.Join(cells, " | ")))
	}

	return &Result{
		Lines:    lines,
		Consumed: consumed,
	}, nil
}
