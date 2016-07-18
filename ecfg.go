// Package ecfg implements the primary interface to interact with ecfg
// documents and keypairs. The CLI implemented by cmd/ecfg is a fairly thin
// wrapper around this package.
package ecfg

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
// referenced document, and the matching private key is searched for in
// keypath. There must exist a file in at least one of the keypath entries
// whose name is the public key from the ecfg document, and whose contents are
// the corresponding private key. See README.md for more details on this.
func DecryptFile(filePath string, keypath []string, fileType FileType) ([]byte, error) {
	data, err := readFile(filePath)
	if err != nil {
		return nil, err
	}

	return DecryptData(data, keypath, fileType)
}

// DecryptData takes a an encrypted ecfg document and returns the same
// document, decrypted. The public key used to encrypt the values is embedded
// in the document, and the matching private key is searched for in keypath.
// There must exist a file in at least one of the keypath entries whose name is
// the public key from the ecfg document, and whose contents are the
// corresponding private key. See README.md for more details on this.
func DecryptData(data []byte, keypath []string, fileType FileType) ([]byte, error) {
	fh := handlerForType(fileType)

	pubkey, err := fh.ExtractPublicKey(data)
	if err != nil {
		return nil, err
	}

	privkey, err := findPrivateKey(pubkey, keypath)
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

// DefaultKeypath is UserKeypath prefixed to SystemKeypath. For root, this will
// be equal to SystemKeypath, and for other users, this will cause key lookups
// to first try their own local keys, falling back to system keys if that
// fails.
func DefaultKeypath() (keypath []string) {
	for _, elem := range UserKeypath() {
		keypath = append(keypath, elem)
	}
	for _, elem := range SystemKeypath() {
		keypath = append(keypath, elem)
	}
	return
}

// UserKeypath returns the user-specific locations at which to search for ecfg
// keys. In most cases, this is empty for root, and ~/.ecfg/keys in other cases.
// If XDG_CONFIG_HOME is set, $XDG_CONFIG_HOME/ecfg/keys is highest priority.
func UserKeypath() (keypath []string) {
	// no user keypath entries for root.
	if getuid() == 0 {
		return
	}

	xdgConfigHome := getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		keypath = append(keypath, filepath.Join(xdgConfigHome, "ecfg", "keys"))
	}
	keypath = append(keypath, filepath.Join(getenv("HOME"), ".ecfg", "keys"))
	return
}

// SystemKeypath returns the default system-wide locations at which to search
// for ecfg keys. /opt/ejson/keys is provided for backwards-compatibility with
// ejson.
func SystemKeypath() (keypath []string) {
	keypath = append(keypath, "/etc/ecfg/keys")
	keypath = append(keypath, "/opt/ejson/keys")
	return
}

func findPrivateKey(pubkey [32]byte, keypath []string) (privkey [32]byte, err error) {
	keyString := os.Getenv("ECFG_PRIVATE_KEY")
	if keyString == "" {
		for _, keydir := range keypath {
			keyFile := fmt.Sprintf("%s/%x", keydir, pubkey)
			fileContents, err := readFile(keyFile)
			if err == nil {
				keyString = strings.TrimSpace(string(fileContents))
				break
			}
		}
	}
	if keyString == "" {
		err = fmt.Errorf("private key not found in keypath")
		return
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
	getuid    = os.Getuid
	getenv    = os.Getenv
)
