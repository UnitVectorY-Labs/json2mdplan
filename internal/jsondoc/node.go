package jsondoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Kind string

const (
	Object  Kind = "object"
	Array   Kind = "array"
	String  Kind = "string"
	Number  Kind = "number"
	Boolean Kind = "boolean"
	Null    Kind = "null"
)

type Node struct {
	Kind   Kind
	Object []Field
	Array  []*Node
	String string
	Number string
	Bool   bool
}

type Field struct {
	Name  string
	Value *Node
}

func Parse(data []byte) (*Node, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	node, err := parseValue(dec)
	if err != nil {
		return nil, err
	}

	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("unexpected trailing JSON content")
		}
		return nil, err
	}

	return node, nil
}

func parseValue(dec *json.Decoder) (*Node, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			fields := make([]Field, 0)
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}

				key, ok := keyTok.(string)
				if !ok {
					return nil, fmt.Errorf("object key must be a string")
				}

				value, err := parseValue(dec)
				if err != nil {
					return nil, err
				}

				fields = append(fields, Field{Name: key, Value: value})
			}

			endTok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			if endTok != json.Delim('}') {
				return nil, fmt.Errorf("expected object end")
			}

			return &Node{Kind: Object, Object: fields}, nil
		case '[':
			items := make([]*Node, 0)
			for dec.More() {
				item, err := parseValue(dec)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			}

			endTok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			if endTok != json.Delim(']') {
				return nil, fmt.Errorf("expected array end")
			}

			return &Node{Kind: Array, Array: items}, nil
		default:
			return nil, fmt.Errorf("unexpected delimiter %q", string(v))
		}
	case string:
		return &Node{Kind: String, String: v}, nil
	case json.Number:
		return &Node{Kind: Number, Number: v.String()}, nil
	case bool:
		return &Node{Kind: Boolean, Bool: v}, nil
	case nil:
		return &Node{Kind: Null}, nil
	default:
		return nil, fmt.Errorf("unsupported JSON token type %T", tok)
	}
}

func (n *Node) IsScalar() bool {
	return n.Kind != Object && n.Kind != Array
}

func (n *Node) FindField(name string) (*Node, bool) {
	if n.Kind != Object {
		return nil, false
	}

	for _, field := range n.Object {
		if field.Name == name {
			return field.Value, true
		}
	}

	return nil, false
}

func (n *Node) FormatScalar() (string, error) {
	switch n.Kind {
	case String:
		return n.String, nil
	case Number:
		return n.Number, nil
	case Boolean:
		if n.Bool {
			return "true", nil
		}
		return "false", nil
	case Null:
		return "null", nil
	default:
		return "", fmt.Errorf("node kind %q is not scalar", n.Kind)
	}
}

func (n *Node) LeafPaths(base []string) []string {
	switch n.Kind {
	case Object:
		paths := make([]string, 0)
		for _, field := range n.Object {
			paths = append(paths, field.Value.LeafPaths(appendToken(base, field.Name))...)
		}
		return paths
	case Array:
		paths := make([]string, 0, len(n.Array))
		for i, item := range n.Array {
			paths = append(paths, item.LeafPaths(appendToken(base, strconv.Itoa(i)))...)
		}
		return paths
	default:
		return []string{EncodePointer(base)}
	}
}

func Resolve(root *Node, current *Node, currentTokens []string, expr string) (*Node, string, error) {
	switch {
	case expr == "" || expr == ".":
		return current, EncodePointer(currentTokens), nil
	case strings.HasPrefix(expr, "/"):
		tokens, err := parseAbsolutePointer(expr)
		if err != nil {
			return nil, "", err
		}
		node, absTokens, err := navigate(root, tokens, nil)
		if err != nil {
			return nil, "", err
		}
		return node, EncodePointer(absTokens), nil
	default:
		tokens := parseRelativePath(expr)
		node, absTokens, err := navigate(current, tokens, currentTokens)
		if err != nil {
			return nil, "", err
		}
		return node, EncodePointer(absTokens), nil
	}
}

func PointerTokens(pointer string) ([]string, error) {
	if pointer == "" {
		return nil, nil
	}

	return parseAbsolutePointer(pointer)
}

func EncodePointer(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}

	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, escapeToken(token))
	}

	return "/" + strings.Join(parts, "/")
}

func appendToken(tokens []string, token string) []string {
	out := make([]string, len(tokens)+1)
	copy(out, tokens)
	out[len(tokens)] = token
	return out
}

func parseAbsolutePointer(expr string) ([]string, error) {
	if expr == "" {
		return nil, nil
	}
	if !strings.HasPrefix(expr, "/") {
		return nil, fmt.Errorf("absolute JSON pointer must start with '/'")
	}

	raw := strings.Split(expr[1:], "/")
	tokens := make([]string, 0, len(raw))
	for _, token := range raw {
		tokens = append(tokens, unescapeToken(token))
	}
	return tokens, nil
}

func parseRelativePath(expr string) []string {
	if expr == "" || expr == "." {
		return nil
	}

	raw := strings.Split(expr, "/")
	tokens := make([]string, 0, len(raw))
	for _, token := range raw {
		if token == "" || token == "." {
			continue
		}
		tokens = append(tokens, unescapeToken(token))
	}
	return tokens
}

func navigate(node *Node, tokens []string, base []string) (*Node, []string, error) {
	current := node
	abs := append([]string{}, base...)

	for _, token := range tokens {
		switch current.Kind {
		case Object:
			next, ok := current.FindField(token)
			if !ok {
				return nil, nil, fmt.Errorf("field %q does not exist", token)
			}
			current = next
			abs = append(abs, token)
		case Array:
			index, err := strconv.Atoi(token)
			if err != nil {
				return nil, nil, fmt.Errorf("array index %q is invalid", token)
			}
			if index < 0 || index >= len(current.Array) {
				return nil, nil, fmt.Errorf("array index %q is out of bounds", token)
			}
			current = current.Array[index]
			abs = append(abs, token)
		default:
			return nil, nil, fmt.Errorf("cannot descend into %q", current.Kind)
		}
	}

	return current, abs, nil
}

func escapeToken(token string) string {
	token = strings.ReplaceAll(token, "~", "~0")
	token = strings.ReplaceAll(token, "/", "~1")
	return token
}

func unescapeToken(token string) string {
	token = strings.ReplaceAll(token, "~1", "/")
	token = strings.ReplaceAll(token, "~0", "~")
	return token
}
