package json

import (
	"reflect"
	"testing"
)

func TestKeyExtraction(t *testing.T) {
	// Key extraction succeeds when given properly-formatted ecfg
	in := `{"_public_key": "6d79b7e50073e5e66a4581ed08bf1d9a03806cc4648cffeb6df71b5775e5eb08"}`
	expected := [32]byte{109, 121, 183, 229, 0, 115, 229, 230, 106, 69, 129, 237, 8, 191, 29, 154, 3, 128, 108, 196, 100, 140, 255, 235, 109, 247, 27, 87, 117, 229, 235, 8}
	key, err := ExtractPublicKey([]byte(in))
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(key, expected) {
		t.Errorf("unexpected key: %#v", key)
	}

	// Key extraction fails if key is too short
	in = `{"_public_key": "6d79b7e50073e5e66a4581ed08bf1d9a03806cc4648cffeb6df71b5775e5eb0"}`
	_, err = ExtractPublicKey([]byte(in))
	if err != ErrPublicKeyInvalid {
		t.Errorf("expected ErrPublicKeyInvalid but got: %v", err)
	}

	// Key extraction fails if key is invalid hex
	in = `{"_public_key": "6d79b7e50073e5e66a45t1ed08bf1d9a03806cc4648cffeb6df71b5775e5eb08"}`
	_, err = ExtractPublicKey([]byte(in))
	if err != ErrPublicKeyInvalid {
		t.Errorf("expected ErrPublicKeyInvalid but got: %v", err)
	}

	// Key extraction fails if key is missing
	in = `{"nope": "dunno"}`
	_, err = ExtractPublicKey([]byte(in))
	if err != ErrPublicKeyMissing {
		t.Errorf("expected ErrPublicKeyMissing but got: %v", err)
	}
}
