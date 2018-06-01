package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/Shopify/ecfg"
)

const access_W_OK = 0x02

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
	n, err := ecfg.EncryptFileInPlace(filePath, ftype)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d bytes to %s.\n", n, filePath)
	return nil
}

func decryptAction(filePath string, keydir, outFile string, ftype ecfg.FileType) error {
	keypath := ecfg.DefaultKeypath()
	if keydir != "" {
		keypath = []string{keydir}
	}

	if filePath == "" { // read from stdin, write to stdout
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		out, err := ecfg.DecryptData(data, keypath, ftype)
		if err != nil {
			return err
		}
		fmt.Printf("%s", string(out))
		return nil
	}

	decrypted, err := ecfg.DecryptFile(filePath, keypath, ftype)
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

	if !wFlag {
		fmt.Printf("Public Key:\n%s\nPrivate Key:\n%s\n", pub, priv)
		return nil
	}

	if keydir == "" {
		kp := ecfg.DefaultKeypath()
		for _, candidate := range kp {
			if syscall.Access(candidate, access_W_OK) == nil {
				keydir = candidate
				break
			}
		}
	}
	if keydir == "" {
		return fmt.Errorf(
			"ecfg keydir not writable. Set ECFG_KEYDIR or ensure directory exists and is writable: %s",
			ecfg.DefaultKeypath()[0])
	}

	keyFile := fmt.Sprintf("%s/%s", keydir, pub)
	err = writeFile(keyFile, []byte(priv), 0440)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote key to %s\n", keydir)
	fmt.Fprintf(os.Stdout, "%s\n", pub)
	return nil
}

// for mocking in tests
var (
	writeFile = ioutil.WriteFile
)
