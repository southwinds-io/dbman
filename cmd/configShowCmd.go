/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	. "southwinds.dev/dbman/core"
)

type ConfigShowCmd struct {
	cmd *cobra.Command
}

func NewConfigShowCmd() *ConfigShowCmd {
	c := &ConfigShowCmd{
		cmd: &cobra.Command{
			Use:     "show",
			Short:   "show the configuration values for the current set",
			Example: `dbman config show`,
		}}
	c.cmd.Run = c.Run
	return c
}

func (c *ConfigShowCmd) Run(cmd *cobra.Command, args []string) {
	fmt.Printf("? I am showing the content of %v\n", DM.Cfg.ConfigFileUsed())
	fmt.Print(DM.ConfigSetAsString())
}
