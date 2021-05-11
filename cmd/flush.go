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
)

type FlushCmd struct {
	Profile string `kong:"arg,required,name='profile',help='AWS Role alias name'"`
}

func (cc *FlushCmd) Run(ctx *RunContext) error {
	cli := *ctx.Cli
	kr, err := OpenKeyring(nil)
	if err != nil {
		return fmt.Errorf("Unable to open KeyChain: %s", err)
	}
	return kr.RemoveSTSSession(cli.Flush.Profile)
}
