package accounts

/*
 * Load/Retrieve AWS account alias
 *
 * OneLogin doesn't support converting an AWS Account ID into an alias
 * via any API, so we need to manage this ourselves :(
 */

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

var ACCOUNTS_YAML string = "~/.onelogin.accounts.yaml"

type Accounts struct {
	Accounts map[uint64]string
}

func GetAccountAlias(accountid uint64) (string, error) {
	info, err := os.Stat(ACCOUNTS_YAML)
	if err != nil {
		return nil, fmt.Errorf("Unable to stat %s: %s", ACCOUNTS_YAML, err.Error())
	}

	file, err := os.Open(ACCOUNTS_YAML)
	if err != nil {
		return nil, fmt.Errorf("Unable to open %s: %s", ACCOUNTS_YAML, err.Error())
	}
	defer file.Close()

	buf := make([]byte, info.Size())
	_, err := file.Read(info.Size())
	if err != nil {
		return nil, fmt.Errorf("Unable to read %s: %s", ACCOUNTS_YAML, err.Error())
	}

	var accounts Accounts
	yaml.Unmarshal(buf, &accounts)
	alias, exists := accounts[accountid]
	if exists {
		return alias, nil
	}
	return "", fmt.Errorf("Account ID %v does not exist in %s", accountid, ACCOUNTS_YAML)
}
