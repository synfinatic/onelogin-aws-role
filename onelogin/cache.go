package onelogin

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/synfinatic/onelogin-aws-role/aws"

	log "github.com/sirupsen/logrus"
)

func saml_cache() string {
	return fmt.Sprintf("%s/.onelogin.cache", os.Getenv("HOME"))
}

type OneLoginCache struct {
	filename    string
	Assertion   map[string]SAMLAssertion `json:"assertion"`
	AccessToken AccessTokenResponse      `json:"accesstoken"`
}

type SAMLAssertion struct {
	NotOnOrAfter int64    `json:"NotOnOrAfter"`
	Assertion    string   `json:"Assertion"`
	Roles        []string `json:"Roles"`
}

func LoadOneLoginCache(filename string) *OneLoginCache {
	fname := filename
	_, err := os.Stat(fname)
	if err != nil {
		fname = saml_cache()
	}
	f, err := os.Open(fname)
	if err != nil {
		c := OneLoginCache{
			filename:    fname,
			Assertion:   map[string]SAMLAssertion{},
			AccessToken: AccessTokenResponse{},
		}
		return &c
	}

	defer f.Close()

	d := json.NewDecoder(f)
	c := OneLoginCache{}
	err = d.Decode(&c)
	if err != nil {
		log.Errorf("Corrupted cache file %s: %s", fname, err.Error())
		c = OneLoginCache{
			filename:    fname,
			Assertion:   map[string]SAMLAssertion{},
			AccessToken: AccessTokenResponse{},
		}
		return &c
	}
	if c.filename == "" {
		c.filename = fname
	}
	return &c
}

func (olc *OneLoginCache) Save() error {
	f, err := os.OpenFile(olc.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to open %s: %s", olc.filename, err.Error())
	}
	defer f.Close()

	e := json.NewEncoder(f)
	err = e.Encode(olc)
	if err != nil {
		return fmt.Errorf("Error writing %s: %s", olc.filename, err.Error())
	}
	return nil
}

// saves our assertion in our keychain
func (olc *OneLoginCache) SaveAssertion(app_id uint32, assertion string) error {
	roles, err := aws.GetRoles(assertion)
	if err != nil {
		return fmt.Errorf("Unable to save assertion: %s", err.Error())
	}
	t, err := aws.GetExpireTime(assertion)
	if err != nil {
		return fmt.Errorf("Unable to save assertion: %s", err.Error())
	}

	id := fmt.Sprintf("%d", app_id)
	x := SAMLAssertion{
		NotOnOrAfter: t.Unix(),
		Assertion:    assertion,
		Roles:        roles,
	}
	if olc.Assertion == nil {
		olc.Assertion = map[string]SAMLAssertion{}
	}
	olc.Assertion[id] = x
	return olc.Save()
}

func (olc *OneLoginCache) SaveAccessToken(token *AccessTokenResponse) error {
	olc.AccessToken = *token
	return olc.Save()
}

func (olc *OneLoginCache) GetAccessToken() (string, error) {
	if olc.AccessToken.AccessToken == "" {
		return "", fmt.Errorf("No current OAuth2 AccessToken")
	}
	now := time.Now()
	expires := time.Unix(olc.AccessToken.ExpiresIn, 0)
	if expires.Before(now) {
		return "", fmt.Errorf("OAuth2 AccessToken has expired")
	}
	return olc.AccessToken.AccessToken, nil
}

func (olc *OneLoginCache) GetAssertion(app_id uint32) (string, error) {
	id := fmt.Sprintf("%d", app_id)
	assertion, ok := olc.Assertion[id]
	if !ok {
		return "", fmt.Errorf("Unable to find assertion %d", app_id)
	}
	t := time.Unix(assertion.NotOnOrAfter, 0)
	if t.Before(time.Now()) {
		return "", fmt.Errorf("SAML Assertion for %d has expired", app_id)
	}
	return assertion.Assertion, nil
}

func (ols *OneLoginCache) GetRoles(app_id uint32) ([]string, error) {
	id := fmt.Sprintf("%d", app_id)
	assertion, ok := ols.Assertion[id]
	if !ok {
		return []string{}, fmt.Errorf("Unable to find assertion %d", app_id)
	}
	t := time.Unix(assertion.NotOnOrAfter, 0)
	if t.Before(time.Now()) {
		return []string{}, fmt.Errorf("SAML Assertion for %d has expired", app_id)
	}
	return assertion.Roles, nil

}
