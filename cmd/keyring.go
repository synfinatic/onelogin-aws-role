package main

import (
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"golang.org/x/crypto/ssh/terminal"
)

var krConfigDefaults = keyring.Config{
	ServiceName:              "OneLoginAWSRole",
	FileDir:                  "~/.onelogin-aws-role/keys/",
	FilePasswordFunc:         fileKeyringPassphrasePrompt,
	LibSecretCollectionName:  "oneloginawsrole",
	KWalletAppID:             "onelogin-aws-role",
	KWalletFolder:            "onelogin-aws-role",
	KeychainTrustApplication: true,
	WinCredPrefix:            "onelogin-aws-role",
}

func fileKeyringPassphrasePrompt(prompt string) (string, error) {
	if password := os.Getenv("ONELOGIN_AWS_ROLE_FILE_PASSPHRASE"); password != "" {
		return password, nil
	}

	fmt.Fprintf(os.Stderr, "%s: ", prompt)
	b, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(b), nil
}
