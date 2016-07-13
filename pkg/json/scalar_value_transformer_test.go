package json

import (
	"testing"
)

func TestScalarValueTransformer(t *testing.T) {
	action := func(a []byte) ([]byte, error) {
		return []byte{'E'}, nil
	}

	for _, tc := range testCases {
		svt := ScalarValueTransformer{}
		act, err := svt.TransformScalarValues([]byte(tc.in), action)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if string(act) != tc.out {
			t.Errorf("unexpected output: '%s'; wanted '%s'", string(act), tc.out)
		}
	}
}

type testCase struct {
	in, out string
}

// "E" means encrypted.
var testCases = []testCase{
	{`{"a": "b"}`, `{"a": "E"}`},                     // encryption
	{`{"a" : "b"}`, `{"a" : "E"}`},                   // weird spacing
	{` {  "a"  :"b" } `, ` {  "a"  :"E"}`},           // we could but don't preserve trailing spaces
	{`{"_a": "b"}`, `{"_a": "b"}`},                   // commenting
	{`{"a": "b", "c": "d"}`, `{"a": "E", "c": "E"}`}, // order-dependence
	{`{"a": 1}`, `{"a": 1}`},                         // numbers
	{`{"a": true}`, `{"a": true}`},                   // booleans
	{`{"a": ["b", "c"]}`, `{"a": ["E", "E"]}`},       // encrypting arrays
	{`{"_a": ["b", "c"]}`, `{"_a": ["b", "c"]}`},     // commenting arrays
	{`{"a": {"b": "c"}}`, `{"a": {"b": "E"}}`},       // nesting
	{`{"a": {"_b": "c"}}`, `{"a": {"_b": "c"}}`},     // nested comment
	{`{"_a": {"b": "c"}}`, `{"_a": {"b": "E"}}`},     // comments don't inherit
}
