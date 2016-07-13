package yaml

import (
	"fmt"
	"testing"
)

const inYaml = `
named: &default
  key: value
ugh: # the worst
  <<: *default
  b: ok
  ? oh look
  : more syntax
  c:
  - !!str yaml stuff
  - "woo"
`

const outYaml = `
named: &default
  key: "ENC[value]"
ugh: # the worst
  <<: *default
  b: "ENC[ok]"
  ? oh look
  : "ENC[more syntax]"
  c:
  - !!str "ENC[yaml stuff]"
  - "ENC[woo]"
`

func TestTranform(t *testing.T) {
	xform := func(a []byte) ([]byte, error) {
		return []byte(fmt.Sprintf("ENC[%s]", []byte(a))), nil
	}
	svt := ScalarValueTransformer{}
	out, err := svt.TransformScalarValues([]byte(inYaml), xform)
	if err != nil {
		t.Errorf("unexpected err")
	}
	if string(out) != outYaml {
		t.Errorf("output mismatch. Got:\n========================\n%s", out)
	}
}
