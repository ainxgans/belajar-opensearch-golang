package product

import (
	"reflect"
	"testing"
)

func TestParsePGTextArray(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{`{}`, nil},
		{`{new,sale}`, []string{"new", "sale"}},
		{`{"back to school"}`, []string{"back to school"}}, // whitespace -> quoted
		{`{"a,b",c}`, []string{"a,b", "c"}},                // comma inside element
		{"{\"say \\\"hi\\\"\"}", []string{`say "hi"`}},     // {"say \"hi\""} -> backslash-escaped quotes
		{"{\"a\\\\b\"}", []string{`a\b`}},                  // {"a\\b"} -> escaped backslash
	}
	for _, tc := range cases {
		if got := parsePGTextArray(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parsePGTextArray(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}
