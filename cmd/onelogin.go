package main

import (
	"github.com/alecthomas/kong"
	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	slt "github.com/onelogin/onelogin-go-sdk/pkg/services/session_login_tokens"
)

// Succeeds or dies
func NewOneLoginAPIClient(ctx *kong.Context, cli CLI) (*client.APIClient, error) {
	c, err := client.NewClient(&client.APIClientConfig{
		Timeout:      60,
		ClientID:     cli.ClientID,
		ClientSecret: cli.ClientSecret,
		Region:       cli.OLRegion,
	})
	return c, err
}

func OneLoginSessionLoginToken(c *client.APIClient, cli CLI) (*slt.SessionLoginToken, error) {
	request := &slt.SessionLoginTokenRequest{
		UsernameOrEmail: oltypes.String(cli.Username),
		Password:        oltypes.String(cli.Password),
		Subdomain:       oltypes.String(cli.Subdomain),
	}
	session_token, err := c.Services.SessionLoginTokensV1.Create(request)
	if err != nil {
		return nil, err
	}
	return session_token, nil
}
