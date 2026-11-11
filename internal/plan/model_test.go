package plan

import (
	"strings"
	"testing"
)

func TestParseRejectsAmbiguousJSON(t *testing.T) {
	for _, input := range []string{
		`{"version":1,"version":2}`,
		`{"Version":1}`,
		`{"version":1,"unknown":true}`,
		`{"version":1} {}`,
		"{\"directives\":[{\"op\":\"\xff\"}]}",
	} {
		if _, err := Parse([]byte(input)); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", input)
		}
	}
}

func TestMarshalV2Output(t *testing.T) {
	data, err := Marshal(Plan{Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{\n  \"version\": 1,\n  \"directives\": []\n}" {
		t.Fatalf("unexpected nil slice output: %s", data)
	}
	data, err = Marshal(Plan{Directives: []Directive{{Op: "<tag>", Fields: []Field{}}}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"fields"`) || !strings.Contains(string(data), `<tag>`) {
		t.Fatalf("unexpected output: %s", data)
	}
}
