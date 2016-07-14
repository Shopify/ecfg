package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Shopify/ecfg"
)

func encryptAction(filePath string, ftype ecfg.FileType) error {
	if filePath == "" { // read from stdin, write to stdout
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		out, err := ecfg.EncryptData(data, ftype)
		if err != nil {
			return err
		}
		fmt.Printf("%s", string(out))
		return nil
	}
	n, err := ecfg.EncryptFileInPlace(filePath, ecfg.FileTypeJSON)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d bytes to %s.\n", n, filePath)
	return nil
}

func decryptAction(filePath string, keydir, outFile string, ftype ecfg.FileType) error {
	if filePath == "" { // read from stdin, write to stdout
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		out, err := ecfg.DecryptData(data, keydir, ftype)
		if err != nil {
			return err
		}
		fmt.Printf("%s", string(out))
		return nil
	}

	decrypted, err := ecfg.DecryptFile(filePath, keydir, ecfg.FileTypeJSON)
	if err != nil {
		return err
	}

	target := os.Stdout
	if outFile != "" {
		target, err = os.Create(outFile)
		if err != nil {
			return err
		}
		defer func() { _ = target.Close() }()
	}

	_, err = target.Write(decrypted)
	return err
}

func keygenAction(args []string, keydir string, wFlag bool) error {
	pub, priv, err := ecfg.GenerateKeypair()
	if err != nil {
		return err
	}

	if wFlag {
		keyFile := fmt.Sprintf("%s/%s", keydir, pub)
		err := writeFile(keyFile, []byte(priv), 0440)
		if err != nil {
			return err
		}
		fmt.Println(pub)
	} else {
		fmt.Printf("Public Key:\n%s\nPrivate Key:\n%s\n", pub, priv)
	}
	return nil
}

// for mocking in tests
var (
	writeFile = ioutil.WriteFile
)
