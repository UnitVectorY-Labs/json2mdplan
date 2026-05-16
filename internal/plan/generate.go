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
		// Check if all items are scalar values
		allScalar := true
		for _, item := range root.Array {
			if !item.IsScalar() {
				allScalar = false
				break
			}
		}

		if allScalar {
			return &Plan{
				Version: 1,
				Directives: []Directive{
					{
						Op:   "bullet_list",
						Path: ".",
					},
				},
			}, nil
		}

		// Check if all items are flat objects with only scalar fields
		allFlatObjects := true
		for _, item := range root.Array {
			if item.Kind != jsondoc.Object {
				allFlatObjects = false
				break
			}
			for _, field := range item.Object {
				if !field.Value.IsScalar() {
					allFlatObjects = false
					break
				}
			}
			if !allFlatObjects {
				break
			}
		}

		if allFlatObjects && len(root.Array) > 0 {
			// Collect field names from all objects in order of first appearance
			seen := make(map[string]struct{})
			fieldNames := make([]string, 0)
			for _, item := range root.Array {
				for _, field := range item.Object {
					if _, ok := seen[field.Name]; !ok {
						seen[field.Name] = struct{}{}
						fieldNames = append(fieldNames, field.Name)
					}
				}
			}

			fields := make([]Field, 0, len(fieldNames))
			for _, name := range fieldNames {
				fields = append(fields, Field{
					Path:  name,
					Label: name,
				})
			}

			return &Plan{
				Version: 1,
				Directives: []Directive{
					{
						Op:     "table",
						Path:   ".",
						Fields: fields,
					},
				},
			}, nil
		}

		return nil, fmt.Errorf("automatic plan generation only supports arrays of scalar values or arrays of flat objects")
	default:
		return nil, fmt.Errorf("automatic plan generation only supports root objects and arrays")
	}
}
