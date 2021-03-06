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
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type ExecCmd struct {
	Profile string `kong:"arg,required,name='profile',help='AWS Profile name to use'"`

	// Command
	Cmd  string   `kong:"arg,optional,name='command',help='Command to execute',env='SHELL'"`
	Args []string `kong:"arg,optional,name='args',help='Associated arguments for the command'"`
}

func (e *ExecCmd) Run(ctx *RunContext) error {
	cli := *ctx.Cli
	session, err := GetSession(ctx, cli.Exec.Profile)
	if err != nil {
		log.Fatal(err)
	}

	// set our ENV & execute the command
	os.Setenv("AWS_ACCESS_KEY_ID", session.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", session.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", session.SessionToken)
	if cli.Region != "" {
		os.Setenv("AWS_DEFAULT_REGION", cli.Region)
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
