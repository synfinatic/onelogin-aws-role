package main

import (
	"fmt"

	"github.com/Songmu/prompter"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/onelogin-aws-role/aws"
	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

type RoleCmd struct {
	Profile string `kong:"arg,required,name='profile',help='AWS Role alias name'"`
}

func (cc *RoleCmd) Run(ctx *RunContext) error {
	cli := *ctx.Cli
	_, err := GetSession(ctx, cli.Role.Profile)
	return err
}

func GetSession(ctx *RunContext, profile string) (aws.STSSession, error) {
	session := aws.STSSession{}
	kr, err := OpenKeyring(nil)
	if err != nil {
		log.WithError(err).Warn("Unable to retrieve STS Session from Keychain")
	} else {
		err = kr.GetSTSSession(profile, &session)
		if err != nil {
			log.WithError(err).Warn("Unable to read STS Session from Keychain")
		}
		if session.Expired() {
			log.Warn("Cached STS SessionToken has expired")
		}
	}

	if session.Expired() {
		session, err = Login(ctx, profile)
		if err != nil {
			log.WithError(err).Fatal("Unable to get STSSession")
		}
		err = kr.SaveSTSSession(profile, session)
		if err != nil {
			log.WithError(err).Warn("Unable to cache STS Session in Keychain")
		}
	}
	return session, nil
}

func Login(ctx *RunContext, profile string) (aws.STSSession, error) {
	cli := *ctx.Cli
	cfile, err := LoadConfigFile(GetPath(cli.ConfigFile))
	if err != nil {
		return aws.STSSession{}, fmt.Errorf("Unable to open %s: %s", cli.ConfigFile, err.Error())
	}

	o, err := onelogin.NewOneLogin(ctx.Config.ClientID, ctx.Config.ClientSecret, ctx.Config.Region)
	if err != nil {
		log.WithError(err).Fatal("Unable to connect to OneLogin")
	}
	log.Debugf("config = %s", spew.Sdump(ctx.Config))
	appid, err := ctx.Config.GetAppIdForRole(profile)
	if err != nil {
		return aws.STSSession{}, err
	}

	ols := &onelogin.OneLoginSAML{}
	need_mfa := false
	passwd_auth_pass := false
	for !passwd_auth_pass {
		passwd := prompter.Password("Enter your OneLogin password")

		if passwd == "" {
			return aws.STSSession{}, fmt.Errorf("OneLogin authentication aborted")
		}
		ols = onelogin.NewOneLoginSAML(o)
		need_mfa, err = ols.GetAssertion(ctx.Config.Username, passwd, ctx.Config.Subdomain, appid, "")
		if err == nil {
			passwd_auth_pass = true
		}
	}

	if need_mfa {
		log.Info("MFA Required")
		success, err := ols.SubmitMFA(ctx.Config.Mfa, appid)
		if err != nil {
			log.Fatal(err)
		}
		if !success {
			log.Fatalf("MFA auth failed.")
		}
	}
	assertion, err := ols.OneLogin.Cache.GetAssertion(appid)
	if err != nil {
		log.Fatalf("Unable to get SAML Assertion: %s", err.Error())
	} else {
		log.Infof("Got SAML Assertion:\n%s", assertion)
	}

	role, err := cfile.GetRoleArn(profile)
	if err != nil {
		log.Fatal(err)
	}

	var region string
	if cli.Region != "" {
		region = cli.Region
	} else {
		region, err = cfile.GetRoleRegion(profile)
		if err != nil {
			log.WithError(err).Warn("Unable to set default AWS region, falling back to us-east-1")
			region = "us-east-1"
		}
	}

	return aws.GetSTSSession(assertion, role, region, cli.Duration*60)
}
