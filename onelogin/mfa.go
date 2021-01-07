package onelogin

import (
	"encoding/json"
	"fmt"

	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type MFA struct {
	client *resty.Client
	params map[string]string
	// Fields when we need MFA
	StateToken  string        `json:"state_token"`
	Devices     []MfaDevice   `json:"devices"`
	CallbackUrl string        `json:"callback_url"`
	User        *OneLoginUser `json:"user"`
}

type MfaDevice struct {
	DeviceType string `json:"device_type"`
	DeviceId   int32  `json:"device_id"`
}

func (mfa *MFA) SetParam(key string, value string) {
	mfa.params[key] = value
}

// returns json encoded result
func (mfa *MFA) SubmitMFA(device_id int32, mfa_code int32) (string, error) {
	data := map[string]string{
		"state_token": mfa.StateToken,
		"device_id":   fmt.Sprintf("%d", device_id),
		"otp_token":   fmt.Sprintf("%d", mfa_code),
	}
	for k, v := range mfa.params {
		data[k] = v
	}
	body, _ := json.Marshal(data)
	resp, err := mfa.client.R().
		SetBody(body).
		Post(mfa.CallbackUrl)
	if err != nil {
		return "", fmt.Errorf("Unable to submit MFA token code: %s", err.Error())
	} else if resp.IsError() {
		return "", fmt.Errorf("Unable to submit MFA token code: %s [%d]", resp.String(), resp.StatusCode())
	}
	return resp.String(), nil
}

// returns json encoded result
func (mfa *MFA) OneLoginProtectPush(notify bool) (string, error) {
	var dnn string = "false"
	if !notify {
		dnn = "true"
	}
	data := map[string]string{
		"state_token":   mfa.StateToken,
		"do_not_notify": dnn,
	}
	for k, v := range mfa.params {
		data[k] = v
	}
	var found_mfa bool

	for _, device := range mfa.Devices {
		if device.DeviceType == "OneLogin Protect" {
			data["device_id"] = fmt.Sprintf("%d", device.DeviceId)
			found_mfa = true
			break
		}
	}

	if !found_mfa {
		return "", fmt.Errorf("Unable to find OneLogin Protect MFA device for your account")
	}

	body, _ := json.Marshal(data)
	log.Debugf("Push MFA: %s", body)
	resp, err := mfa.client.R().
		SetBody(body).
		Post(mfa.CallbackUrl)
	if err != nil {
		return "", fmt.Errorf("Unable to use OneLogin Protect Push: %s", err.Error())
	} else if resp.IsError() {
		return "", fmt.Errorf("Unable to use OneLogin Protect Push: %s [%d]", resp.String(), resp.StatusCode())
	}
	return resp.String(), nil
}

/*
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
			o.SessionToken = result.Data[0].SessionToken
			return nil
		}
		timeout -= 1
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Timed out waiting for MFA Push response")
}
*/
