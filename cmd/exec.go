package main

import (
	"fmt"
	"strings"
)

type ExecCmd struct {
	Name   string `arg required help:"AWS Role alias name"`
	Region string `optional short:"r" help:"AWS Region" env:"AWS_DEFAULT_REGION"`

	Cmd  string   `arg required name:"command" help:"Command to execute"`
	Args []string `arg optional name:"args" help:"Associated arguments for the command"`
}

func (e *ExecCmd) Run(ctx *RunContext) error {
	fmt.Printf("exec: %s %s\n", e.Cmd, strings.Join(e.Args, " "))
	return nil
}
