package onelogin

/*
 * OneLogin AWS Role
 * Copyright (c) 2020-2021 Aaron Turner  <aturner at synfin dot net>
 *
 * This program is free software: you can redistribute it
 * and/or modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or with the authors permission any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/Songmu/prompter"
	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/onelogin-aws-role/utils"
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
	DeviceType string `json:"device_type" header:"MFA Device Type"`
	DeviceId   int32  `json:"device_id" header:"MFA Device ID"`
}

const HEADER_TAG = "header"

type MfaSelect struct {
	Select     string `header:"Select"`
	DeviceType string `header:"MFA Device Type"`
	DeviceId   int32  `header:"MFA Device ID"`
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

func GenerateMfaSelect(devices []MfaDevice) *[]MfaSelect {
	mfaDevices := make([]MfaSelect, len(devices))
	for i, mfa := range devices {
		mfaDevices[i] = MfaSelect{
			Select:     fmt.Sprintf("%d", 1+i),
			DeviceId:   mfa.DeviceId,
			DeviceType: mfa.DeviceType,
		}
	}

	return &mfaDevices
}

func SelectMfaDevice(mfaDevices []MfaDevice) int32 {
	// If there is only one, don't bother prompting user
	if len(mfaDevices) == 1 {
		return mfaDevices[0].DeviceId
	}

	m := GenerateMfaSelect(mfaDevices)
	mfaSelect := *m
	fields := []string{
		"Select",
		"DeviceType",
		"DeviceId",
	}

	ts := []utils.TableStruct{}
	for _, mfa := range mfaSelect {
		ts = append(ts, mfa)
	}
	utils.GenerateTable(ts, fields)
	fmt.Printf("\n")

	var mfaid int32 = 0
	for mfaid == 0 {
		sel := prompter.Prompt("Select MFA Device", "")
		x, err := strconv.ParseInt(sel, 10, 32)
		if err != nil || x > int64(len(mfaDevices)) || x < 1 {
			log.Errorf("Invalid MFA selector: please choose 1-%d", len(mfaDevices))
			mfaid = 0
			continue
		}
		mfa := mfaSelect[x-1]
		mfaid = mfa.DeviceId
	}
	return mfaid
}

func (mfa MfaDevice) GetHeader(fieldName string) (string, error) {
	v := reflect.ValueOf(mfa)
	return utils.GetHeaderTag(v, fieldName)
}

func (mfa MfaSelect) GetHeader(fieldName string) (string, error) {
	v := reflect.ValueOf(mfa)
	return utils.GetHeaderTag(v, fieldName)
}
