package main

/*
 * Load/Retrieve AWS account alias
 *
 * OneLogin doesn't support converting an AWS Account ID into an alias
 * via any API, so we need to manage this ourselves :(
 */

import (
	"fmt"
	"os"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

type Config struct {
	ClientID     string
	ClientSecret string
	Subdomain    string
	Email        string
	Accounts     map[uint64]string
	Apps         map[uint64]yaml.MapItem
}

var ACCOUNTS_YAML string = "~/.onelogin.accounts.yaml"

func GetPath(path string) string {
	return strings.Replace(path, "~", os.Getenv("HOME"), 1)
}

func LoadConfig() (*Config, error) {
	path := GetPath(ACCOUNTS_YAML)
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to stat %s: %s", path, err.Error())
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to open %s: %s", path, err.Error())
	}

	defer file.Close()

	buf := make([]byte, info.Size())
	_, err = file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to read %s: %s", path, err.Error())
	}

	c := Config{}
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s: %s", path, err.Error())
	}

	return &c, nil
}

func (c *Config) GetAccountAlias(accountid uint64) (string, error) {
	//	account := fmt.Sprintf("%d", accountid)
	alias, exists := c.Accounts[accountid]
	if exists {
		return alias, nil
	}
	return "", fmt.Errorf("Account ID %v does not exist", accountid)
}

func (c *Config) GetApps() []uint64 {
	var ret []uint64 = []uint64{}

	for appid, _ := range c.Apps {
		ret = append(ret, appid)
	}
	return ret
}

func (c *Config) GetApp(appid uint64) yaml.MapItem {
	return c.Apps[appid]
}

// merges config values into the cli struct
func (c *Config) MergeCLI(cli *CLI) {
	if cli.ClientID == "" {
		cli.ClientID = c.ClientID
	}
	if cli.ClientSecret == "" {
		cli.ClientSecret = c.ClientSecret
	}
	if cli.Email == "" {
		cli.Email = c.Email
	}
	if cli.Subdomain == "" {
		cli.Subdomain = c.Subdomain
	}
}
