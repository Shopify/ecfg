package toml

import (
	"fmt"
	"testing"
)

const inToml = `
# This is a TOML document. Boom.

_public_key = "1234"

title = "TOML Example"

[owner]
name = "Tom Preston-Werner"
organization = "GitHub"
bio = '''
GitHub Cofounder & CEO
Likes tater tots and beer.'''
dob = 1979-05-27T07:32:00Z # First class dates? Why not?

[database]
server = '192.168.1.1'
ports = [ 8001, 8001, 8002 ]
connection_max = 5000
enabled = true

[servers]

  # You can indent as you please. Tabs or spaces. TOML don't care.
  [servers.alpha]
  ip = "10.\t0.0.1"
  dc = 'eqd\tc10'

  [servers.beta]
  ip = "10.0.0.2"
  dc = "eqdc10"

[clients]
data = [ ["gamma", "delta"], [1, 2] ] # just an update to make sure parsers support it

# Line breaks are OK when inside arrays
hosts = [
  "alpha",
  "omega"
]
`

const outToml = `
# This is a TOML document. Boom.

_public_key = "1234"

title = "ENC[TOML Example]"

[owner]
name = "ENC[Tom Preston-Werner]"
organization = "ENC[GitHub]"
bio = "ENC[GitHub Cofounder & CEO\nLikes tater tots and beer.]"
dob = 1979-05-27T07:32:00Z # First class dates? Why not?

[database]
server = "ENC[192.168.1.1]"
ports = [ 8001, 8001, 8002 ]
connection_max = 5000
enabled = true

[servers]

  # You can indent as you please. Tabs or spaces. TOML don't care.
  [servers.alpha]
  ip = "ENC[10.\t0.0.1]"
  dc = "ENC[eqd\\tc10]"

  [servers.beta]
  ip = "ENC[10.0.0.2]"
  dc = "ENC[eqdc10]"

[clients]
data = [ ["ENC[gamma]", "ENC[delta]"], [1, 2] ] # just an update to make sure parsers support it

# Line breaks are OK when inside arrays
hosts = [
  "ENC[alpha]",
  "ENC[omega]"
]
`

func TestTranform(t *testing.T) {
	xform := func(a []byte) ([]byte, error) {
		return []byte(fmt.Sprintf("ENC[%s]", []byte(a))), nil
	}
	fh := FormatHandler{}
	out, err := fh.TransformScalarValues([]byte(inToml), xform)
	if err != nil {
		t.Errorf("unexpected err")
	}
	if string(out) != outToml {
		t.Errorf("output mismatch. Got:\n========================\n%s", out)
	}
}
