package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/state"
)

// ClientStateCapnpServer provides the capnp RPC server for state store
// This implements the capnproto generated interface ClientState_Server
// See hub.capnp/go/hubapi/State.capnp.go for the interface.
type ClientStateCapnpServer struct {
	srv state.IClientState
}

func (capsrv *ClientStateCapnpServer) Get(
	ctx context.Context, call hubapi.CapClientState_get) error {
	args := call.Args()
	key, _ := args.Key()
	value, err := capsrv.srv.Get(ctx, key)
	if err == nil {
		res, err := call.AllocResults()
		if err == nil {
			err = res.SetValue(value)
		}
	}
	return err
}

func (capsrv *ClientStateCapnpServer) Set(
	ctx context.Context, call hubapi.CapClientState_set) error {
	args := call.Args()
	key, _ := args.Key()
	value, _ := args.Value()
	err := capsrv.srv.Set(ctx, key, value)
	return err
}
