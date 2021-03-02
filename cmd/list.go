package main

import (
	"fmt"
	"strings"
)

type ListCmd struct {
	// no args
}

func (cc *ListCmd) Run(ctx *RunContext) error {
	cfile, err := LoadConfigFile(GetPath(CONFIG_YAML))
	if err != nil {
		return fmt.Errorf("Unable to open %s: %s", CONFIG_YAML, err.Error())
	}
	max_alias := len("Role Alias")
	max_arn := len("ARN")
	roles := cfile.GetRoles()
	for alias, arn := range *roles {
		if len(alias) > max_alias {
			max_alias = len(alias)
		}
		if len(arn) > max_arn {
			max_arn = len(arn)
		}
	}

	fstring := fmt.Sprintf("%%-%ds | %%-%ds\n", max_alias, max_arn)
	header := fmt.Sprintf(fstring, "Role Alias", "ARN")
	fmt.Printf(header)
	fmt.Printf("%s\n", strings.Repeat("=", len(header)-1))

	for alias, arn := range *roles {
		fmt.Printf(fstring, alias, arn)
	}

	fmt.Printf("\n\n")

	apps := cfile.GetApps()
	max_name := len("App Name")
	max_id := len("AppID")
	for name, id := range *apps {
		if len(name) > max_name {
			max_name = len(name)
		}
		if len(fmt.Sprintf("%d", id)) > max_id {
			max_id = len(fmt.Sprintf("%d", id))
		}

	}

	fstring = fmt.Sprintf("%%-%ds | %%-%ds\n", max_name, max_id)
	header = fmt.Sprintf(fstring, "App Name", "AppID")
	fmt.Printf(header)
	fmt.Printf("%s\n", strings.Repeat("=", len(header)-1))

	for name, id := range *apps {
		fmt.Printf(fstring, name, fmt.Sprintf("%d", id))
	}
	return nil
}
