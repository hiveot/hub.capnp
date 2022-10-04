// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"
	"time"

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
	ctxCancel  context.CancelFunc
}

func (cl *StateCapnpClient) CapClientState(ctx context.Context, clientID string, appID string) state.IClientState {
	getCap, _ := cl.capability.CapClientState(ctx,
		func(params hubapi.CapState_capClientState_Params) error {
			err2 := params.SetClientID(clientID)
			_ = params.SetAppID(appID)
			return err2
		})
	capability := getCap.Cap()
	return NewClientStateCapnpClient(capability)
}

// NewStateCapnpClient returns a state store client using the capnp protocol
// Intended for bootstrapping the capability chain
func NewStateCapnpClient(address string, isUDS bool) (*StateCapnpClient, error) {
	var cl *StateCapnpClient
	network := "tcp"
	if isUDS {
		network = "unix"
	}
	connection, err := net.Dial(network, address)
	if err == nil {
		transport := rpc.NewStreamTransport(connection)
		rpcConn := rpc.NewConn(transport, nil)
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*60)
		capability := hubapi.CapState(rpcConn.Bootstrap(ctx))

		cl = &StateCapnpClient{
			connection: rpcConn,
			capability: capability,
			ctx:        ctx,
			ctxCancel:  ctxCancel,
		}
	}
	return cl, nil
}
