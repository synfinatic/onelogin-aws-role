package main

/*
 *

import (
	"fmt"

	"github.com/99designs/keyring"
	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

type LoginCmd struct {
	Name string `arg required help:"OneLogin App name"`

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
}

func (cc *LoginCmd) Run(ctx *RunContext) error {
	cli := *ctx.Cli
	kr, err := keyring.Open(krConfigDefaults)
	if err != nil {
		log.Fatalf("Unable to open key store: %s", err.Error())
	}

	o := onelogin.NewOneLogin(cli.ClientID, cli.ClientSecret, cli.OLRegion)

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
	return nil
}
*/
