package main

import (
	"fmt"
	"strings"
)

type ExecCmd struct {
	Cmd  string   `arg required name:"command" help:"Command to execute" type:"path"`
	Args []string `arg optional name:"args" help:"Associated arguments for the command"`
}

func (e *ExecCmd) Run(ctx *RunContext) error {
	fmt.Printf("exec: %s %s\n", e.Cmd, strings.Join(e.Args, " "))
	return nil
}
