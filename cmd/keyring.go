package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"github.com/synfinatic/onelogin-aws-role/aws"
	"golang.org/x/crypto/ssh/terminal"
)

type KeyringCache struct {
	keyring keyring.Keyring
	config  keyring.Config
}

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

func OpenKeyring(cfg *keyring.Config) (*KeyringCache, error) {
	if cfg == nil {
		cfg = &krConfigDefaults
	}
	ring, err := keyring.Open(*cfg)
	if err != nil {
		return nil, err
	}
	kr := KeyringCache{
		keyring: ring,
		config:  *cfg,
	}
	return &kr, nil
}

// Save our STS Session in the key chain
func (kr *KeyringCache) SaveSTSSession(profile string, session aws.STSSession) error {
	jdata, err := json.Marshal(session)
	if err != nil {
		return err
	}
	err = kr.keyring.Set(keyring.Item{
		Key:  profile,
		Data: jdata,
	})
	return err
}

// Get our STS Session from the key chain
func (kr *KeyringCache) GetSTSSession(profile string, session *aws.STSSession) error {
	data, err := kr.keyring.Get(profile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data.Data, session)
	if err != nil {
		return err
	}
	return nil
}
