package format

import (
	"encoding/hex"
	"errors"
)

const (
	// PublicKeyField is the key name at which the public key should be
	// stored in an ecfg document.
	PublicKeyField = "_public_key"
)

// ErrPublicKeyMissing indicates that the PublicKeyField key was not found
// at the top level of the JSON document provided.
var ErrPublicKeyMissing = errors.New("public key not present in ecfg file")

// ErrPublicKeyInvalid means that the PublicKeyField key was found, but the
// value could not be parsed into a valid key.
var ErrPublicKeyInvalid = errors.New("public key has invalid format")

type FormatHandler interface {
	TransformScalarValues([]byte, func([]byte) ([]byte, error)) ([]byte, error)
	ExtractPublicKey([]byte) ([32]byte, error)
}

func ExtractPublicKeyHelper(obj map[string]interface{}) (key [32]byte, err error) {
	var (
		ks string
		ok bool
		bs []byte
	)
	k, ok := obj[PublicKeyField]
	if !ok {
		goto missing
	}
	ks, ok = k.(string)
	if !ok {
		goto invalid
	}
	if len(ks) != 64 {
		goto invalid
	}
	bs, err = hex.DecodeString(ks)
	if err != nil {
		goto invalid
	}
	if len(bs) != 32 {
		goto invalid
	}
	copy(key[:], bs)
	return
missing:
	err = ErrPublicKeyMissing
	return
invalid:
	err = ErrPublicKeyInvalid
	return
}
