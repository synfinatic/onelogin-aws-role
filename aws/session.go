package aws

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
	"fmt"
	"time"
)

type STSSession struct {
	RoleARN         string    `json:"ROLE_ARN"`
	AccessKeyID     string    `json:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string    `json:"AWS_SECRET_ACCESS_KEY"`
	SessionToken    string    `json:"AWS_SESSION_TOKEN"`
	Expiration      time.Time `json:"AWS_SESSION_EXPIRATION"`
	Provider        string    `json:"STS_PROVIDER"`
	Issuer          string    `json:"STS_ISSUER"`
	Region          string    `json:"-"`
}

func (s *STSSession) Expired() bool {
	// 5 seconds of fuzz
	if s.Expiration.Before(time.Now().Add(time.Second * 5)) {
		return true
	}
	return false
}

func (s *STSSession) GetExpireTimeString() string {
	t := s.Expiration
	now := time.Now()
	delta := t.Sub(now)
	if delta < 0 {
		return "Expired"
	}

	ret := ""
	min := delta.Minutes()
	if min > 60.0 {
		hours := int(min) % 60
		min = min - float64(hours*60)
		ret = fmt.Sprintf("%02dh ", int(hours))
	} else {
		ret = fmt.Sprintf("00h ")
	}
	minutes := int(min)
	ret = fmt.Sprintf("%s%02dm", ret, minutes)
	return ret
}
