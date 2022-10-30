// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/history"
)

// HistoryCapnpClient provides a POGS wrapper around the capnp client API
// This implements the IHistory interface
type HistoryCapnpClient struct {
	connection *rpc.Conn         // connection to capnp server
	capability hubapi.CapHistory // capnp client
	ctx        context.Context
}

// CapReadHistory the capability to read the history
func (cl *HistoryCapnpClient) CapReadHistory() history.IReadHistory {
	getCap, release := cl.capability.CapReadHistory(cl.ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewReadHistoryCapnpClient(capability)
}

// CapUpdateHistory provides the capability to update the history
func (cl *HistoryCapnpClient) CapUpdateHistory() history.IUpdateHistory {
	// The use of a result 'future' avoids a round trip, making this more efficient
	getCap, release := cl.capability.CapUpdateHistory(cl.ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()

	return NewUpdateHistoryCapnpClient(capability)
}

// NewHistoryCapnpClient returns a history store client using the capnp protocol
//  ctx is the context for getting capabilities from the server
//  connection is the connection to the capnp server
func NewHistoryCapnpClient(ctx context.Context, connection net.Conn) (*HistoryCapnpClient, error) {
	var cl *HistoryCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapHistory(rpcConn.Bootstrap(ctx))

	cl = &HistoryCapnpClient{
		connection: rpcConn,
		capability: capability,
		ctx:        ctx,
	}
	return cl, nil
}
