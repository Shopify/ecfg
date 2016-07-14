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
	"github.com/Shopify/ecfg/pkg/format"
	"github.com/Shopify/ecfg/pkg/json"
	"github.com/Shopify/ecfg/pkg/toml"
	"github.com/Shopify/ecfg/pkg/yaml"
)

type FileType int

const (
	FileTypeJSON = iota
	FileTypeYAML
	FileTypeTOML
)

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
func EncryptFileInPlace(filePath string, fileType FileType) (int, error) {
	data, err := readFile(filePath)
	if err != nil {
		return -1, err
	}

	fileMode, err := getMode(filePath)
	if err != nil {
		return -1, err
	}

	newdata, err := EncryptData(data, fileType)
	if err != nil {
		return -1, err
	}

	if err := writeFile(filePath, newdata, fileMode); err != nil {
		return -1, err
	}

	return len(newdata), nil
}

func EncryptData(data []byte, fileType FileType) ([]byte, error) {
	fh := handlerForType(fileType)

	var myKP crypto.Keypair
	if err := myKP.Generate(); err != nil {
		return nil, err
	}

	pubkey, err := fh.ExtractPublicKey(data)
	if err != nil {
		return nil, err
	}

	encrypter := myKP.Encrypter(pubkey)

	return fh.TransformScalarValues(data, encrypter.Encrypt)
}

// DecryptFile takes a path to an encrypted ecfg file and returns the data
// decrypted. The public key used to encrypt the values is embedded in the
// referenced document, and the matching private key is searched for in keydir.
// There must exist a file in keydir whose name is the public key from the
// ecfg document, and whose contents are the corresponding private key. See
// README.md for more details on this.
func DecryptFile(filePath, keydir string, fileType FileType) ([]byte, error) {
	data, err := readFile(filePath)
	if err != nil {
		return nil, err
	}

	return DecryptData(data, keydir, fileType)
}

func DecryptData(data []byte, keydir string, fileType FileType) ([]byte, error) {
	fh := handlerForType(fileType)

	pubkey, err := fh.ExtractPublicKey(data)
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

	return fh.TransformScalarValues(data, decrypter.Decrypt)
}

func findPrivateKey(pubkey [32]byte, keydir string) (privkey [32]byte, err error) {
	keyString := os.Getenv("ECFG_PRIVATE_KEY")
	if keyString == "" {
		keyFile := fmt.Sprintf("%s/%x", keydir, pubkey)
		var fileContents []byte
		fileContents, err = readFile(keyFile)
		if err != nil {
			err = fmt.Errorf("couldn't read key file (%s)", err.Error())
			return
		}
		keyString = strings.TrimSpace(string(fileContents))
	}

	bs, err := hex.DecodeString(keyString)
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

func handlerForType(typ FileType) format.FormatHandler {
	switch typ {
	case FileTypeJSON:
		return &json.FormatHandler{}
	case FileTypeYAML:
		return &yaml.FormatHandler{}
	case FileTypeTOML:
		return &toml.FormatHandler{}
	default:
		panic("bug: invalid file type")
	}
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
