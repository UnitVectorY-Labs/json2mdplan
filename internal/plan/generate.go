package plan

import (
	"fmt"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/jsondoc"
)

func Generate(root *jsondoc.Node) (*Plan, error) {
	switch root.Kind {
	case jsondoc.Object:
		fields := make([]Field, 0, len(root.Object))
		for _, field := range root.Object {
			if !field.Value.IsScalar() {
				return nil, fmt.Errorf("automatic plan generation only supports flat objects with scalar fields")
			}
			fields = append(fields, Field{
				Path:  field.Name,
				Label: field.Name,
			})
		}

		return &Plan{
			Version: 1,
			Directives: []Directive{
				{
					Op:     "named_bullets",
					Path:   ".",
					Fields: fields,
				},
			},
		}, nil
	case jsondoc.Array:
		for _, item := range root.Array {
			if !item.IsScalar() {
				return nil, fmt.Errorf("automatic plan generation only supports arrays of scalar values")
			}
		}

		return &Plan{
			Version: 1,
			Directives: []Directive{
				{
					Op:   "bullet_list",
					Path: ".",
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("automatic plan generation only supports root objects and arrays")
	}
}
