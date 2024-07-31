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
	"os"
	. "southwinds.dev/dbman/core"
	"strings"
)

type DbRunCmd struct {
	cmd *cobra.Command
}

func NewDbRunCmd() *DbRunCmd {
	c := &DbRunCmd{
		cmd: &cobra.Command{
			Use:   "run [command1, command2, ...]",
			Short: "runs one or more commands defined in the release manifest",
			Long:  ``,
		},
	}
	c.cmd.Run = c.Run
	return c
}

func (c *DbRunCmd) Run(_ *cobra.Command, args []string) {
	// check the query name has been passed in
	if len(args) == 0 {
		fmt.Printf("!!! You forgot to tell me the name of the command(s) you want to run\n")
		return
	}
	output, err, elapsed := DM.Run(strings.Split(args[0], ","))
	fmt.Print(output.String())
	if err != nil {
		fmt.Printf("!!! I cannot execute the requested commands\n")
		fmt.Printf("%v\n", err)
		fmt.Printf("? the execution time was %v\n", elapsed)
		os.Exit(1)
	}
	fmt.Printf("? I have executed the requested commands in %v\n", elapsed)
}
