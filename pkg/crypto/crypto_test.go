package crypto

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestKeypairGeneration(t *testing.T) {
	var kp Keypair

	err := kp.Generate()
	assertNoError(t, err)

	if kp.PublicString() == kp.PrivateString() {
		t.Errorf("public and private keys are the same")
	}
	if strings.Contains(kp.PublicString(), "00000") {
		t.Errorf("public key looks sketchy")
	}
	if strings.Contains(kp.PrivateString(), "00000") {
		t.Errorf("private key looks sketchy")
	}

	if kp.Public[0] == 0 && kp.Public[1] == 0 && kp.Public[2] == 0 {
		t.Errorf("public key is null")
	}
	if kp.Private[0] == 0 && kp.Private[1] == 0 && kp.Private[2] == 0 {
		t.Errorf("private key is null")
	}
}

func TestNonceGeneration(t *testing.T) {
	// generated nonces should be unique
	n1, _ := genNonce()
	n2, _ := genNonce()
	if reflect.DeepEqual(n1, n2) {
		t.Errorf("nonces were equal!")
	}

	// generated nonces should pass a super basic sanity check
	n, err := genNonce()
	assertNoError(t, err)
	text := fmt.Sprintf("%x", n)
	if strings.Contains(text, "00000") {
		t.Errorf("nonce looks sketchy: %s", text)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRoundtrip(t *testing.T) {
	var kpEphemeral, kpSecret Keypair
	kpEphemeral.Generate()
	kpSecret.Generate()

	encrypter := kpEphemeral.Encrypter(kpSecret.Public)
	decrypter := kpSecret.Decrypter()
	message := []byte("This is a test of the emergency broadcast system.")
	ct, err := encrypter.Encrypt(message)
	assertNoError(t, err)
	ct2, err := encrypter.Encrypt(ct) // this one will leave the message unchanged
	assertNoError(t, err)
	if !reflect.DeepEqual(ct2, ct) {
		t.Errorf("unexpected ciphertext")
	}
	pt, err := decrypter.Decrypt(ct2)
	assertNoError(t, err)
	if !reflect.DeepEqual(pt, message) {
		t.Errorf("unexpected plaintext")
	}
	if reflect.DeepEqual(pt, ct) {
		t.Errorf("ciphertext shouldn't equal plaintext")
	}
	if len(ct) <= len(pt) {
		t.Errorf("ciphertext should be longer than plaintext")
	}
}

func ExampleEncrypt(peerPublic [32]byte) {
	var kp Keypair
	if err := kp.Generate(); err != nil {
		panic(err)
	}

	encrypter := kp.Encrypter(peerPublic)
	boxed, err := encrypter.Encrypt([]byte("this is my message"))
	fmt.Println(boxed, err)
}

func ExampleDecrypt(myPublic, myPrivate [32]byte, encrypted []byte) {
	kp := Keypair{
		Public:  myPublic,
		Private: myPrivate,
	}

	decrypter := kp.Decrypter()
	plaintext, err := decrypter.Decrypt(encrypted)
	fmt.Println(plaintext, err)
}
