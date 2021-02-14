package main

import (
	"fmt"

	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

type CheckCmd struct {
	// No sub options
}

func (cc *CheckCmd) Run(ctx *RunContext) error {
	c := onelogin.LoadOneLoginCache("")
	expired := c.AccessToken.IsExpired()
	if expired {
		fmt.Println("OneLogin OAuth2 token has expired.")
	} else {
		fmt.Printf("OneLogin OAuth2 token will expire at: %s\n", c.AccessToken.ExpiresAt())
	}

	return nil
}
