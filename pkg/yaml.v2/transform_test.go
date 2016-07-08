package yaml

import (
	"fmt"
	"testing"
)

const inYaml = `
omg: &default
  a: words
ugh:
  <<: *default
  b: ok
  ? and this
  : bullshit
  c:
  - !!str yaml stuff
  - "woo"
`

const outYaml = `
omg: &default
  a: "ENC[words]"
ugh:
  <<: *default
  b: "ENC[ok]"
  ? and this
  : "ENC[bullshit]"
  c:
  - !!str "ENC[yaml stuff]"
  - "ENC[woo]"
`

func TestTranform(t *testing.T) {
	xform := func(a string) (string, error) {
		return fmt.Sprintf("\"ENC[%s]\"", a), nil
	}
	out, err := TransformValues(inYaml, xform)
	if err != nil {
		t.Errorf("unexpected err")
	}
	if out != outYaml {
		t.Errorf("output mismatch. Got:\n========================\n%s", out)
	}
}
