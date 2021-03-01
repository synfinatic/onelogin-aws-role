package main

import (
	"fmt"

	"github.com/99designs/keyring"
	"github.com/Songmu/prompter"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/onelogin-aws-role/aws"
	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

type RoleCmd struct {
	Name string `kong:"arg,required help:'AWS Role alias name'"`
	// AWS params
	Region   string `kong:"optional,short:'r',help:'AWS Region',env:'AWS_DEFAULT_REGION'"`
	Duration int    `kong:"optional,short:'D',default:'60',help:'AWS credential duration (minutes)'"`
}

func (cc *RoleCmd) Run(ctx *RunContext) error {
	cli := *ctx.Cli
	kr, err := keyring.Open(krConfigDefaults)
	if err != nil {
		log.Fatalf("Unable to open key store: %s", err.Error())
	}

	o, err := onelogin.NewOneLogin(ctx.Config.ClientID, ctx.Config.ClientSecret, ctx.Config.Region)
	if err != nil {
		log.WithError(err).Fatal("Unable to connect to OneLogin")
	}
	log.Debugf("config = %s", spew.Sdump(ctx.Config))
	appid, err := ctx.Config.GetAppIdForRole(cli.Role.Name)
	if err != nil {
		return err
	}
	passwd := prompter.Password("Enter your OneLogin password:")

	ols := onelogin.NewOneLoginSAML(o, &kr)
	need_mfa, err := ols.GetAssertion(ctx.Config.Username, passwd, ctx.Config.Subdomain, appid, "")
	if err != nil {
		return err
	}
	if need_mfa {
		log.Debug("Need MFA")
		ok, err := ols.OneLoginProtectPush(appid, 10)
		if err != nil {
			log.Fatalf("Error doing ProtectPush: %s", err.Error())
		}
		if !ok {
			log.Fatalf("MFA push failed/timed out")
		}
	}
	assertion, err := ols.OneLogin.Cache.GetAssertion(appid)
	if err != nil {
		log.Fatalf("Unable to get SAML Assertion: %s", err.Error())
	} else {
		log.Infof("Got SAML Assertion:\n%s", assertion)
	}
	roles, err := ols.OneLogin.Cache.GetRoles(appid)
	if err != nil {
		log.Errorf("Unable to get roles: %s", err.Error())
	}
	fmt.Printf("Roles: %v", roles)

	_, err = aws.GetSTSSession(assertion, cli.Role.Name, "us-east-1", 3600)
	if err != nil {
		log.WithError(err).Fatal("Unable to get STSSession")
	}
	return nil
}
