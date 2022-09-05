/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package plugin

import "net/rpc"

// Database Provider RPC client
type DatabaseProviderRPC struct {
	Client *rpc.Client
}

func (db *DatabaseProviderRPC) Setup(config string) string {
	var result string
	err := db.Client.Call("Plugin.Setup", config, &result)
	if err != nil {
		output := NewParameter()
		output.SetError(err)
		return output.ToString()
	}
	return result
}

func (db *DatabaseProviderRPC) GetVersion() string {
	var result string
	err := db.Client.Call("Plugin.GetVersion", "", &result)
	if err != nil {
		return db.errorToString(err)
	}
	return result
}

func (db *DatabaseProviderRPC) RunCommand(cmd string) string {
	var result string
	err := db.Client.Call("Plugin.RunCommand", cmd, &result)
	if err != nil {
		return db.errorToString(err)
	}
	return result
}

func (db *DatabaseProviderRPC) SetVersion(args string) string {
	var result string
	err := db.Client.Call("Plugin.SetVersion", args, &result)
	if err != nil {
		return db.errorToString(err)
	}
	return result
}

func (db *DatabaseProviderRPC) RunQuery(query string) string {
	var result string
	err := db.Client.Call("Plugin.RunQuery", query, &result)
	if err != nil {
		return db.errorToString(err)
	}
	return result
}

func (db *DatabaseProviderRPC) GetInfo() string {
	var result string
	err := db.Client.Call("Plugin.GetInfo", "", &result)
	if err != nil {
		return db.errorToString(err)
	}
	return result
}

func (db *DatabaseProviderRPC) errorToString(err error) string {
	output := NewParameter()
	output.SetError(err)
	return output.ToString()
}
