/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	"github.com/spf13/cobra"
	"southwinds.dev/dbman/core"
)

type ReleaseCmd struct {
	cmd  *cobra.Command
	info *core.ScriptManager
}

func NewReleaseCmd() *ReleaseCmd {
	c := &ReleaseCmd{
		cmd: &cobra.Command{
			Use:   "release",
			Short: "shows release information",
			Long:  ``,
		}}
	return c
}
