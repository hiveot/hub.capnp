// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// ClientStateCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IClientState interface
type ClientStateCapnpClient struct {
	capability hubapi.CapClientState // capnp client of the state store
}

func (cl *ClientStateCapnpClient) Release() {
	cl.capability.Release()
}

// Get reads the state
func (cl *ClientStateCapnpClient) Get(ctx context.Context, key string) (string, error) {
	var err error
	var val string

	method, release := cl.capability.Get(ctx,
		func(params hubapi.CapClientState_get_Params) error {
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
func (cl *ClientStateCapnpClient) Set(ctx context.Context, key string, value string) error {
	var err error
	method, release := cl.capability.Set(ctx,
		func(params hubapi.CapClientState_set_Params) error {
			err = params.SetKey(key)
			_ = params.SetValue(value)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// NewClientStateCapnpClient returns the capability to store client application state over capnp RPC
func NewClientStateCapnpClient(capability hubapi.CapClientState) *ClientStateCapnpClient {
	cl := &ClientStateCapnpClient{
		capability: capability,
	}
	return cl
}
