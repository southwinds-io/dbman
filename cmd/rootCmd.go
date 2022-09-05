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
	"southwinds.dev/dbman/core"
)

type RootCmd struct {
	*cobra.Command
}

// https://textkool.com/en/ascii-art-generator?hl=default&vl=default&font=Broadway%20KB&text=dbman%0A

func NewRootCmd() *RootCmd {
	c := &RootCmd{
		&cobra.Command{
			Use:   "dbman",
			Short: "database manager",
			Long: `
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++
|              ___   ___   _       __    _              |
|             | | \ | |_) | |\/|  / /\  | |\ |          |
|             |_|_/ |_|_) |_|  | /_/--\ |_| \|          |
|                  Manage Database Schemas              |
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++
dbman is a CLI tool to manage database schema versions and upgrades.
dbman can also be run from a container (when in http mode) to manage the data / schema life cycle of databases from a container platform.`,
		},
	}
	cobra.OnInitialize(c.initConfig)
	return c
}

// initConfig reads in config file and ENV variables if set.
func (c *RootCmd) initConfig() {
	dm, err := core.NewDbMan()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	core.DM = dm
}
