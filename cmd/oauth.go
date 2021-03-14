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
	"fmt"
	"reflect"

	"github.com/Songmu/prompter"
	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/onelogin-aws-role/utils"
)

type OauthCmd struct {
	Show OauthShowCmd `kong:"cmd,help='Show Oauth UserId/Secret',default='1'"`
	Set  OauthSetCmd  `kong:"cmd,help='Set Oauth UserId/Secret'"`
}

type OauthSetCmd struct{}
type OauthShowCmd struct{}

type OauthConfig struct {
	ClientId string `json:"clientid" header:"OneLogin ClientId"`
	Secret   string `json:"secret" header:"OneLogin Secret"`
}

func (oc *OauthShowCmd) Run(ctx *RunContext) error {
	kr, err := OpenKeyring(nil)
	if err != nil {
		return fmt.Errorf("Unable to get OauthConfig: %s", err)
	}

	oauth := &OauthConfig{}
	err = kr.GetOauthConfig(oauth)
	if err != nil {
		return fmt.Errorf("Unable to get OauthConfig: %s", err)
	}

	ts := []utils.TableStruct{}
	ts = append(ts, *oauth)
	fields := []string{"ClientId", "Secret"}
	utils.GenerateTable(ts, fields)

	return nil
}

func (oc *OauthSetCmd) Run(ctx *RunContext) error {
	kr, err := OpenKeyring(nil)
	if err != nil {
		return err
	}

	clientid := ""
	for len(clientid) != 64 {
		clientid = prompter.Prompt("OneLogin ClientId", "")

		if len(clientid) != 64 {
			log.Error("Invalid OneLogin ClientId: Must be 64 characters long")
		}
	}

	secret := ""
	for len(secret) != 64 {
		secret = prompter.Password("OneLogin Secret")

		if len(secret) != 64 {
			log.Error("Invalid OneLogin Secret: Must be 64 characters long")
		}
	}

	o := OauthConfig{
		ClientId: clientid,
		Secret:   secret,
	}
	return kr.SaveOauthConfig(o)
}

// Necessary for util.GenerateTable
func (oc OauthConfig) GetHeader(fieldName string) (string, error) {
	v := reflect.ValueOf(oc)
	return utils.GetHeaderTag(v, fieldName)
}
