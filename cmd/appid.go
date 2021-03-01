package main

import (
// 	log "github.com/sirupsen/logrus"
// 	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

type AppIdCmd struct {
	Name string `arg required help:"OneLogin AppID alias or number"`
	// AWS Params
	Region   string `optional short:"r" help:"AWS Region" env:"AWS_DEFAULT_REGION"`
	Duration int    `optional short:"D" default:"60" help:"AWS credential duration (minutes)"`
}

func (ac *AppIdCmd) Run(ctx *RunContext) error {
	// _ := *ctx.Cli

	return nil
}
