package jsondoc

import "testing"

func TestParseStrictJSON(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
	}{
		{"duplicate names", `{"a":1,"\u0061":2}`},
		{"nested duplicate names", `{"a":[{"b":1,"b":2}]}`},
		{"invalid UTF-8", "{\"a\":\"\xff\"}"},
		{"trailing value", `{} []`},
		{"unterminated object", `{"a":1`},
		{"unterminated array", `[1`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse([]byte(tc.input)); err == nil {
				t.Fatal("expected invalid JSON to be rejected")
			}
		})
	}
}

func TestParsePreservesOrderAndNumberText(t *testing.T) {
	root, err := Parse([]byte(`{"z":1.2300e+40,"a":[true,false,null,"text"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Object) != 2 || root.Object[0].Name != "z" || root.Object[0].Value.Number != "1.2300e+40" || root.Object[1].Name != "a" {
		t.Fatalf("unexpected parsed object: %#v", root.Object)
	}
	items := root.Object[1].Value.Array
	if len(items) != 4 || !items[0].Bool || items[1].Bool || items[2].Kind != Null || items[3].String != "text" {
		t.Fatalf("unexpected parsed array: %#v", items)
	}
}
