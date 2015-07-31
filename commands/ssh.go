// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package commands

import (
	"fmt"
	"os/exec"

	"github.com/pagodabox/nanobox-cli/ui"
	"github.com/pagodabox/nanobox-golang-stylish"
)

// SSHCommand satisfies the Command interface
type SSHCommand struct{}

// Help
func (c *SSHCommand) Help() {
	ui.CPrint(`
Description:
  SSHes into the nanobox VM by issuing a "vagrant ssh"

Usage:
  nanobox ssh
  `)
}

// Run
func (c *SSHCommand) Run(opts []string) {

	// run 'vagrant ssh'
	fmt.Printf(stylish.Bullet("SSHing into nanobox VM..."))
	runVagrantCommand(exec.Command("vagrant", "ssh"))
}
