package crypto

import (
	"reflect"
	"testing"
)

func TestBoxedMessageRoundtripping(t *testing.T) {
	pk := [32]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	nonce := [24]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	wire := "EJ[1:AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=:AgICAgICAgICAgICAgICAgICAgICAgIC:AwMD]"

	bm := boxedMessage{
		SchemaVersion:   1,
		EncrypterPublic: pk,
		Nonce:           nonce,
		Box:             []byte{3, 3, 3},
	}

	// Dump
	if string(bm.Dump()) != wire {
		t.Errorf("boxedmessage didn't serialize the way we expected")
	}

	// Load
	bm = boxedMessage{}
	if err := bm.Load([]byte(wire)); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(bm.EncrypterPublic, pk) {
		t.Errorf("unexpected EncrypterPublic")
	}
	if !reflect.DeepEqual(bm.Nonce, nonce) {
		t.Errorf("unexpected Nonce")
	}
	if !reflect.DeepEqual(bm.Box, []byte{3, 3, 3}) {
		t.Errorf("unexpected Box")
	}

	// isBoxedMessage
	if !isBoxedMessage([]byte(wire)) {
		t.Errorf("isBoxedMessage incorrect")
	}
	if isBoxedMessage([]byte("nope")) {
		t.Errorf("isBoxedMessage incorrect")
	}
	if isBoxedMessage([]byte("EJ[]")) {
		t.Errorf("isBoxedMessage incorrect")
	}
	if !isBoxedMessage([]byte("EJ[1:12345678901234567890123456789012345678901234:12345678901234567890123456789012:a]")) {
		t.Errorf("isBoxedMessage incorrect")
	}
}
