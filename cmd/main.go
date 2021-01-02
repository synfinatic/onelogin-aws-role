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

	"github.com/alecthomas/kong"

	slt "github.com/onelogin/onelogin-go-sdk/pkg/services/session_login_tokens"
	log "github.com/sirupsen/logrus"
)

// These variables are defined in the Makefile
var Version = "unknown"
var Buildinfos = "unknown"
var Tag = "NO-TAG"
var CommitID = "unknown"

type RunContext struct {
	SessionLoginToken *slt.SessionLoginToken
	Kctx              *kong.Context
}

type CLI struct {
	LogLevel string `name:"loglevel" default:"warn" enum:"error,warn,debug" help:"Logging level"`
	// AWS params
	Account  uint64 `optional help:"AWS Account ID" env:"AWS_ACCOUNT_ID"`
	Region   string `optional help:"AWS Region" env:"AWS_REGION"`
	Role     string `optional help:"AWS Role to assume"`
	Duration int    `optional help:"AWS credential duration (minutes)" default:60`

	// OneLogin params
	ClientID     string `optional help:"OneLogin ClientID" env:"OL_CLIENT_ID"`
	ClientSecret string `optional help:"OneLogin Client Secret" env:"OL_CLIENT_SECRET"`
	AppId        int32  `optional help:"OneLogin App ID" env:"OL_APP_ID"`
	Subdomain    string `optional help:"OneLogin Subdomain" env:"OL_SUBDOMAIN"`
	Email        string `optional help:"OneLogin login email" env:"OL_EMAIL"`
	Password     string `optional help:"OneLogin password" env:"OL_PASSWORD"`
	OLRegion     string `optional help:"OneLogin region" default:"us" enum:"us,eu" env:"OL_REGION"`

	// Commands
	Exec ExecCmd `cmd help:"Execute command using specified AWS Role." default:"1"`
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

func main() {
	cli := CLI{}
	_ = parse_args(&cli)

	c, err := LoadConfig()
	if err != nil {
		log.Fatalf("Unable to load config: %s", err.Error())
	}
	c.MergeCLI(&cli)
	log.Debugf("CLI Config: %v", cli)

	alias, err := c.GetAccountAlias(cli.Account)
	if err != nil {
		log.Warnf("Unable to lookup %d: %s", cli.Account, err.Error())
	} else {
		fmt.Printf("Account: %s [%d]\n", alias, cli.Account)
	}

	olc, err := NewOneLoginAPIClient(cli)
	if err != nil {
		log.Fatal(err)
	}
	/*
		app, err := OneLoginGetApp(olc, cli.AppId)
		if err != nil {
			log.Fatalf("Unable to get App: %s", err.Error())
		}
		log.Debugf("App: %v", app)
	*/

	token, err := OneLoginSessionLoginToken(olc, cli)
	if err != nil {
		log.Fatalf("Unable to get SessionLoginToken: %s", err.Error())
	}

	log.Debugf("token: %v\n", token)
	/*
		cli_args := RunContext{
			// variables not part of ExecCmd go here
			SessionLoginToken: resp.Value,
		}
	*/

	//	err := ctx.Run(&cli)
}
