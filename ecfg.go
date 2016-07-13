// Package ecfg implements the primary interface to interact with ecfg
// documents and keypairs. The CLI implemented by cmd/ecfg is a fairly thin
// wrapper around this package.
package ecfg

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Shopify/ecfg/pkg/crypto"
	"github.com/Shopify/ecfg/pkg/json"
)

type ScalarValueTransformer interface {
	TransformScalarValues([]byte, func([]byte) ([]byte, error))
}

// GenerateKeypair is used to create a new ecfg keypair. It returns the keys as
// hex-encoded strings, suitable for printing to the screen. hex.DecodeString
// can be used to load the true representation if necessary.
func GenerateKeypair() (pub string, priv string, err error) {
	var kp crypto.Keypair
	if err := kp.Generate(); err != nil {
		return "", "", err
	}
	return kp.PublicString(), kp.PrivateString(), nil
}

// EncryptFileInPlace takes a path to a file on disk, which must be a valid ecfg file
// (see README.md for more on what constitutes a valid ecfg file). Any
// encryptable-but-unencrypted fields in the file will be encrypted using the
// public key embdded in the file, and the resulting text will be written over
// the file present on disk.
func EncryptFileInPlace(filePath string) (int, error) {
	data, err := readFile(filePath)
	if err != nil {
		return -1, err
	}

	fileMode, err := getMode(filePath)
	if err != nil {
		return -1, err
	}

	var myKP crypto.Keypair
	if err := myKP.Generate(); err != nil {
		return -1, err
	}

	pubkey, err := json.ExtractPublicKey(data)
	if err != nil {
		return -1, err
	}

	encrypter := myKP.Encrypter(pubkey)

	svt := json.ScalarValueTransformer{}
	newdata, err := svt.TransformScalarValues(data, encrypter.Encrypt)
	if err != nil {
		return -1, err
	}

	if err := writeFile(filePath, newdata, fileMode); err != nil {
		return -1, err
	}

	return len(newdata), nil
}

// DecryptFile takes a path to an encrypted ecfg file and returns the data
// decrypted. The public key used to encrypt the values is embedded in the
// referenced document, and the matching private key is searched for in keydir.
// There must exist a file in keydir whose name is the public key from the
// ecfg document, and whose contents are the corresponding private key. See
// README.md for more details on this.
func DecryptFile(filePath, keydir string) ([]byte, error) {
	data, err := readFile(filePath)
	if err != nil {
		return nil, err
	}

	pubkey, err := json.ExtractPublicKey(data)
	if err != nil {
		return nil, err
	}

	privkey, err := findPrivateKey(pubkey, keydir)
	if err != nil {
		return nil, err
	}

	myKP := crypto.Keypair{
		Public:  pubkey,
		Private: privkey,
	}

	decrypter := myKP.Decrypter()

	svt := json.ScalarValueTransformer{}
	return svt.TransformScalarValues(data, decrypter.Decrypt)
}

func findPrivateKey(pubkey [32]byte, keydir string) (privkey [32]byte, err error) {
	keyFile := fmt.Sprintf("%s/%x", keydir, pubkey)
	var fileContents []byte
	fileContents, err = readFile(keyFile)
	if err != nil {
		err = fmt.Errorf("couldn't read key file (%s)", err.Error())
		return
	}

	bs, err := hex.DecodeString(strings.TrimSpace(string(fileContents)))
	if err != nil {
		return
	}

	if len(bs) != 32 {
		err = fmt.Errorf("invalid private key retrieved from keydir")
		return
	}

	copy(privkey[:], bs)
	return
}

// for mocking in tests
func _getMode(path string) (os.FileMode, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Mode(), nil
}

// for mocking in tests
var (
	readFile  = ioutil.ReadFile
	writeFile = ioutil.WriteFile
	getMode   = _getMode
)
