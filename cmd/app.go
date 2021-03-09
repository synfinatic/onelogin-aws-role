package main

import (
// 	log "github.com/sirupsen/logrus"
// 	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

type AppCmd struct {
	AppId string `kong:"arg,required,name='appid',help='OneLogin AppID alias or number'"`
}

func (ac *AppCmd) Run(ctx *RunContext) error {
	// _ := *ctx.Cli

	return nil
}
