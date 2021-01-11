package onelogin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/99designs/keyring"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/synfinatic/onelogin-aws-role/aws"
)

func saml_cache() string {
	return fmt.Sprintf("%s/.onelogin.aws.role.saml.json", os.Getenv("HOME"))
}

type OneLoginSAML struct {
	OneLogin  *OneLogin
	Assertion map[uint32]string
	Response  *SAMLResponse
	keyring   *keyring.Keyring
}

type SAMLResponse struct {
	// successful response
	Data    string        `json:"data"`
	Message string        `json:"message"`
	User    *OneLoginUser `json:"user"`

	// Fields when we need MFA
	StateToken  string      `json:"state_token"`
	Devices     []MfaDevice `json:"devices"`
	CallbackUrl string      `json:"callback_url"`
}

type SAMLAssertion struct {
	NotOnOrAfter int64  `json:"NotOnOrAfter"`
	Assertion    string `json:"Assertion"`
}

func (sr *SAMLResponse) NewMFA(o *OneLogin) *MFA {
	mfa := MFA{
		client:      o.Client,
		params:      map[string]string{},
		StateToken:  sr.StateToken,
		Devices:     sr.Devices,
		CallbackUrl: sr.CallbackUrl,
		User:        sr.User,
	}
	return &mfa
}

func NewOneLoginSAML(o *OneLogin, kr *keyring.Keyring) *OneLoginSAML {
	ols := OneLoginSAML{
		OneLogin:  o,
		Assertion: map[uint32]string{},
		keyring:   kr,
	}

	return &ols
}

// Returns true/false if MFA is required
func (ols *OneLoginSAML) GetAssertion(username string, password string, subdomain string, app_id uint32, ip string) (bool, error) {

	assertion, err := ols.LoadAssertion(app_id)
	if err == nil {
		// Loaded SAML Assertion from cache, so use that.
		ols.Assertion[app_id] = assertion
		return false, nil
	} else {
		log.Debugf("Unable to load assertion: %s", err.Error())
		err = nil
	}

	url := fmt.Sprintf("%s/api/2/saml_assertion", ols.OneLogin.Url)

	data := map[string]string{
		"username_or_email": username,
		"password":          password,
		"subdomain":         subdomain,
		"app_id":            fmt.Sprintf("%d", app_id),
	}
	if ip != "" {
		data["ip_address"] = ip
	}

	body, _ := json.Marshal(data)
	resp, err := ols.OneLogin.Client.R().
		SetResult(&SAMLResponse{}).
		SetBody(body).
		Post(url)
	if err != nil {
		return false, err
	} else if resp.IsError() {
		return false, fmt.Errorf("Unable to GetAssertion: %s [%d]", resp.String(), resp.StatusCode())
	}
	result := resp.Result().(*SAMLResponse)
	if result.Data != "" {
		decoded, err := base64.StdEncoding.DecodeString(result.Data)
		if err != nil {
			return false, fmt.Errorf("Unable to decode assertion: %s", err.Error())
		}

		ols.Assertion[app_id] = string(decoded)

		// save our assertion in the cache for later
		err = ols.SaveAssertion(app_id, ols.Assertion[app_id])
		if err != nil {
			log.Warn(err.Error())
		}

		log.Debugf("result.Data = %s", result.Data)
		return false, nil
	} else {
		log.Debug("no Data")
	}
	ols.Response = result

	err = ols.SaveAssertion(app_id, ols.Assertion[app_id])
	if err != nil {
		log.Warn(err.Error())
	}
	return true, nil
}

// Returns true/false if we got our assertion
func (ols *OneLoginSAML) SubmitMFA(app_id uint32, device_id int32, mfa_code int32) (bool, error) {
	mfa := ols.Response.NewMFA(ols.OneLogin)
	mfa.SetParam("app_id", fmt.Sprintf("%d", app_id))

	resp, err := mfa.SubmitMFA(device_id, mfa_code)
	if err != nil {
		return false, err
	}
	sr := SAMLResponse{}
	err = json.Unmarshal([]byte(resp), &sr)
	if err != nil {
		return false, err
	}

	if sr.Data != "" {
		ols.Assertion[app_id] = sr.Data
		return true, nil
	}
	return false, nil
}

