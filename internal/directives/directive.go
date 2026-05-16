package directives

import (
	"fmt"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/diagnostics"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/jsondoc"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/plan"
)

type Result struct {
	Lines    []string
	Consumed []string
}

type Handler interface {
	Execute(root *jsondoc.Node, directiveIndex int, directive plan.Directive) (*Result, error)
}

var handlers = map[string]Handler{
	"bullet_list":   bulletListHandler{},
	"named_bullets": namedBulletsHandler{},
	"table":         tableHandler{},
}

func Execute(root *jsondoc.Node, directiveIndex int, directive plan.Directive) (*Result, error) {
	handler, ok := handlers[directive.Op]
	if !ok {
		return nil, diagnostics.New(
			"unknown_directive",
			directiveIndex,
			directive.Path,
			"directive %q is not supported",
			directive.Op,
		)
	}

	return handler.Execute(root, directiveIndex, directive)
}

func requirePath(root *jsondoc.Node, directiveIndex int, expr string, expected jsondoc.Kind, op string) (*jsondoc.Node, string, error) {
	node, absolutePath, err := jsondoc.Resolve(root, root, nil, expr)
	if err != nil {
		return nil, "", diagnostics.New(
			"invalid_path",
			directiveIndex,
			expr,
			"path %q could not be resolved: %s",
			expr,
			err.Error(),
		)
	}

	if node.Kind != expected {
		return nil, "", diagnostics.New(
			"type_mismatch",
			directiveIndex,
			expr,
			"directive %q requires path %q to resolve to %s",
			op,
			displayPath(expr),
			expected,
		)
	}

	return node, absolutePath, nil
}

func displayPath(path string) string {
	if path == "" {
		return "."
	}
	return path
}

func missingFieldError(directiveIndex int, path string) error {
	return diagnostics.New(
		"missing_field",
		directiveIndex,
		path,
		"field path %q does not exist relative to %q",
		path,
		".",
	)
}

func nonScalarFieldError(directiveIndex int, path string) error {
	return diagnostics.New(
		"non_scalar_field",
		directiveIndex,
		path,
		"field path %q must resolve to a scalar value",
		path,
	)
}

func unexpectedPlanShape(directiveIndex int, path string, op string, problem string) error {
	return diagnostics.New(
		"invalid_plan",
		directiveIndex,
		path,
		"directive %q is invalid: %s",
		op,
		problem,
	)
}

func formatBullet(value string) string {
	return fmt.Sprintf("- %s", value)
}
