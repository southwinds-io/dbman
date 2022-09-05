/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	"github.com/spf13/cobra"
	. "southwinds.dev/dbman/core"
)

type ServeCmd struct {
	cmd *cobra.Command
}

func NewServeCmd() *ServeCmd {
	c := &ServeCmd{
		cmd: &cobra.Command{
			Use:     "serve",
			Short:   "runs dbman as an http service",
			Example: `dbman serve`,
		}}
	c.cmd.Run = c.Run
	return c
}

func (c *ServeCmd) Run(cmd *cobra.Command, args []string) {
	DM.Serve()
}
