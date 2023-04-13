// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"capnproto.org/go/capnp/v3"
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IDirectory interface
type DirectoryCapnpClient struct {
	connection *rpc.Conn                  // connection to capnp server
	capability hubapi.CapDirectoryService // capnp client of the directory
}

// CapReadDirectory returns the capability to read the directory
// The returned release function must be called after the capability is no longer needed.
func (cl *DirectoryCapnpClient) CapReadDirectory(
	ctx context.Context, clientID string) (directory.IReadDirectory, error) {

	// The use of a result 'future' avoids a round trip, making this more efficient
	getCapMethod, getCapRelease := cl.capability.CapReadDirectory(ctx,
		func(params hubapi.CapDirectoryService_capReadDirectory_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	capRead := getCapMethod.Cap()
	defer getCapRelease()
	newCap := NewReadDirectoryCapnpClient(capRead.AddRef())
	return newCap, nil
}

// CapUpdateDirectory returns the capability to update the directory
func (cl *DirectoryCapnpClient) CapUpdateDirectory(
	ctx context.Context, clientID string) (directory.IUpdateDirectory, error) {

	// The use of a result 'future' avoids a round trip, making this more efficient
	getCapMethod, getCapRelease := cl.capability.CapUpdateDirectory(ctx,
		func(params hubapi.CapDirectoryService_capUpdateDirectory_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer getCapRelease()
	capability := getCapMethod.Cap()
	newCap := NewUpdateDirectoryCapnpClient(capability.AddRef())
	return newCap, nil
}

// Release the client capability
// Release MUST be called after use
func (cl *DirectoryCapnpClient) Release() error {
	cl.capability.Release()
	return nil
}

// NewDirectoryCapnpClientConnection returns a directory store client using the capnp protocol
//
//	ctx is the context for retrieving capabilities
//	connection is the client connection to the capnp server
func NewDirectoryCapnpClientConnection(ctx context.Context, connection net.Conn) *DirectoryCapnpClient {
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	cl := NewDirectoryCapnpClient(rpcConn.Bootstrap(ctx))
	cl.connection = rpcConn
	return cl
}

// NewDirectoryCapnpClient creates a new client for using the directory service
// The capnp client can be that of the service, the resolver or the gateway
func NewDirectoryCapnpClient(capClient capnp.Client) *DirectoryCapnpClient {
	// use a direct connection to the service
	capability := hubapi.CapDirectoryService(capClient)
	cl := &DirectoryCapnpClient{
		connection: nil,
		capability: capability,
	}
	return cl
}
