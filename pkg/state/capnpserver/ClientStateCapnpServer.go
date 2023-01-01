package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
	"github.com/hiveot/hub/pkg/bucketstore/capnpserver"
	"github.com/hiveot/hub/pkg/state"
)

// ClientStateCapnpServer provides the capnp RPC server for state store
// This implements the capnproto generated interface ClientState_Server
// See hub.capnp/go/hubapi/State.capnp.go for the interface.
type ClientStateCapnpServer struct {
	srv state.IClientState
}

// Cursor returns the capability to iterate the bucket to the client
func (capsrv *ClientStateCapnpServer) Cursor(
	ctx context.Context, call hubapi.CapClientState_cursor) error {

	// first create the instance of the POGS server for this client application
	pogoCursor := capsrv.srv.Cursor(ctx)

	// second, wrap it in a capnp binding which implements the capnp generated cursor API
	bucketCursorCapnpServer := capnpserver.NewBucketCursorCapnpServer(pogoCursor)
	// create the capnp RPC server for this capability
	capability := hubapi.CapBucketCursor_ServerToClient(bucketCursorCapnpServer)
	// and return the result to the caller
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}

	return err
}

func (capsrv *ClientStateCapnpServer) Delete(
	ctx context.Context, call hubapi.CapClientState_delete) error {
	args := call.Args()
	key, _ := args.Key()
	err := capsrv.srv.Delete(ctx, key)
	return err
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
func (capsrv *ClientStateCapnpServer) GetMultiple(
	ctx context.Context, call hubapi.CapClientState_getMultiple) error {
	args := call.Args()
	keysCapnp, _ := args.Keys()
	keys := caphelp.UnmarshalStringList(keysCapnp)
	docs, err := capsrv.srv.GetMultiple(ctx, keys)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		if err == nil {
			kvmapCapnp := caphelp.MarshalKeyValueMap(docs)
			err = res.SetDocs(kvmapCapnp)
		}
	}
	return err
}

func (capsrv *ClientStateCapnpServer) Set(
	ctx context.Context, call hubapi.CapClientState_set) error {
	args := call.Args()
	key, _ := args.Key()
	value, _ := args.Value() // value points to capnp buffer so copy it
	err := capsrv.srv.Set(ctx, key, value)
	return err
}

func (capsrv *ClientStateCapnpServer) SetMultiple(
	ctx context.Context, call hubapi.CapClientState_setMultiple) error {
	args := call.Args()
	kvmapCapnp, err := args.Docs()
	if err == nil {
		docs := caphelp.UnmarshalKeyValueMap(kvmapCapnp)
		err = capsrv.srv.SetMultiple(ctx, docs)
	}
	return err
}

func (capsrv *ClientStateCapnpServer) Shutdown() {
	// Release on the client calls capnp Shutdown.
	// Pass this to the server to cleanup
	capsrv.srv.Release()
}
