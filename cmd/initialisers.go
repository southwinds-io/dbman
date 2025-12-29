/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package cmd

func InitialiseRootCmd() *RootCmd {
	rootCmd := NewRootCmd()
	serveCmd := NewServeCmd()
	configCmd := InitialiseConfigCmd()
	releaseCmd := InitialiseReleaseCmd()
	dbCmd := InitialiseDbCmd()
	rootCmd.Command.AddCommand(releaseCmd.cmd, dbCmd.cmd, configCmd.cmd, serveCmd.cmd)
	return rootCmd
}

func InitialiseReleaseCmd() *ReleaseCmd {
	releaseCmd := NewReleaseCmd()
	releaseInfoCmd := NewReleaseInfoCmd()
	releasePlanCmd := NewReleasePlanCmd()
	releaseCmd.cmd.AddCommand(releaseInfoCmd.cmd, releasePlanCmd.cmd)
	return releaseCmd
}

func InitialiseDbCmd() *DbCmd {
	dbCmd := NewDbCmd()
	dbVersionCmd := NewDbVersionCmd()
	dbDiffCmd := NewDbDiffCmd()
	dbCreateCmd := NewDbCreateCmd()
	dbDeployCmd := NewDbDeployCmd()
	dbUpgradeCmd := NewDbUpgradeCmd()
	dbQueryCmd := NewDbQueryCmd()
	dbQueriesCmd := NewDbQueriesCmd()
	dbBackupCmd := NewDbBackupCmd()
	dbRestoreCmd := NewDbRestoreCmd()
	dbInfoCmd := NewDbInfoCmd()
	dbWaitCmd := NewWaitCmd()
	dbRunCmd := NewDbRunCmd()
	dbCmd.cmd.AddCommand(dbVersionCmd.cmd,
		dbDiffCmd.cmd,
		dbDeployCmd.cmd,
		dbCreateCmd.cmd,
		dbDeployCmd.cmd,
		dbRunCmd.cmd,
		dbUpgradeCmd.cmd,
		dbQueryCmd.cmd,
		dbQueriesCmd.cmd,
		dbBackupCmd.cmd,
		dbRestoreCmd.cmd,
		dbInfoCmd.cmd,
		dbWaitCmd.cmd)
	return dbCmd
}

func InitialiseConfigCmd() *ConfigCmd {
	cfgCmd := NewConfigCmd()
	cfgSetCmd := NewConfigSetCmd()
	cfgShowCmd := NewConfigShowCmd()
	cfgUseCmd := NewConfigUseCmd()
	cfgListCmd := NewConfigListCmd()
	cfgRmCmd := NewConfigDeleteCmd()
	checkCmd := NewCheckCmd()
	cfgCmd.cmd.AddCommand(cfgSetCmd.cmd, cfgShowCmd.cmd, cfgUseCmd.cmd, cfgListCmd.cmd, cfgRmCmd.cmd, checkCmd.cmd)
	return cfgCmd
}
