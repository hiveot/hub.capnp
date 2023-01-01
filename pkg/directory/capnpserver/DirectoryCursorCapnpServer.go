package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryCursorCapnpServer is a capnproto RPC server for reading of the directory store
type DirectoryCursorCapnpServer struct {
	svc directory.IDirectoryCursor
}

func (capsrv DirectoryCursorCapnpServer) First(
	_ context.Context, call hubapi.CapDirectoryCursor_first) error {

	thingValue, valid := capsrv.svc.First()
	res, err := call.AllocResults()
	if err == nil {
		thingValueCapnp := caphelp.MarshalThingValue(thingValue)
		res.SetValid(valid)
		err = res.SetTv(thingValueCapnp)
	}
	return err
}

func (capsrv DirectoryCursorCapnpServer) Next(
	_ context.Context, call hubapi.CapDirectoryCursor_next) error {
	thingValue, valid := capsrv.svc.Next()
	res, err := call.AllocResults()
	if err == nil {
		res.SetValid(valid)
		thingValueCapnp := caphelp.MarshalThingValue(thingValue)
		err = res.SetTv(thingValueCapnp)
	}
	return err
}

func (capsrv DirectoryCursorCapnpServer) NextN(
	_ context.Context, call hubapi.CapDirectoryCursor_nextN) error {
	args := call.Args()
	steps := args.Steps()
	thingValueList, valid := capsrv.svc.NextN(uint(steps))
	res, err := call.AllocResults()
	if err == nil {
		res.SetValid(valid)
		thingValueListCapnp := caphelp.MarshalThingValueList(thingValueList)
		err = res.SetBatch(thingValueListCapnp)
	}
	return err
}

func (capsrv *DirectoryCursorCapnpServer) Shutdown() {
	// Release on the client calls capnp Shutdown.
	// Pass this to the server to cleanup
	capsrv.svc.Release()
}

func NewDirectoryCursorCapnpServer(cursor directory.IDirectoryCursor) *DirectoryCursorCapnpServer {
	cursorCapnpServer := &DirectoryCursorCapnpServer{
		svc: cursor,
	}
	return cursorCapnpServer
}
