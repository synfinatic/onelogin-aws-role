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
	"reflect"
	"strconv"
	"strings"

	yaml "github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
)

// ConfigFile structure
type ConfigFile struct {
	ClientID     string                `yaml:"client_id"`
	ClientSecret string                `yaml:"client_secret"`
	Region       string                `yaml:"region"`                           // OneLogin Region
	Username     string                `yaml:"username"`                         // or email address
	Subdomain    string                `yaml:"subdomain"`                        // XXXX.onelogin.com
	Mfa          string                `yaml:"mfa"`                              // MFA device_id to use by default
	Duration     uint32                `yaml:"duration"`                         // Default duration (in seconds) for credentials
	Accounts     *map[uint64]string    `yaml:"aws_accounts" header:"AccountID"`  // AWS AccountID is the key
	Apps         *map[uint32]AppConfig `yaml:"apps" header:"AppID"`              // OneLogin AppID is the key
	Fields       *[]string             `yaml:"fields,omitempty" header:"Fields"` // List of fields to report with `list` command
}

// App config
type AppConfig struct {
	Name  string        `yaml:"name" header:"App Name"`
	Alias string        `yaml:"alias" header:"App Alias"`
	Roles *[]RoleConfig `yaml:"roles"`
}

// Role config
type RoleConfig struct {
	Arn    string `yaml:"arn" header:"ARN"`
	Alias  string `yaml:"alias" header:"Role Alias"`
	Region string `yaml:"region" header:"Default Region"` // Default AWS Region
}

// Flattened Config for displaying report
type FlatConfig struct {
	AccountId   uint64 `header:"AWS AccountID"`
	AccountName string `header:"Account Name"`
	AppId       uint32 `header:"OneLogin AppID"`
	AppName     string `header:"App Name"`
	AppAlias    string `header:"App Alias"`
	Arn         string `header:"Role ARN"`
	RoleAlias   string `header:"Role Alias"`
	Region      string `header:"Default Region"`
}

const (
	CONFIG_YAML     string = "~/.onelogin.yaml"
	FLAT_CONFIG_TAG        = "header"
)

// Returns the config file path.  If `path` is empty, use CONFIG_YAML
func GetPath(path string) string {
	cfg := CONFIG_YAML
	if path != "" {
		cfg = path
	}
	return strings.Replace(cfg, "~", os.Getenv("HOME"), 1)
}

func (c *ConfigFile) GetFlatConfig() []FlatConfig {
	fc := []FlatConfig{}

	for appid, app := range *c.Apps {
		for _, role := range *app.Roles {
			accountid, err := GetAccountFromARN(role.Arn)
			if err != nil {
				log.WithError(err).Warnf("Unable to get AWS Account ID for role '%s'", role.Arn)
			}
			a := *c.Accounts
			accountname, _ := a[accountid]
			fc = append(fc, FlatConfig{
				AccountId:   accountid,
				AccountName: accountname,
				AppId:       appid,
				AppName:     app.Name,
				AppAlias:    app.Alias,
				Arn:         role.Arn,
				RoleAlias:   role.Alias,
				Region:      role.Region,
			})
		}
	}
	return fc
}

// Parses the AWS Account ID from an ARN
func GetAccountFromARN(arn string) (uint64, error) {
	fields := strings.Split(arn, ":")
	if len(fields) < 5 || fields[4] == "" {
		return 0, fmt.Errorf("unable to parse %s", arn)
	}
	val, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// Loads our config file at the given path
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

//  Get Roles.  Returns alias => ARN
func (c *ConfigFile) GetRoles() *map[string]string {
	ret := map[string]string{}

	for _, app := range *c.Apps {
		for _, role := range *app.Roles {
			ret[role.Alias] = role.Arn
		}
	}
	return &ret
}

//  Get Apps.  Returns id => alias
func (c *ConfigFile) GetApps() *map[uint32]string {
	ret := map[uint32]string{}

	for id, app := range *c.Apps {
		ret[id] = app.Alias
	}
	return &ret
}

// Return the AWS Role ARN based on the alias
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

func getHeader(v reflect.Value, fieldName string) (string, error) {
	field, ok := v.Type().FieldByName(fieldName)
	if !ok {
		return "", fmt.Errorf("Invalid field '%s' in %s", fieldName, v.Type().Name())
	}
	tag := string(field.Tag.Get(FLAT_CONFIG_TAG))
	return tag, nil
}

func (cf *ConfigFile) GetHeader(fieldName string) (string, error) {
	v := reflect.ValueOf(*cf)
	return getHeader(v, fieldName)
}

func (ac *AppConfig) GetHeader(fieldName string) (string, error) {
	v := reflect.ValueOf(*ac)
	return getHeader(v, fieldName)
}

func (rc *RoleConfig) GetHeader(fieldName string) (string, error) {
	v := reflect.ValueOf(*rc)
	return getHeader(v, fieldName)
}

func (c *FlatConfig) GetHeader(fieldName string) (string, error) {
	v := reflect.ValueOf(*c)
	return getHeader(v, fieldName)
}
