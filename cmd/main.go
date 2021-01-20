package main

/*
 *  OneLogin Go AWS Assume Role
 *  Copyright (C) 2020  Aaron Turner
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/99designs/keyring"
	"github.com/alecthomas/kong"
	"github.com/synfinatic/onelogin-aws-role/onelogin"
	"golang.org/x/crypto/ssh/terminal"
)

// These variables are defined in the Makefile
var Version = "unknown"
var Buildinfos = "unknown"
var Tag = "NO-TAG"
var CommitID = "unknown"

type RunContext struct {
	OneLogin *onelogin.OneLogin
	Kctx     *kong.Context
}

type CLI struct {
	LogLevel string `optional short:"l" name:"loglevel" default:"warn" enum:"error,warn,debug" help:"Logging level"`
	// AWS params
	Account  uint64 `optional short:"a" help:"AWS Account ID" env:"AWS_ACCOUNT_ID"`
	Region   string `optional short:"r" help:"AWS Region" env:"AWS_REGION"`
	Role     string `optional short:"R" help:"AWS Role to assume"`
	Duration int    `optional short:"D" default:"60" help:"AWS credential duration (minutes)"`

	// OneLogin params
	ClientID     string `optional help:"OneLogin ClientID" env:"OL_CLIENT_ID"`
	ClientSecret string `optional help:"OneLogin Client Secret" env:"OL_CLIENT_SECRET"`
	AppId        uint32 `optional short:"i" help:"OneLogin App ID" env:"OL_APP_ID"`
	Subdomain    string `optional help:"OneLogin Subdomain" env:"OL_SUBDOMAIN"`
	Email        string `optional short:"e" help:"OneLogin login email" env:"OL_EMAIL"`
	Password     string `optional hidden help:"OneLogin password" env:"OL_PASSWORD"` // FIXME to be a pure ENV lookup
	OLRegion     string `optional help:"OneLogin region" default:"us" enum:"us,eu" env:"OL_REGION"`
	MfaType      string `optional short:"m" help:"OneLogin MFA name" env:"OL_MFA"`
	MfaPush      bool   `optional short:"p" help:"Use MFA Push with OneLogin Protect" env:"OL_MFA_PUSH"`
	Mfa          int32  `optional short:"c" help:"MFA Code"`

	// Commands
	Exec ExecCmd `cmd default:"1" help:"Execute command using specified AWS Role."`
	//	Metadata MetadataCmd `cmd help:"Start metadata service."`

}

func parse_args(cli *CLI) *kong.Context {
	ctx := kong.Parse(cli)

	switch cli.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}

	return ctx
}

var krConfigDefaults = keyring.Config{
	ServiceName:              "OneLoginAWSRole",
	FileDir:                  "~/.oneloginawsrole/keys/",
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

func main() {
	cli := CLI{}
	ctx := parse_args(&cli)
	log.Errorf("ctx: %v", ctx)
	log.Errorf("cli: %v", cli)

	c, err := LoadConfig()
	if err != nil {
		log.Fatalf("Unable to load config: %s", err.Error())
	}
	c.MergeCLI(&cli)

	kr, err := keyring.Open(krConfigDefaults)
	if err != nil {
		log.Fatalf("Unable to open key store: %s", err.Error())
	}

	o := onelogin.NewOneLogin(cli.ClientID, cli.ClientSecret, cli.OLRegion)
	/*
		_, err = o.GetRateLimit()
		if err != nil {
			log.Fatalf("%s", err.Error())
		} else {
			os.Exit(3)
		}
	*/
	ols := onelogin.NewOneLoginSAML(o, &kr)
	need_mfa, err := ols.GetAssertion(cli.Email, cli.Password, cli.Subdomain, cli.AppId, "")
	if err != nil {
		log.Fatalf("GetAssertion: %s", err.Error())
	}
	if need_mfa {
		log.Debug("Need MFA")
		ok, err := ols.OneLoginProtectPush(cli.AppId, 10)
		if err != nil {
			log.Fatalf("Error doing ProtectPush: %s", err.Error())
		}
		if !ok {
			log.Fatalf("MFA push failed/timed out")
		}
	}
	assertion, err := ols.OneLogin.Cache.GetAssertion(cli.AppId)
	if err != nil {
		log.Fatalf("Unable to get SAML Assertion: %s", err.Error())
	} else {
		log.Infof("Got SAML Assertion:\n%s", assertion)
	}
	roles, err := ols.OneLogin.Cache.GetRoles(cli.AppId)
	if err != nil {
		log.Errorf("Unable to get roles: %s", err.Error())
	}
	fmt.Printf("Roles: %v", roles)
}
