package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Songmu/prompter"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/onelogin-aws-role/aws"
	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

type ExecCmd struct {
	Profile string `kong:"arg,required,name='profile',help='AWS Profile name to use'"`

	// AWS Params
	Region   string `kong:"optional,short='r',help='AWS Region',env='AWS_DEFAULT_REGION'"`
	Duration int64  `kong:"optional,short='d',help='AWS Session duration in minutes (default: 1hr)',default=60"`

	// Command
	Cmd  string   `kong:"arg,optional,name='command',help='Command to execute',env='SHELL'"`
	Args []string `kong:"arg,optional,name='args',help='Associated arguments for the command'"`
}

func (e *ExecCmd) Run(ctx *RunContext) error {
	cli := *ctx.Cli
	session := aws.STSSession{}

	kr, err := OpenKeyring(nil)
	if err != nil {
		log.WithError(err).Warn("Unable to retrieve STS Session from Keychain")
	} else {
		err = kr.GetSTSSession(cli.Exec.Profile, &session)
		if err != nil {
			log.WithError(err).Warn("Unable to read STS Session from Keychain")
		}
		if session.Expired() {
			log.Warn("Cached STS SessionToken has expired")
		}
	}

	if session.Expired() {
		session, err = Login(ctx)
		if err != nil {
			log.WithError(err).Fatal("Unable to get STSSession")
		}
		err = kr.SaveSTSSession(cli.Exec.Profile, session)
		if err != nil {
			log.WithError(err).Warn("Unable to cache STS Session in Keychain")
		}
	}

	// set our ENV & execute the command
	os.Setenv("AWS_ACCESS_KEY_ID", session.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", session.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", session.SessionToken)
	if cli.Exec.Region != "" {
		os.Setenv("AWS_DEFAULT_REGION", cli.Exec.Region)
	}
	os.Setenv("AWS_SESSION_EXPIRATION", session.Expiration.String())
	os.Setenv("AWS_ENABLED_PROFILE", cli.Exec.Profile)
	os.Setenv("AWS_ROLE_ARN", session.RoleARN)

	// ready our command and connect everything up
	cmd := exec.Command(cli.Exec.Cmd, cli.Exec.Args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	// just do it!
	return cmd.Run()
}

func Login(ctx *RunContext) (aws.STSSession, error) {
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
	appid, err := ctx.Config.GetAppIdForRole(cli.Exec.Profile)
	if err != nil {
		return aws.STSSession{}, err
	}
	passwd := prompter.Password("Enter your OneLogin password")

	ols := onelogin.NewOneLoginSAML(o)
	need_mfa, err := ols.GetAssertion(ctx.Config.Username, passwd, ctx.Config.Subdomain, appid, "")
	if err != nil {
		return aws.STSSession{}, err
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

	role, err := cfile.GetRoleArn(cli.Exec.Profile)
	if err != nil {
		log.Fatal(err)
	}

	var region string
	if cli.Exec.Region != "" {
		region = cli.Exec.Region
	} else {
		region, err = cfile.GetRoleRegion(cli.Exec.Profile)
		if err != nil {
			log.WithError(err).Warn("Unable to set default AWS region, falling back to us-east-1")
			region = "us-east-1"
		}
	}

	return aws.GetSTSSession(assertion, role, region, cli.Exec.Duration*60)
}
