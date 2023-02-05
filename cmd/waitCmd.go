/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	. "southwinds.dev/dbman/core"
)

type WaitCmd struct {
	cmd      *cobra.Command
	attempts int
	interval int
}

func NewWaitCmd() *WaitCmd {
	c := &WaitCmd{
		cmd: &cobra.Command{
			Use:   "wait",
			Short: "wait until a connection to the database server can be established",
			Long: `wait until a connection to the database server can be established, trying a number of attempts every interval
`,
		}}
	c.cmd.Flags().IntVarP(&c.attempts, "attempts", "a", 10, "-a 10; the number of attempts before exiting with code 1")
	c.cmd.Flags().IntVarP(&c.attempts, "interval", "i", 3, "-i 3; the number of seconds between attempts")
	c.cmd.Run = c.Run
	return c
}

func (c *WaitCmd) Run(cmd *cobra.Command, args []string) {
	if err := DM.WaitForConnection(c.attempts, c.interval); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}
