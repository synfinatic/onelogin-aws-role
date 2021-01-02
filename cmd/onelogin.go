package main

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/apps"
	slt "github.com/onelogin/onelogin-go-sdk/pkg/services/session_login_tokens"
	log "github.com/sirupsen/logrus"
)

// Succeeds or dies
func NewOneLoginAPIClient(cli CLI) (*client.APIClient, error) {
	cc := client.APIClientConfig{
		Timeout:      60,
		ClientID:     cli.ClientID,
		ClientSecret: cli.ClientSecret,
		Region:       cli.OLRegion,
	}
	c, err := client.NewClient(&cc)
	return c, err
}

func OneLoginSessionLoginToken(c *client.APIClient, cli CLI) (*slt.SessionLoginToken, error) {
	request := slt.SessionLoginTokenRequest{
		UsernameOrEmail: oltypes.String(cli.Email),
		Password:        oltypes.String(cli.Password),
		Subdomain:       oltypes.String(cli.Subdomain),
	}
	session_token, err := c.Services.SessionLoginTokensV1.Create(&request)
	if err != nil {
		return nil, err
	}
	return session_token, nil
}

func OneLoginGetApp(c *client.APIClient, appid int32) (*apps.App, error) {
	log.Debugf("getting AppId: %d", appid)
	return c.Services.AppsV2.GetOne(appid)
}
