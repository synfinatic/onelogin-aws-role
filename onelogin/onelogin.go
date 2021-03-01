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
	Client      *resty.Client
	Cache       *OneLoginCache
	Url         string // api url for onelogin
	AccessToken string // generated via OAuth2.  Required for all other API calls

}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	CreatedAt   string `json:"created_at"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
	AccountId   int64  `json:"account_id"`
}

// Not used by this code, but is common to many other API calls
type OneLoginUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Id        int64  `json:"id"`
}

// Common Status response
type OneLoginStatus struct {
	Error   bool   `json:"error"`
	Code    uint16 `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

/*
 * Returns a new OneLogin struct with our AccessToken configured
 *
 * OneLogin OAuth2 tokens are good for 10hrs
 */
func NewOneLogin(clientid string, client_secret string, region string) (*OneLogin, error) {
	cache := LoadOneLoginCache("")
	o := OneLogin{
		Cache: cache,
	}

	if clientid == "" {
		return nil, fmt.Errorf("Missing client_id value in config file")
	}
	if client_secret == "" {
		return nil, fmt.Errorf("Missing client_secret value in config file")
	}
	if region == "" {
		region = "us"
	}
	token, err := o.Cache.GetAccessToken()
	if err != nil || token == "" {

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
		token = result.AccessToken
		o.Cache.SaveAccessToken(result)
	}
	//	result.Status.Check(true)

	//	if result.Status.Error {
	//		log.Fatalf("Error %s: %s", result.Status.Type, result.Status.Message)
	//	}

	// make other API calls shorter
	o.Client.SetAuthToken(token).
		SetHeader("Content-Type", "application/json")

	return &o, nil
}

// returns true if the given OAuth2 token has expired
func (token *AccessTokenResponse) IsExpired() bool {
	created_at, err := time.Parse("2006-01-02T15:04:05.000Z", token.CreatedAt)
	if err != nil {
		log.Fatalf("Unable to parse %s: %s", token.CreatedAt, err.Error())
	}
	expires_at := created_at.Add(time.Second * time.Duration(token.ExpiresIn))
	now := time.Now()
	if now.After(expires_at) {
		return true
	}
	return false
}

// returns when our token expires
func (token *AccessTokenResponse) ExpiresAt() string {
	created_at, err := time.Parse("2006-01-02T15:04:05.000Z", token.CreatedAt)
	if err != nil {
		log.Fatalf("Unable to parse %s: %s", token.CreatedAt, err.Error())
	}
	expires_at := created_at.Add(time.Second * time.Duration(token.ExpiresIn))
	return expires_at.Local().String()
}

type RateLimit struct {
	Status OneLoginStatus `json:"status"`
	Data   RateLimitData  `json:"data"`
}

type RateLimitData struct {
	Limit     uint32 `json:"X-RateLimit-Limit"`
	Remaining uint32 `json:"X-RateLimit-Remaining"`
	Reset     uint32 `json:"X-RateLimit-Reset"`
}

/*
 * There is a rate limit for generating auth tokens:
 * https://developers.onelogin.com/api-docs/2/oauth20-tokens/generate-tokens-2
 *
 * This API call returns how many calls have been made.
 * Not valid with Authentication Only tokens
 */
func (o *OneLogin) GetRateLimit() (*RateLimit, error) {
	url := fmt.Sprintf("%s/auth/rate_limit", o.Url)
	resp, err := o.Client.R().
		SetResult(&RateLimit{}).
		Post(url)
	if err != nil {
		return nil, fmt.Errorf("Unable to get rate_limit for %s: %s", url, err.Error())
	} else if resp.IsError() {
		return nil, fmt.Errorf("Unable to get rate_limit for %s: %s [%d]", url, resp.String(), resp.StatusCode())
	}
	result := resp.Result().(*RateLimit)
	log.Debugf("RateLimit: %s", resp.String())
	return result, nil
}
