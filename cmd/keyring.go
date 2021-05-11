package main

/*
 * OneLogin AWS Role
 * Copyright (c) 2020-2021 Aaron Turner  <aturner at synfin dot net>
 *
 * This program is free software: you can redistribute it
 * and/or modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or with the authors permission any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/99designs/keyring"
	"github.com/synfinatic/onelogin-aws-role/aws"
	"golang.org/x/crypto/ssh/terminal"
)

type KeyringCache struct {
	keyring keyring.Keyring
	config  keyring.Config
}

// https://github.com/99designs/keyring/blob/master/config.go
var krConfigDefaults = keyring.Config{
	ServiceName: "OneLoginAWSRole", // generic
	// OSX Keychain
	KeychainName:                   "OneLoginAWSRole",
	KeychainSynchronizable:         false,
	KeychainAccessibleWhenUnlocked: false,
	// KeychainPasswordFunc: ???,
	// Other systems below this line
	FileDir:                 "~/.onelogin-aws-role/keys/",
	FilePasswordFunc:        fileKeyringPassphrasePrompt,
	LibSecretCollectionName: "oneloginawsrole",
	KWalletAppID:            "onelogin-aws-role",
	KWalletFolder:           "onelogin-aws-role",
	WinCredPrefix:           "onelogin-aws-role",
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
		Key:  fmt.Sprintf("profile:%s", profile),
		Data: jdata,
	})
	return err
}

// Get our STS Session from the key chain
func (kr *KeyringCache) GetSTSSession(profile string, session *aws.STSSession) error {
	data, err := kr.keyring.Get(fmt.Sprintf("profile:%s", profile))
	if err != nil {
		return err
	}
	err = json.Unmarshal(data.Data, session)
	if err != nil {
		return err
	}
	return nil
}

func (kr *KeyringCache) GetOauthConfig(oauth *OauthConfig) error {
	data, err := kr.keyring.Get("oauth:config")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data.Data, oauth)
	if err != nil {
		return err
	}
	return nil
}

func (kr *KeyringCache) SaveOauthConfig(oauth OauthConfig) error {
	jdata, err := json.Marshal(oauth)
	if err != nil {
		return err
	}
	err = kr.keyring.Set(keyring.Item{
		Key:  "oauth:config",
		Data: jdata,
	})
	return err
}

func (kr *KeyringCache) RemoveSTSSession(profile string) error {
	keys, err := kr.keyring.Keys()
	if err != nil {
		return err
	}

	// make sure we have this profile stored
	hasKey := false
	key := fmt.Sprintf("profile:%s", profile)
	for _, k := range keys {
		if k == key {
			hasKey = true
			break
		}
	}
	if !hasKey {
		return fmt.Errorf("Unknown profile: %s", profile)
	}

	// Can't just call keyring.Remove() because it's broken, so we'll update the record instead
	session := aws.STSSession{}
	err = kr.GetSTSSession(profile, &session)
	if err != nil {
		return err
	}
	session.Expiration = time.Now()
	session.SessionToken = ""
	session.SecretAccessKey = ""
	return kr.SaveSTSSession(profile, session)
}
