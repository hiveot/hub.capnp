// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"
	"time"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// StateCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IState interface
type StateCapnpClient struct {
	connection *rpc.Conn       // connection to capnp server
	capability hubapi.CapState // capnp client of the state store
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

// Get reads the state
func (cl *StateCapnpClient) Get(ctx context.Context, key string) (string, error) {
	var err error
	var val string

	method, release := cl.capability.Get(ctx,
		func(params hubapi.CapState_get_Params) error {
			err = params.SetKey(key)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		val, err = resp.Value()
	}
	return val, err
}

// Set reads the state
func (cl *StateCapnpClient) Set(ctx context.Context, key string, value string) error {
	var err error
	method, release := cl.capability.Set(ctx,
		func(params hubapi.CapState_set_Params) error {
			err = params.SetKey(key)
			_ = params.SetValue(value)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// NewStateCapnpClient returns a state store client using the capnp protocol
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
