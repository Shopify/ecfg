package ecfg

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestGenerateKeypair(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	assertNoError(t, err)
	if pub == priv {
		t.Errorf("pub == priv")
	}
	if strings.Contains(pub, "00000") {
		t.Errorf("pubkey looks sketchy")
	}
	if strings.Contains(priv, "00000") {
		t.Errorf("privkey looks sketchy")
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

}

func TestEncryptFileInPlace(t *testing.T) {
	getMode = func(p string) (os.FileMode, error) {
		return 0400, nil
	}
	defer func() { getMode = _getMode }()

	_, err := EncryptFileInPlace("/does/not/exist")
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist, got %v", err)
	}

	// invalid json file
	readFile = func(p string) ([]byte, error) {
		return []byte(`{"a": "b"]`), nil
	}
	_, err = EncryptFileInPlace("/doesnt/matter")
	readFile = ioutil.ReadFile
	if err == nil {
		t.Errorf("expected error, but none was received")
	} else {
		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("wanted json error, but got %v", err)
		}
	}

	// invalid key
	readFile = func(p string) ([]byte, error) {
		return []byte(`{"_public_key": "invalid"}`), nil
	}
	_, err = EncryptFileInPlace("/doesnt/matter")
	readFile = ioutil.ReadFile
	if err == nil {
		t.Errorf("expected error, but none was received")
	} else {
		if !strings.Contains(err.Error(), "public key has invalid format") {
			t.Errorf("wanted key error, but got %v", err)
		}
	}

	// valid keypair
	readFile = func(p string) ([]byte, error) {
		return []byte(`{"_public_key": "8d8647e2eeb6d2e31228e6df7da3df921ec3b799c3f66a171cd37a1ed3004e7d", "a": "b"}`), nil
	}

	var output []byte
	writeFile = func(path string, data []byte, mode os.FileMode) error {
		output = data
		return nil
	}
	_, err = EncryptFileInPlace("/doesnt/matter")
	readFile = ioutil.ReadFile
	writeFile = ioutil.WriteFile
	assertNoError(t, err)
	match := regexp.MustCompile(`{"_public_key": "8d8.*", "a": "EJ.*"}`)
	if match.Find(output) == nil {
		t.Errorf("unexpected output: %s", output)
	}

}

func TestDecryptFile(t *testing.T) {

	_, err := DecryptFile("/does/not/exist", "/doesnt/matter")
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist, but got %v", err)
	}

	// invalid json file
	readFile = func(p string) ([]byte, error) {
		return []byte(`{"a": "b"]`), nil
	}
	_, err = DecryptFile("/doesnt/matter", "/doesnt/matter")
	readFile = ioutil.ReadFile
	if err == nil {
		t.Errorf("expected error, but none was received")
	} else {
		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("wanted json error, but got %v", err)
		}
	}

	readFile = func(p string) ([]byte, error) {
		return []byte(`{"_public_key": "invalid"}`), nil
	}
	_, err = DecryptFile("/doesnt/matter", "/doesnt/matter")
	readFile = ioutil.ReadFile
	if err == nil {
		t.Errorf("expected error, but none was received")
	} else {
		if !strings.Contains(err.Error(), "public key has invalid format") {
			t.Errorf("wanted key error, but got %v", err)
		}
	}

	// valid keypair but no corresponding entry in keydir
	readFile = func(p string) ([]byte, error) {
		if p == "a" {
			return []byte(`{"_public_key": "8d8647e2eeb6d2e31228e6df7da3df921ec3b799c3f66a171cd37a1ed3004e7d", "a": "b"}`), nil
		}
		return ioutil.ReadFile("/does/not/exist")
	}
	_, err = DecryptFile("a", "b")
	readFile = ioutil.ReadFile
	if err == nil {
		t.Errorf("expected error, but none was received")
	} else {
		if !strings.Contains(err.Error(), "couldn't read key file") {
			t.Errorf("wanted key file error, but got %v", err)
		}
	}

	// valid keypair and a corresponding entry in keydir
	readFile = func(p string) ([]byte, error) {
		if p == "a" {
			return []byte(`{"_public_key": "8d8647e2eeb6d2e31228e6df7da3df921ec3b799c3f66a171cd37a1ed3004e7d", "a": "EJ[1:KR1IxNZnTZQMP3OR1NdOpDQ1IcLD83FSuE7iVNzINDk=:XnYW1HOxMthBFMnxWULHlnY4scj5mNmX:ls1+kvwwu2ETz5C6apgWE7Q=]"}`), nil
		}
		return []byte("c5caa31a5b8cb2be0074b37c56775f533b368b81d8fd33b94181f79bd6e47f87"), nil
	}
	out, err := DecryptFile("a", "b")
	readFile = ioutil.ReadFile
	assertNoError(t, err)
	if string(out) != `{"_public_key": "8d8647e2eeb6d2e31228e6df7da3df921ec3b799c3f66a171cd37a1ed3004e7d", "a": "b"}` {
		t.Errorf("unexpected output")
	}
}
