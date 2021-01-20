package main

import (
	"fmt"
	"strings"
)

type ExecCmd struct {
	Cmd  string   `arg required name:"command" type:"path" help:"Command to execute"`
	Args []string `arg optional name:"args" help:"Associated arguments for the command"`
}

func (e *ExecCmd) Run(ctx *RunContext) error {
	fmt.Printf("exec: %s %s\n", e.Cmd, strings.Join(e.Args, " "))
	return nil
}
