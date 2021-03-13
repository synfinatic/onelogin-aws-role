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
	log "github.com/sirupsen/logrus"

	"github.com/alecthomas/kong"
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
	ConfigFile string `kong:"optional,short='c',name='config',default='~/.onelogin.yaml',help='Config file'"`
	// AWS Params
	Region   string `kong:"optional,short='r',help='AWS Region',env='AWS_DEFAULT_REGION'"`
	Duration int64  `kong:"optional,short='d',help='AWS Session duration in minutes (default: 1hr)',default=60"`

	// Commands
	Role RoleCmd `kong:"cmd,help='Fetch & cache AWS STS Token for a given Role/Profile'"`
	App  AppCmd  `kong:"cmd,help='Fetch & cache all AWS STS Tokens for a given OneLogin AppID'"`
	Exec ExecCmd `kong:"cmd,help='Execute command using specified AWS Role/Profile.'"`
	List ListCmd `kong:"cmd,help='List all role / appid aliases (default command)',default='1'"`
	// Revoke -- much later
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

	c, err := LoadConfigFile(cli.ConfigFile)
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
