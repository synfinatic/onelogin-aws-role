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
	// log "github.com/sirupsen/logrus"
)

type ConfigFile struct {
	ClientID     string                `yaml:"client_id"`
	ClientSecret string                `yaml:"client_secret"`
	Region       string                `yaml:"region"`       // OneLogin Region
	Username     string                `yaml:"username"`     // or email address
	Subdomain    string                `yaml:"subdomain"`    // XXXX.onelogin.com
	Mfa          string                `yaml:"mfa"`          // MFA device_id to use by default
	Accounts     *map[uint64]string    `yaml:"aws_accounts"` // AWS AccountID is the key
	Apps         *map[uint32]AppConfig `yaml:"apps"`         // OneLogin AppID is the key
}

type AppConfig struct {
	Name  string        `yaml:"name"`
	Alias string        `yaml:"alias"`
	Roles *[]RoleConfig `yaml:"roles"`
}

type RoleConfig struct {
	Arn    string `yaml:"arn"`
	Alias  string `yaml:"alias"`
	Region string `yaml:"region"` // AWS Region
}

const (
	CONFIG_YAML string = "~/.onelogin.yaml"
)

func GetPath(path string) string {
	cfg := CONFIG_YAML
	if path != "" {
		cfg = path
	}
	return strings.Replace(cfg, "~", os.Getenv("HOME"), 1)
}

func LoadConfigFile(path string) (*ConfigFile, error) {
	fullpath := GetPath(path)
	info, err := os.Stat(fullpath)
	if err != nil {
		return nil, fmt.Errorf("Unable to stat %s: %s", fullpath, err.Error())
	}

	file, err := os.Open(fullpath)
	if err != nil {
		return nil, fmt.Errorf("Unable to open %s: %s", fullpath, err.Error())
	}

	defer file.Close()

	buf := make([]byte, info.Size())
	_, err = file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to read %s: %s", fullpath, err.Error())
	}

	c := ConfigFile{}
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s: %s", fullpath, err.Error())
	}

	return &c, nil
}

/*
 * Get Roles.  Returns alias => ARN
 */
func (c *ConfigFile) GetRoles() *map[string]string {
	ret := map[string]string{}

	for _, app := range *c.Apps {
		for _, role := range *app.Roles {
			ret[role.Alias] = role.Arn
		}
	}
	return &ret
}

/*
 * Get Apps.  Returns alias => id
 */
func (c *ConfigFile) GetApps() *map[string]uint32 {
	ret := map[string]uint32{}

	for id, app := range *c.Apps {
		ret[app.Alias] = id
	}
	return &ret
}

/*
 * Return the AWS Role ARN based on the alias
 */
func (c *ConfigFile) GetRoleArn(alias string) (string, error) {
	for _, app := range *c.Apps {
		for _, role := range *app.Roles {
			if role.Alias == alias {
				return role.Arn, nil
			}
		}
	}
	return "", fmt.Errorf("Unable to locate Role: %s", alias)
}

/*
 * Find an app using the Id or alias
 */
func (c *ConfigFile) GetApp(alias_or_id string) (*AppConfig, error) {
	for id, val := range *c.Apps {
		if val.Alias == alias_or_id || fmt.Sprintf("%d", id) == alias_or_id {
			return &val, nil
		}
	}
	return nil, fmt.Errorf("Unable to locate AppId: %s", alias_or_id)
}

/*
 * Find the AppID for a Role Alias
 */
func (c *ConfigFile) GetAppIdForRole(alias string) (uint32, error) {
	for id, app := range *c.Apps {
		if *app.Roles == nil {
			continue
		}
		for _, role := range *app.Roles {
			if role.Alias == alias {
				return id, nil
			}
			if role.Arn == alias {
				return id, nil
			}
		}
	}
	return 0, fmt.Errorf("Unable to find Role with alias or name: %s", alias)
}
