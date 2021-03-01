package main

import (
	"fmt"
	"strings"
)

type ExecCmd struct {
	Name string `kong:"arg,required help:'AWS Role alias name'"`

	// AWS Params
	Region string `kong:"optional,short:'r',help:'AWS Region',env:'AWS_DEFAULT_REGION'"`

	// Command
	Cmd  string   `kong:"arg,required,name:'command',help:'Command to execute'"`
	Args []string `kong:"arg,optional,name:'args',help:'Associated arguments for the command'"`
}

func (e *ExecCmd) Run(ctx *RunContext) error {
	fmt.Printf("exec: %s %s\n", e.Cmd, strings.Join(e.Args, " "))
	return nil
}
