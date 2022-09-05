/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package main

import (
	"fmt"
	"southwinds.dev/dbman/core"
	. "southwinds.dev/dbman/plugin"
	"testing"
)

func TestPgSQLProvider_GetVersion(t *testing.T) {
	// read the config
	cfg := core.NewConfig("", "")
	// transform cfg into map
	conf, _ := NewConf(cfg.All())
	// creates the provider
	dbProvider := &PgSQLProvider{
		cfg: conf,
	}
	// test get version
	v, err := dbProvider.GetVersion()
	// check for error
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(v)
}

func TestPgSQLProvider_GetDbInfo(t *testing.T) {
	// read the config
	cfg := core.NewConfig("", "")
	// transform cfg into map
	conf, _ := NewConf(cfg.All())
	// creates the provider
	dbProvider := &PgSQLProvider{
		cfg: conf,
	}
	// test get version
	i, err := dbProvider.GetInfo()
	// check for error
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(i)
}
