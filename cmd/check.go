package main

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

	"github.com/synfinatic/onelogin-aws-role/onelogin"
)

type CheckCmd struct {
	// No sub options
}

func (cc *CheckCmd) Run(ctx *RunContext) error {
	c := onelogin.LoadOneLoginCache("")
	expired := c.AccessToken.IsExpired()
	if expired {
		fmt.Println("OneLogin OAuth2 token has expired.")
	} else {
		fmt.Printf("OneLogin OAuth2 token will expire at: %s\n", c.AccessToken.ExpiresAt())
	}

	return nil
}
