// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"
	"time"

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
	ctxCancel  context.CancelFunc
}

// CapReadHistory the capability to read the history
func (cl *HistoryCapnpClient) CapReadHistory() history.IReadHistory {
	getCap, _ := cl.capability.CapReadHistory(cl.ctx, nil)
	capability := getCap.Cap()
	return NewReadHistoryCapnpClient(capability)
}

// CapUpdateHistory provides the capability to update the history
func (cl *HistoryCapnpClient) CapUpdateHistory() history.IUpdateHistory {
	// The use of a result 'future' avoids a round trip, making this more efficient
	getCap, _ := cl.capability.CapUpdateHistory(cl.ctx, nil)
	capability := getCap.Cap()

	return NewUpdateHistoryCapnpClient(capability)
}

// NewHistoryCapnpClient returns a history store client using the capnp protocol
func NewHistoryCapnpClient(address string, isUDS bool) (*HistoryCapnpClient, error) {
	var cl *HistoryCapnpClient
	network := "tcp"
	if isUDS {
		network = "unix"
	}
	connection, err := net.Dial(network, address)
	if err == nil {
		transport := rpc.NewStreamTransport(connection)
		rpcConn := rpc.NewConn(transport, nil)
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*60)
		capability := hubapi.CapHistory(rpcConn.Bootstrap(ctx))

		cl = &HistoryCapnpClient{
			connection: rpcConn,
			capability: capability,
			ctx:        ctx,
			ctxCancel:  ctxCancel,
		}
	}
	return cl, nil
}
