package onelogin

/*
 * The Official OneLogin Go SDK doesn't support MFA :(
 */

import (
	"encoding/json"
	"fmt"

	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type OneLogin struct {
	Client      *resty.Client
	Url         string // api url for onelogin
	AccessToken string // generated via OAuth2.  Required for all other API calls

}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	CreatedAt   string `json:"created_at"`
	ExpiresIn   int32  `json:"expires_in"`
	TokenType   string `json:"token_type"`
	AccountId   int32  `json:"account_id"`
}

// Not used by this code, but is common to many other API calls
type OneLoginUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Id        int32  `json:"id"`
}

// Returns a new OneLogin struct with our AccessToken configured
func NewOneLogin(clientid string, client_secret string, region string) *OneLogin {
	o := OneLogin{}
	o.Url = fmt.Sprintf("https://api.%s.onelogin.com", region)
	o.Client = resty.New()
	o.Client.SetHeader("Content-Type", "application/json")
	o.Client.SetHeader("Accept", "application/json")

	data := map[string]string{
		"grant_type": "client_credentials",
	}

	url := fmt.Sprintf("%s/auth/oauth2/v2/token", o.Url)
	body, _ := json.Marshal(data)
	resp, err := o.Client.R().
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
	//	result.Status.Check(true)

	//	if result.Status.Error {
	//		log.Fatalf("Error %s: %s", result.Status.Type, result.Status.Message)
	//	}

	// make other API calls shorter
	o.Client.SetAuthToken(result.AccessToken).
		SetHeader("Content-Type", "application/json")

	o.AccessToken = result.AccessToken
	return &o
}
