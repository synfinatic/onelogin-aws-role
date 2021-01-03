package onelogin

/*
 * The Official OneLogin Go SDK doesn't support MFA :(
 */

import (
	"encoding/json"
	"fmt"
	"time"

	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type OneLogin struct {
	client            *resty.Client
	Url               string // api url for onelogin
	SessionLoginToken string // SLT generated via OAuth2
	// User Login values post OAuth2
	StateToken  string
	AccessToken string
	MfaDevices  []MfaDevice
	CallbackUrl string
	ReturnToUrl string
	ExpiresAt   string
	UserId      int32
}

type OneLoginStatus struct {
	Error   bool   `json:"error"`
	Code    int32  `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (ols *OneLoginStatus) Check(fatal bool) error {
	var err error = nil
	if ols.Error {
		if fatal {
			log.Fatalf("Error %s [%d]: %s", ols.Type, ols.Code, ols.Message)
		} else {
			err = fmt.Errorf("Error %s [%d]: %s", ols.Type, ols.Code, ols.Message)
		}
	}
	return err
}

type AccessTokenResponse struct {
	Status      OneLoginStatus `json:"status"`
	AccessToken string         `json:"access_token"`
	CreatedAt   string         `json:"created_at"`
	ExpiresIn   int32          `json:"expires_in"`
	TokenType   string         `json:"token_type"`
	AccountId   int32          `json:"account_id"`
}
type MfaDevice struct {
	DeviceType string `json:"device_type"`
	DeviceId   int32  `json:"device_id"`
}

type OneLoginUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Id        int32  `json:"id"`
}

type SessionLoginTokenResponse struct {
	Status OneLoginStatus          `json:"status"`
	Data   []SessionLoginTokenData `json:"data"`
}

type SessionLoginTokenData struct {
	Status       OneLoginStatus `json:"statusitempty"`
	User         OneLoginUser   `json:"user"`
	ReturnToUrl  string         `json:"return_to_url"`
	ExpiresAt    string         `json:"expires_at"`
	SessionToken string         `json:"session_token"`
	// If MFA is required then these are set
	StateToken  string      `json:"state_token"`
	CallbackUrl string      `json:"callback_url"`
	Devices     []MfaDevice `json:"devices"`
}

func NewOneLogin(clientid string, client_secret string, region string) *OneLogin {
	o := OneLogin{}
	o.Url = fmt.Sprintf("https://api.%s.onelogin.com", region)
	o.client = resty.New()
	o.client.SetHeader("Content-Type", "application/json")
	o.client.SetHeader("Accept", "application/json")

	data := map[string]string{
		"grant_type": "client_credentials",
	}

	url := fmt.Sprintf("%s/auth/oauth2/v2/token", o.Url)
	body, _ := json.Marshal(data)
	resp, err := o.client.R().
		SetBasicAuth(clientid, client_secret).
		SetResult(&AccessTokenResponse{}).
		SetBody(body).
		Post(url)
	if err != nil {
		log.Fatalf("Unable to auth with clientid/client_secret: %s", err)
	} else if resp.IsError() {
		log.Fatalf("Unable auth with clientid/client_secret: %s [%d]", resp.String(), resp.StatusCode())
	}

	result := resp.Result().(*AccessTokenResponse)
	result.Status.Check(true)

	if result.Status.Error {
		log.Fatalf("Error %s: %s", result.Status.Type, result.Status.Message)
	}

	// make other API calls shorter
	o.client.SetAuthToken(result.AccessToken).
		SetHeader("Content-Type", "application/json")

	o.AccessToken = result.AccessToken
	return &o
}

type MFA struct {
	DeviceId   int32
	DeviceType string
}

func (o *OneLogin) CreateSessionLoginToken(username string, password string, subdomain string) error {
	data := map[string]string{
		"username_or_email": username,
		"password":          password,
		"subdomain":         subdomain,
	}

	url := fmt.Sprintf("%s/api/1/login/auth", o.Url)
	body, _ := json.Marshal(data)
	resp, err := o.client.R().
		SetResult(&SessionLoginTokenResponse{}).
		SetBody(body).
		Post(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf("Unable to CreateSessionLoginToken: %s [%d]", resp.String(), resp.StatusCode())
	}
	result := resp.Result().(*SessionLoginTokenResponse)
	result.Status.Check(true)
	o.StateToken = result.Data[0].StateToken
	o.SessionLoginToken = result.Data[0].SessionToken
	o.MfaDevices = result.Data[0].Devices
	o.CallbackUrl = result.Data[0].CallbackUrl
	o.ReturnToUrl = result.Data[0].ReturnToUrl
	o.ExpiresAt = result.Data[0].ExpiresAt
	o.UserId = result.Data[0].User.Id
	return nil
}

func (o *OneLogin) OneLoginProtectPush(timeout int32) error {
	data := map[string]string{
		"state_token":   o.StateToken,
		"do_not_notify": "false",
	}
	var found_mfa bool

	for _, mfa := range o.MfaDevices {
		if mfa.DeviceType == "OneLogin Protect" {
			data["device_id"] = fmt.Sprintf("%d", mfa.DeviceId)
			found_mfa = true
			break
		}
	}

	if !found_mfa {
		return fmt.Errorf("Unable to find OneLogin Protect MFA device for your account")
	}

	body, _ := json.Marshal(data)
	log.Debugf("body = %s", body)
	resp, err := o.client.R().
		SetBody(body).
		SetResult(&SessionLoginTokenResponse{}).
		Post(o.CallbackUrl)
	if err != nil {
		return fmt.Errorf("Unable to use OneLogin Protect Push: %s", err.Error())
	} else if resp.IsError() {
		return fmt.Errorf("Unable to use OneLogin Protect Push: %s [%d]", resp.String(), resp.StatusCode())
	}
	result := resp.Result().(*SessionLoginTokenResponse)
	result.Status.Check(true)
	log.Debugf("%s: %v", o.CallbackUrl, result)

	// Don't notify user for subsequent calls
	data["do_not_notify"] = "true"
	body, _ = json.Marshal(data)

	for timeout != 0 {
		// keep asking the API to see if we got the response from user via App
		resp, err := o.client.R().
			SetBody(body).
			SetResult(&SessionLoginTokenResponse{}).
			Post(o.CallbackUrl)
		if err != nil {
			return fmt.Errorf("Unable to checks status of Protect Push: %s", err.Error())
		} else if resp.IsError() {
			return fmt.Errorf("Unable to check status of Protect Push: %s [%d]", resp.String(), resp.StatusCode())
		}
		result := resp.Result().(*SessionLoginTokenResponse)
		result.Status.Check(true)
		log.Debugf("%s: %v", o.CallbackUrl, resp)
		if result.Data != nil {
			o.SessionLoginToken = result.Data[0].SessionToken
			return nil
		}
		timeout -= 1
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Timed out waiting for MFA Push response")
}
