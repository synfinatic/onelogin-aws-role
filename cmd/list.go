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
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/synfinatic/onelogin-aws-role/utils"
	//	log "github.com/sirupsen/logrus"
)

/*
 * This command just does a pretty print of our YAML config file basically
 */

type ListCmd struct {
	Fields     []string `kong:"optional,arg,enum='AccountId,AccountName,AppId,AppName,AppAlias,Arn,Profile,Region',help='Fields to display (default: AppAlias AccountName RoleAlias Arn)'"`
	ListFields bool     `kong:"optional,short='f',help='List available fields'"`
}

// Fields match those in FlatConfig.  Used when user doesn't have the `fields` in
// their YAML config file or provided list on the CLI
var defaultFields = []string{
	"AppAlias",
	"AccountName",
	"Profile",
	"Arn",
}

func (cc *ListCmd) Run(ctx *RunContext) error {
	cli := *ctx.Cli

	cfile, err := LoadConfigFile(GetPath(cli.ConfigFile))
	if err != nil {
		return fmt.Errorf("Unable to open %s: %s", cli.ConfigFile, err.Error())
	}

	// If `-f` then print our fields and exit
	fc := cfile.GetFlatConfig()
	if cli.List.ListFields {
		listFlatConfigFields(fc[0])
		os.Exit(1)
	}

	// List our AWS account aliases by abusing the FlatConfig struct
	accounts := []FlatConfig{}
	for k, v := range *cfile.Accounts {
		accounts = append(accounts, FlatConfig{
			AccountId:   k,
			AccountName: v,
		})
	}
	accountList := []string{
		"AccountId",
		"AccountName",
	}
	// manually convert our []FlatConfig into []TableStruct because Go is lame
	ts := []utils.TableStruct{}
	for _, x := range accounts {
		ts = append(ts, x)
	}
	utils.GenerateTable(ts, accountList)

	fmt.Printf("\n\n")

	// manually convert our []FlatConfig into []TableStruct because Go is lame
	ts = []utils.TableStruct{}
	for _, x := range fc {
		ts = append(ts, x)
	}

	// List our configured Roles
	if len(cli.List.Fields) > 0 {
		utils.GenerateTable(ts, cli.List.Fields)
	} else if cfile.Fields != nil && len(*cfile.Fields) > 0 {
		utils.GenerateTable(ts, *cfile.Fields)
	} else {
		utils.GenerateTable(ts, defaultFields)
	}

	return nil
}

func listFlatConfigFields(fc FlatConfig) {
	fields := map[string]string{}
	t := reflect.TypeOf(fc)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields[field.Name] = field.Tag.Get(utils.TABLE_HEADER_TAG)
	}

	max_key := len("Field")
	max_val := len("Description")
	for k, v := range fields {
		if len(k) > max_key {
			max_key = len(k)
		}
		if len(v) > max_val {
			max_val = len(v)
		}
	}
	fstring := fmt.Sprintf("%%-%ds | %%-%ds\n", max_key, max_val)
	headerLine := fmt.Sprintf(fstring, "Field", "Description")
	fmt.Printf("%s%s\n", headerLine, strings.Repeat("=", len(headerLine)-1))

	// sort keys
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf(fstring, k, fields[k])
	}
}
