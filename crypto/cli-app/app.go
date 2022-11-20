package main

import (
	"crypto-example/encryption"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {

	var key string
	var text string

	app := &cli.App{
		Name:    "crypto-cli",
		Usage:   "Command line interface for encryption and decryption",
		Version: "v1.0",
		Commands: []*cli.Command{
			{
				Name:    "encrypt",
				Aliases: []string{"e"},
				Usage:   "encrypt a text",
				Action: func(*cli.Context) error {
					cypherText, err := encryption.Encrypt(text, key)
					if err != nil {
						return cli.Exit("Input parameters are not valid for encryption, please provide valid key or text!", 86)
					}
					fmt.Println(cypherText)
					return nil
				},
			},
			{
				Name:    "decrypt",
				Aliases: []string{"d"},
				Usage:   "decrypt a text",
				Action: func(*cli.Context) error {
					plainText, err := encryption.Decrypt(text, key)
					if err != nil {
						return cli.Exit("Input parameters are not valid for decryption, please provide valid key or text!", 86)
					}
					fmt.Println(plainText)
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "key",
				Aliases:     []string{"k"},
				Usage:       "(Required)-Cypher key used for encryption/decryption, It's length should be a multiple of 16",
				Destination: &key,
			},
			&cli.StringFlag{
				Name:        "text",
				Aliases:     []string{"t"},
				Usage:       "(Required)-Text to encrypt/decrypt",
				Destination: &text,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