// Returns true/false if we got our assertion
func (ols *OneLoginSAML) OneLoginProtectPush(app_id uint32, tries uint32) (bool, error) {
	mfa := ols.Response.NewMFA(ols.OneLogin)
	mfa.SetParam("app_id", fmt.Sprintf("%d", app_id))

	resp, err := mfa.OneLoginProtectPush(true)
	if err != nil {
		return false, err
	}
	log.Debugf("First MFA Push: %s", resp)
	sr := SAMLResponse{}
	err = json.Unmarshal([]byte(resp), &sr)

	var i uint32
	for i = 0; i < tries; i++ {
		if sr.Data != "" {
			break
		}
		time.Sleep(1 * time.Second)
		resp, err = mfa.OneLoginProtectPush(false)
		if err != nil {
			return false, err
		}
		log.Debugf("OLPP result: %s", resp)
		err = json.Unmarshal([]byte(resp), &sr)
	}

	if sr.Data != "" {
		decoded, err := base64.StdEncoding.DecodeString(sr.Data)
		if err != nil {
			return false, fmt.Errorf("Unable to decode assertion: %s", err.Error())
		}

		ols.Assertion[app_id] = string(decoded)

		// save our assertion in the cache for later
		err = ols.SaveAssertion(app_id, ols.Assertion[app_id])
		if err != nil {
			log.Warn(err.Error())
		}

		return true, nil
	}
	// timed out
	return false, nil
}

func (ols *OneLoginSAML) HasAssertion(app_id uint32) bool {
	_, ok := ols.Assertion[app_id]
	return ok
}

// saves our assertion in our keychain
func (ols *OneLoginSAML) SaveAssertion(app_id uint32, assertion string) error {
	f, err := os.OpenFile(saml_cache(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("Unable to open SAML Cache %s: %s", saml_cache(), err.Error())
	}
	defer f.Close()

	d := json.NewDecoder(f)
	assertions := map[string]SAMLAssertion{}

	d.Decode(&assertions)
	if err != nil {
		return err
	}

	t, err := aws.GetExpireTime(assertion)
	if err != nil {
		return err
	}

	f.Seek(0, 0)
	assertions[fmt.Sprintf("%d", app_id)] = SAMLAssertion{
		NotOnOrAfter: t.Unix(),
		Assertion:    assertion,
	}

	e := json.NewEncoder(f)
	err = e.Encode(assertions)
	if err != nil {
		return fmt.Errorf("Error writing to %s: %s", saml_cache(), err.Error())
	}

	return nil
}

func (ols *OneLoginSAML) LoadAssertion(app_id uint32) (string, error) {
	f, err := os.Open(saml_cache())
	if err != nil {
		return "", fmt.Errorf("Unable to open SAML Cache %s: %s", saml_cache(), err.Error())
	}
	defer f.Close()

	d := json.NewDecoder(f)
	assertions := map[string]SAMLAssertion{}
	err = d.Decode(&assertions)
	if err != nil {
		return "", fmt.Errorf("Error decoding %s: %s", saml_cache(), err.Error())
	}

	assertion, ok := assertions[fmt.Sprintf("%d", app_id)]
	if !ok {
		return "", fmt.Errorf("Unable to find %d in %s", app_id, saml_cache())
	}
	t := time.Unix(assertion.NotOnOrAfter, 0)
	if t.Before(time.Now()) {
		return "", fmt.Errorf("SAML Assertion for %d has expired", app_id)
	}
	return assertion.Assertion, nil
}

type SAMLInputOptions struct {
	Duration     *int64 `min:"900" type:"integer"` // seconds, default is 3600
	PrincipalArn string `min:"20" type:"string" required:"true"`
	RoleArn      string `min:"20" type:"string" required:"true"`
}

func (ols *OneLoginSAML) BuildSAMLInput(app_id uint32, options SAMLInputOptions) (*sts.AssumeRoleWithSAMLInput, error) {
	if !ols.HasAssertion(app_id) {
		return nil, fmt.Errorf("Missing SAML assertion for %d", app_id)
	}

	var duration int64 = 3600
	if options.Duration != nil {
		duration = *options.Duration
	}

	var assertion string = ols.Assertion[app_id]
	input := sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &duration,
		Policy:          nil,
		PolicyArns:      nil,
		PrincipalArn:    &options.PrincipalArn,
		RoleArn:         &options.RoleArn,
		SAMLAssertion:   &assertion,
	}
	return &input, nil
}
