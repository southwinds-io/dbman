/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package plugin

// the interface implemented by database specific implementations of a database provider
type DatabaseProvider interface {
	// setup the provider with the specified configuration information
	Setup(config string) string

	// get database server general information
	GetInfo() string

	// get database release version information
	GetVersion() string

	// set database release version information
	SetVersion(versionInfo string) string

	// execute the specified command
	RunCommand(cmd string) string

	// execute the specified query
	RunQuery(query string) string
}
