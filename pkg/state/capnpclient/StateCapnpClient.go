// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/state"
)

// StateCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IState interface
type StateCapnpClient struct {
	connection *rpc.Conn       // connection to capnp server
	capability hubapi.CapState // capnp client of the state store
	ctx        context.Context
}

func (cl *StateCapnpClient) CapClientState(ctx context.Context, clientID string, appID string) state.IClientState {
	getCap, release := cl.capability.CapClientState(ctx,
		func(params hubapi.CapState_capClientState_Params) error {
			err2 := params.SetClientID(clientID)
			_ = params.SetAppID(appID)
			return err2
		})
	defer release()
	capability := getCap.Cap()
	return NewClientStateCapnpClient(capability.AddRef())
}

// NewStateCapnpClient returns a state store client using the capnp protocol
//  ctx is the context for retrieving capabilities
//  connection is the client connection to the capnp RPC server
func NewStateCapnpClient(ctx context.Context, connection net.Conn) (*StateCapnpClient, error) {
	var cl *StateCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapState(rpcConn.Bootstrap(ctx))

	cl = &StateCapnpClient{
		connection: rpcConn,
		capability: capability,
		ctx:        ctx,
	}
	return cl, nil
}
