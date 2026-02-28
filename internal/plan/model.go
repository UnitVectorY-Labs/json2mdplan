package plan

import (
	"bytes"
	"encoding/json"
	"io"
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
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var parsed Plan
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}

	return &parsed, nil
}

func Marshal(pretty Plan) ([]byte, error) {
	return json.MarshalIndent(pretty, "", "  ")
}
