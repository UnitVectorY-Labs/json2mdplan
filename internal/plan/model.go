package plan

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

type Plan struct {
	Version    int         `json:"version"`
	Directives []Directive `json:"directives"`
}

type Directive struct {
	Op     string  `json:"op"`
	Path   string  `json:"path"`
	Fields []Field `json:"fields,omitempty"`
}

type Field struct {
	Path  string `json:"path"`
	Label string `json:"label"`
}

func Parse(data []byte) (*Plan, error) {
	var parsed Plan
	if err := json.Unmarshal(data, &parsed, json.RejectUnknownMembers(true)); err != nil {
		return nil, err
	}

	return &parsed, nil
}

func Marshal(pretty Plan) ([]byte, error) {
	return json.Marshal(pretty, jsontext.WithIndent("  "))
}
