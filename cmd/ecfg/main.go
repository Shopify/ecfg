package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/Shopify/ecfg"
	"github.com/urfave/cli"
)

var version string // set by Makefile via LDFLAGS

func execManpage(sec, page string) {
	if err := syscall.Exec("/usr/bin/env", []string{"/usr/bin/env", "man", sec, page}, os.Environ()); err != nil {
		fmt.Println("Exec error:", err)
	}
	os.Exit(1)
}

func main() {
	// Encryption is expensive. We'd rather burn cycles on many cores than wait.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Rather than using the built-in help printer, display the bundled manpages.
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		if cmd, ok := data.(cli.Command); ok {
			switch cmd.Name {
			case "encrypt", "decrypt", "keygen":
				execManpage("1", "ecfg-"+cmd.Name)
			}
		}
		execManpage("1", "ecfg")
	}

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "keydir, k",
			Value:  "/opt/ecfg/keys",
			Usage:  "Directory containing ecfg keys",
			EnvVar: "ECFG_KEYDIR",
		},
	}
	app.Usage = "manage encrypted secrets using public key encryption"
	app.Version = version
	app.Author = "Burke Libbey"
	app.Email = "burke.libbey@shopify.com"
	app.Commands = []cli.Command{
		{
			Name:      "encrypt",
			ShortName: "e",
			Usage:     "(re-)encrypt one or more ecfg files",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type, t",
					Usage: "Specify the filetype (json, yaml, or toml)",
				},
			},
			Action: func(c *cli.Context) error {
				args := c.Args()
				if len(args) > 1 {
					return errors.New("ecfg decrypt only operates on one file at a time")
				}
				firstArg := ""
				if len(args) == 1 {
					firstArg = args[0]
				}
				fileType, err := determineFileType(c.String("t"), firstArg)
				if err != nil {
					return err
				}
				return encryptAction(firstArg, fileType)
			},
		},
		{
			Name:      "decrypt",
			ShortName: "d",
			Usage:     "decrypt an ecfg file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "o",
					Usage: "print output to the provided file, rather than stdout",
				},
				cli.StringFlag{
					Name:  "type, t",
					Usage: "Specify the filetype (json, yaml, or toml)",
				},
			},
			Action: func(c *cli.Context) error {
				args := c.Args()
				if len(args) > 1 {
					return errors.New("ecfg decrypt only operates on one file at a time")
				}
				firstArg := ""
				if len(args) == 1 {
					firstArg = args[0]
				}
				fileType, err := determineFileType(c.String("t"), firstArg)
				if err != nil {
					return err
				}
				return decryptAction(firstArg, c.GlobalString("keydir"), c.String("o"), fileType)
			},
		},
		{
			Name:      "keygen",
			ShortName: "g",
			Usage:     "generate a new ecfg keypair",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "write, w",
					Usage: "rather than printing both keys, print the public and write the private into the keydir",
				},
			},
			Action: func(c *cli.Context) error {
				return keygenAction(c.Args(), c.GlobalString("keydir"), c.Bool("write"))
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func determineFileType(typeArg string, firstArg string) (ecfg.FileType, error) {
	switch typeArg {
	case "json":
		return ecfg.FileTypeJSON, nil
	case "yaml":
		return ecfg.FileTypeYAML, nil
	case "toml":
		return ecfg.FileTypeTOML, nil
	case "":
		if firstArg == "" {
			return ecfg.FileTypeJSON, errors.New("--type must be passed when not inferrable from file name")
		}
		if strings.HasSuffix(firstArg, "json") {
			return ecfg.FileTypeJSON, nil
		}
		if strings.HasSuffix(firstArg, "yaml") || strings.HasSuffix(firstArg, "yml") {
			return ecfg.FileTypeYAML, nil
		}
		if strings.HasSuffix(firstArg, "toml") {
			return ecfg.FileTypeTOML, nil
		}
		return ecfg.FileTypeJSON, errors.New("can't infer filetype from filename. rename file or specify type with --type")
	default:
		return ecfg.FileTypeJSON, errors.New("invalid filetype: specify 'json', 'yaml', or 'toml'")
	}
	if firstArg == "" && typeArg == "" {
		return ecfg.FileTypeJSON, errors.New("--type must be passed when not inferrable from file name")
	}
	return ecfg.FileTypeJSON, nil
}
