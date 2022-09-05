/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package plugin

// The RPC server that the DatabaseProviderRPC client talks to, conforming to
// the requirements of net/rpc
type DatabaseProviderRPCServer struct {
	// This is the real implementation
	Impl DatabaseProvider
}

func (s *DatabaseProviderRPCServer) Setup(args string, resp *string) error {
	*resp = s.Impl.Setup(args)
	return nil
}

func (s *DatabaseProviderRPCServer) GetInfo(args string, resp *string) error {
	*resp = s.Impl.GetInfo()
	return nil
}

func (s *DatabaseProviderRPCServer) GetVersion(args string, resp *string) error {
	*resp = s.Impl.GetVersion()
	return nil
}

func (s *DatabaseProviderRPCServer) SetVersion(args string, resp *string) error {
	*resp = s.Impl.SetVersion(args)
	return nil
}

func (s *DatabaseProviderRPCServer) RunCommand(args string, resp *string) error {
	*resp = s.Impl.RunCommand(args)
	return nil
}

func (s *DatabaseProviderRPCServer) RunQuery(args string, resp *string) error {
	*resp = s.Impl.RunQuery(args)
	return nil
}
