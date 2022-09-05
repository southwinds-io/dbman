/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package main

import (
	. "southwinds.dev/dbman/plugin"
)

// the entry point for the PGSQL database plugin
// this plugin is also implemented as a native DbMan provider (not as a plugin) but
// offered here as an archetype of a database plugin to help understand how to create
// other plugins
func main() {
	// launch the plugin process
	// the plugin name must not start with "_" as it is reserved for native plugins
	ServeDbPlugin("pgsql", new(PgSQLProvider))
}
