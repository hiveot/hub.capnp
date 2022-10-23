// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IDirectory interface
type DirectoryCapnpClient struct {
	connection *rpc.Conn           // connection to capnp server
	capability hubapi.CapDirectory // capnp client of the directory
}

// CapReadDirectory returns the capability to read the directory
func (cl *DirectoryCapnpClient) CapReadDirectory(ctx context.Context) directory.IReadDirectory {
	// The use of a result 'future' avoids a round trip, making this more efficient
	getCap, _ := cl.capability.CapReadDirectory(ctx, nil)
	capability := getCap.Cap()
	return NewReadDirectoryCapnpClient(capability)
}

// CapUpdateDirectory returns the capability to update the directory
func (cl *DirectoryCapnpClient) CapUpdateDirectory(ctx context.Context) directory.IUpdateDirectory {
	// The use of a result 'future' avoids a round trip, making this more efficient
	getCap, _ := cl.capability.CapUpdateDirectory(ctx, nil)
	capability := getCap.Cap()
	return NewUpdateDirectoryCapnpClient(capability)
}

// NewDirectoryCapnpClient returns a directory store client using the capnp protocol
//  ctx is the context for retrieving capabilities
//  connection is the client connection to the capnp server
func NewDirectoryCapnpClient(ctx context.Context, connection net.Conn) (*DirectoryCapnpClient, error) {
	var cl *DirectoryCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapDirectory(rpcConn.Bootstrap(ctx))

	cl = &DirectoryCapnpClient{
		connection: rpcConn,
		capability: capability,
	}
	return cl, nil
}
