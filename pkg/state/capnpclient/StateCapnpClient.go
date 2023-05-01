// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/state"
)

// StateCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IStateService interface
type StateCapnpClient struct {
	connection *rpc.Conn       // connection to capnp server
	capability hubapi.CapState // capnp client of the state store
}

func (cl *StateCapnpClient) CapClientState(
	ctx context.Context, clientID string, appID string) (state.IClientState, error) {

	method, release := cl.capability.CapClientState(ctx,
		func(params hubapi.CapState_capClientState_Params) error {
			err2 := params.SetClientID(clientID)
			_ = params.SetAppID(appID)
			return err2
		})
	defer release()
	capability := method.Cap()
	newCap := NewClientStateCapnpClient(capability.AddRef())
	return newCap, nil
}

func (cl *StateCapnpClient) Release() {
	// release will release  client service instance
	cl.capability.Release()
}

// NewStateCapnpClient returns a state store client using the capnp protocol
func NewStateCapnpClient(client capnp.Client) *StateCapnpClient {
	capability := hubapi.CapState(client)

	cl := &StateCapnpClient{
		connection: nil,
		capability: capability,
	}
	return cl
}
