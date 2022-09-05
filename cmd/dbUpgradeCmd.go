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

type DbUpgradeCmd struct {
	cmd *cobra.Command
}

func NewDbUpgradeCmd() *DbUpgradeCmd {
	c := &DbUpgradeCmd{
		cmd: &cobra.Command{
			Use:   "upgrade",
			Short: "upgrade an existing database to the current Application Version",
			Long:  ``,
		},
	}
	c.cmd.Run = c.Run
	return c
}

func (c *DbUpgradeCmd) Run(cmd *cobra.Command, args []string) {
	output, err, elapsed := DM.Upgrade()
	fmt.Print(output.String())
	if err != nil {
		fmt.Printf("!!! I cannot upgrade the database\n")
		fmt.Printf("%v\n", err)
		fmt.Printf("? the execution time was %v\n", elapsed)
		return
	}
	fmt.Printf("? I have upgraded the database in %v\n", elapsed)
}
