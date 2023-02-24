package capnpserver

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/directory"
)

// UpdateDirectoryCapnpServer provides the capnp RPC server for updating the directory
// This implements the capnproto generated interface UpdateDirectory_Server
// TODO: option to restrict capability to a specific deviceID or publisher
type UpdateDirectoryCapnpServer struct {
	srv directory.IUpdateDirectory
}

func (capsrv *UpdateDirectoryCapnpServer) RemoveTD(
	ctx context.Context, call hubapi.CapUpdateDirectory_removeTD) (err error) {

	args := call.Args()
	publisherID, _ := args.PublisherID()
	thingID, _ := args.ThingID()
	err = capsrv.srv.RemoveTD(ctx, publisherID, thingID)
	return err
}

func (capsrv *UpdateDirectoryCapnpServer) Shutdown() {
	// Release on the client calls capnp Shutdown.
	// Pass this to the server to cleanup
	capsrv.srv.Release()
}

func (capsrv *UpdateDirectoryCapnpServer) UpdateTD(
	ctx context.Context, call hubapi.CapUpdateDirectory_updateTD) (err error) {

	args := call.Args()
	publisherID, _ := args.PublisherID()
	thingID, _ := args.ThingID()
	tdDoc, _ := args.TdDoc()
	err = capsrv.srv.UpdateTD(ctx, publisherID, thingID, tdDoc)
	return err
}
