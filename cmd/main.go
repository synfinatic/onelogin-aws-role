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
	Account  int    `name:"account" help:"AWS Account ID" env:"AWS_ACCOUNT_ID"`
	Region   string `name:"region" help:"AWS Region" env:"AWS_REGION"`
	Role     string `name:"role" help:"AWS Role to assume"`
	Duration int    `name:"duration" help:"AWS credential duration (minutes)" default:60`

	// OneLogin params
	ClientID     string `name:"client-id" help:"OneLogin ClientID" env:"OL_CLIENT_ID"`
	ClientSecret string `name:"client-secret" help:"OneLogin Client Secret" env:"OL_CLIENT_SECRET"`
	AppId        string `name:"app-id" help:"OneLogin App ID" env:"OL_APP_ID"`
	Subdomain    string `name:"subdomain" help:"OneLogin Subdomain" env:"OL_SUBDOMAIN"`
	Username     string `name:"username" help:"OneLogin username" env:"OL_USERNAME"`
	Password     string `name:"password" help:"OneLogin password" env:"OL_PASSWORD"`
	OLRegion     string `name:"ol-region" help:"OneLogin region" default:"us" enum:"us,eu" env:"OL_REGION"`

	// Commands
	Exec ExecCmd `cmd help:"Execute command using specified AWS Role."`
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
	ctx := parse_args(&cli)

	olc, err := NewOneLoginAPIClient(ctx, cli)
	if err != nil {
		log.Fatal(err)
	}
	token, err := OneLoginSessionLoginToken(olc, cli)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	fmt.Printf("token: %v", token)
	/*
		cli_args := RunContext{
			// variables not part of ExecCmd go here
			SessionLoginToken: resp.Value,
		}
	*/

	//	err := ctx.Run(&cli)
}
