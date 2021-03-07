package main

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
 * This command just does a pretty print of our YAML config file basically
 */

type ListCmd struct {
	Fields     []string `kong:"optional,arg,enum='AccountId,AccountName,AppId,AppName,AppAlias,Arn,RoleAlias,Region',help='Fields to display (default: AppAlias AccountName RoleAlias Arn)'"`
	ListFields bool     `kong:"optional,short='f',help='List available fields'"`
}

// Fields match those in FlatConfig
var defaultFields = []string{
	"AppAlias",
	"AccountName",
	"RoleAlias",
	"Arn",
}

const (
	ALIAS       = "Role Alias"
	ARN         = "ARN"
	APPNAME     = "App Alias"
	APP         = "AppID"
	ACCOUNT     = "AccountID"
	ACCOUNTNAME = "Account Alias"
)

func (cc *ListCmd) Run(ctx *RunContext) error {
	cli := *ctx.Cli

	cfile, err := LoadConfigFile(GetPath(cli.ConfigFile))
	if err != nil {
		return fmt.Errorf("Unable to open %s: %s", cli.ConfigFile, err.Error())
	}

	fc := cfile.GetFlatConfig()
	if cli.List.ListFields {
		listFlatConfigFields(fc[0])
		os.Exit(1)
	}

	// accounts
	accounts := map[string]string{}
	for k, v := range *cfile.Accounts {
		accounts[fmt.Sprintf("%d", k)] = v
	}
	report(ACCOUNT, ACCOUNTNAME, accounts)

	fmt.Printf("\n\n")

	if len(cli.List.Fields) > 0 {
		newReport(fc, cli.List.Fields)
	} else if cfile.Fields != nil && len(*cfile.Fields) > 0 {
		newReport(fc, *cfile.Fields)
	} else {
		newReport(fc, defaultFields)
	}
	return nil
}

func listFlatConfigFields(fc FlatConfig) {
	fields := map[string]string{}
	t := reflect.TypeOf(fc)
	//	v := reflect.ValueOf(fc)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields[field.Name] = field.Tag.Get(FLAT_CONFIG_TAG)
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

func newReport(data []FlatConfig, fields []string) {
	headers := make([]interface{}, len(fields)) // must be interface
	headerLen := map[string]int{}

	// get length of selected headers
	for i, field := range fields {
		f, err := data[0].GetHeader(field)
		if err != nil {
			log.Fatal(err)
		}
		headers[i] = f
		headerLen[field] = len(f)
	}

	// get length of selected field values
	for _, row := range data {
		r := reflect.ValueOf(row)
		for _, field := range fields {
			val := r.FieldByName(field).String()
			if len(val) > headerLen[field] {
				headerLen[field] = len(val)
			}
		}
	}

	// build our fstring
	fstringList := []string{}
	for _, field := range fields {
		fstringList = append(fstringList, fmt.Sprintf("%%-%ds", headerLen[field]))
	}
	fstring := strings.Join(fstringList, " | ")
	fstring = fmt.Sprintf("%s\n", fstring)

	// print the header
	headerLine := fmt.Sprintf(fstring, headers...)
	fmt.Printf("%s%s\n", headerLine, strings.Repeat("=", len(headerLine)-1))

	// print each row
	for _, row := range data {
		r := make([]interface{}, len(fields))
		for i, field := range fields {
			f := reflect.ValueOf(row).FieldByName(field)
			// value is a string or Uint
			if f.Type().Name() == "string" {
				r[i] = f.String()
			} else {
				r[i] = fmt.Sprintf("%d", f.Uint())
			}
		}
		fmt.Printf(fstring, r...)
	}
}

func report(key, value string, values map[string]string) {
	// calculate max len of key & value columns
	max_key := len(key)
	max_value := len(value)
	for k, v := range values {
		if len(k) > max_key {
			max_key = len(k)
		}
		if len(v) > max_value {
			max_value = len(v)
		}
	}

	// generate header
	fstring := fmt.Sprintf("%%-%ds | %%-%ds\n", max_key, max_value)
	header := fmt.Sprintf(fstring, key, value)
	fmt.Printf(header)
	fmt.Printf("%s\n", strings.Repeat("=", len(header)-1))

	// generate report values
	for k, v := range values {
		fmt.Printf(fstring, k, v)
	}
}
