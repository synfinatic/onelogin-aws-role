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

/*
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
func Login(ctx *RunContext, profile string) (aws.STSSession, error) {
	cli := *ctx.Cli
	cfile, err := LoadConfigFile(GetPath(cli.ConfigFile))
	if err != nil {
		return aws.STSSession{}, fmt.Errorf("Unable to open %s: %s", cli.ConfigFile, err.Error())
	}

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
*/
