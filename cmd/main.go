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
	"fmt"

	"github.com/Songmu/prompter"
	"github.com/alecthomas/kong"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/onelogin-aws-role/aws"
	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

// These variables are defined in the Makefile
var Version = "unknown"
var Buildinfos = "unknown"
var Tag = "NO-TAG"
var CommitID = "unknown"

type RunContext struct {
	OneLogin *onelogin.OneLogin
	Kctx     *kong.Context
	Cli      *CLI
	Config   *ConfigFile
}

type CLI struct {
	// Common Arguments
	LogLevel string `kong:"optional,short='l',name='loglevel',default='info',enum='error,warn,info,debug',help='Logging level [error|warn|info|debug]'"`
	// have to hard code CONFIG_YAML value here because no way to do string interpolation in a strcture tag.
	ConfigFile string `kong:"optional,short='c',name='config',default='~/.onelogin-aws-role.yaml',help='Config file'"`
	// AWS Params
	Region   string `kong:"optional,short='r',help='AWS Region',env='AWS_DEFAULT_REGION'"`
	Duration int64  `kong:"optional,short='d',help='AWS Session duration in minutes (default: 1hr)',default=60"`

	// Commands
	//	Role RoleCmd `kong:"cmd,help='Fetch & cache AWS STS Token for a given Role/Profile'"`
	//	App   AppCmd   `kong:"cmd,help='Fetch & cache all AWS STS Tokens for a given OneLogin AppID'"`
	Exec  ExecCmd  `kong:"cmd,help='Execute command using specified AWS Role/Profile.'"`
	List  ListCmd  `kong:"cmd,help='List all role / appid aliases (default command)',default='1'"`
	Oauth OauthCmd `kong:"cmd,help='Manage OneLogin Oauth credentials'"`
	// Revoke -- much later
	Version VersionCmd `kong:"cmd,help='Print version and exit'"`
}

func parse_args(cli *CLI) *kong.Context {
	op := kong.Description("Utility to manage temporary AWS API Credentials issued via OneLogin")
	ctx := kong.Parse(cli, op)

	switch cli.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}

	return ctx
}

func main() {
	cli := CLI{}
	ctx := parse_args(&cli)

	c, err := LoadConfigFile(GetPath(cli.ConfigFile))
	if err != nil {
		log.Fatalf("Unable to load config: %s", err.Error())
	}
	// c.MergeCLI(&cli)

	run_ctx := RunContext{
		OneLogin: nil,
		Kctx:     ctx,
		Cli:      &cli,
		Config:   c,
	}
	err = ctx.Run(&run_ctx)
	if err != nil {
		log.Fatalf("Error running command: %s", err.Error())
	}
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
	kr, err := OpenKeyring(nil)
	if err != nil {
		return aws.STSSession{}, fmt.Errorf("Unable to open KeyChain for OneLogin Oauth: %s", err)
	}
	oauth := OauthConfig{}
	err = kr.GetOauthConfig(&oauth)
	if err != nil {
		return aws.STSSession{}, fmt.Errorf("Please configure Oauth credentials")
	}

	o, err := onelogin.NewOneLogin(oauth.ClientId, oauth.Secret, ctx.Config.Region)
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
		log.Debugf("Got SAML Assertion:\n%s", assertion)
	}

	role, err := ctx.Config.GetRoleArn(profile)
	if err != nil {
		log.Fatal(err)
	}

	var region string
	if cli.Region != "" {
		region = cli.Region
	} else {
		region, err = ctx.Config.GetRoleRegion(profile)
		if err != nil {
			log.WithError(err).Warn("Unable to set default AWS region, falling back to us-east-1")
			region = "us-east-1"
		}
	}

	return aws.GetSTSSession(assertion, role, region, cli.Duration*60)
}

type VersionCmd struct {
}

func (cc *VersionCmd) Run(ctx *RunContext) error {
	fmt.Printf("OneLogin AWS Role Version %s -- Copyright 2021 Aaron Turner\n", Version)
	fmt.Printf("%s (%s) built at %s\n", CommitID, Tag, Buildinfos)
	return nil
}
