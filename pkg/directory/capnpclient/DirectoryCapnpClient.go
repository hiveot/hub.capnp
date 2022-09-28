// Package client that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"
	"time"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IDirectory interface
type DirectoryCapnpClient struct {
	connection *rpc.Conn           // connection to capnp server
	capability hubapi.CapDirectory // capnp client of the directory
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

// CapReadDirectory returns the capability to read the directory
func (cl *DirectoryCapnpClient) CapReadDirectory() directory.IReadDirectory {
	// The use of a result 'future' avoids a round trip, making this more efficient
	getCap, _ := cl.capability.CapReadDirectory(cl.ctx, nil)
	capability := getCap.Cap()
	return NewReadDirectoryCapnpClient(capability)
}

// CapUpdateDirectory returns the capability to update the directory
func (cl *DirectoryCapnpClient) CapUpdateDirectory() directory.IUpdateDirectory {
	// The use of a result 'future' avoids a round trip, making this more efficient
	getCap, _ := cl.capability.CapUpdateDirectory(cl.ctx, nil)
	cap := getCap.Cap()
	//res, _ := getCap.Struct()
	//cap := res.Cap()

	return NewUpdateDirectoryCapnpClient(cap)
}

// NewDirectoryStoreCapnpClient returns a directory store client using the capnp protocol
func NewDirectoryStoreCapnpClient(address string, isUDS bool) (*DirectoryCapnpClient, error) {
	var cl *DirectoryCapnpClient
	network := "tcp"
	if isUDS {
		network = "unix"
	}
	connection, err := net.Dial(network, address)
	if err == nil {
		transport := rpc.NewStreamTransport(connection)
		rpcConn := rpc.NewConn(transport, nil)
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*60)
		capability := hubapi.CapDirectory(rpcConn.Bootstrap(ctx))

		cl = &DirectoryCapnpClient{
			connection: rpcConn,
			capability: capability,
			ctx:        ctx,
			ctxCancel:  ctxCancel,
		}
	}
	return cl, nil
}
