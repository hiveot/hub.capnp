// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

// ClientStateCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IClientState interface
type ClientStateCapnpClient struct {
	capability hubapi.CapClientState // capnp client of the state store
}

func (cl *ClientStateCapnpClient) Release() {
	cl.capability.Release()
}

func (cl *ClientStateCapnpClient) Delete(ctx context.Context, key string) (err error) {
	method, release := cl.capability.Delete(ctx,
		func(params hubapi.CapClientState_delete_Params) error {
			err = params.SetKey(key)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
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

func (cl *ClientStateCapnpClient) GetMultiple(
	ctx context.Context, keys []string) (docs map[string]string, err error) {

	method, release := cl.capability.GetMultiple(ctx,
		func(params hubapi.CapClientState_getMultiple_Params) error {
			keyListCapnp := caphelp.MarshalStringList(keys)
			err = params.SetKeys(keyListCapnp)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		kvMapCapnp, err2 := resp.Docs()
		err = err2
		if err == nil {
			docs = caphelp.UnmarshalKeyValueMap(kvMapCapnp)
		}
	}
	return docs, err
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

// SetMultiple sets a batch of key-value state
func (cl *ClientStateCapnpClient) SetMultiple(ctx context.Context, docs map[string]string) error {
	var err error
	method, release := cl.capability.SetMultiple(ctx,
		func(params hubapi.CapClientState_setMultiple_Params) error {
			docsCapnp := caphelp.MarshalKeyValueMap(docs)
			err = params.SetDocs(docsCapnp)
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
