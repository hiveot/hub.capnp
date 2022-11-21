// Package client that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/directory"
)

// ReadDirectoryCapnpClient is the POGS client to reading a directory
// It can only be obtained from the DirectoryCapnpClient
// This implements the IReadDirectory interface
type ReadDirectoryCapnpClient struct {
	capability hubapi.CapReadDirectory
}

func (cl *ReadDirectoryCapnpClient) Cursor(
	ctx context.Context) (cursor directory.IDirectoryCursor) {

	// The use of a result 'future' avoids a round trip, making this more efficient
	getCapMethod, release := cl.capability.Cursor(ctx, nil)
	capCursor := getCapMethod.Cursor()
	capability := capCursor.AddRef()
	defer release()
	cursor = NewDirectoryCursorCapnpClient(capability)
	return cursor
}

// GetTD returns a thing value containing the TD document for the given Thing address
func (cl *ReadDirectoryCapnpClient) GetTD(ctx context.Context, thingAddr string) (tv *thing.ThingValue, err error) {

	method, release := cl.capability.GetTD(ctx,
		func(params hubapi.CapReadDirectory_getTD_Params) error {
			err2 := params.SetThingAddr(thingAddr)
			return err2
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		tvCapnp, _ := resp.Tv()
		tv = caphelp.UnmarshalThingValue(tvCapnp)
	}
	return tv, err
}

// ListTDReceiver implements the capnp 'server' for receiving callbacks.
// This is a capnp server for the client side.
// This implements the CapListCallback interface
type ListTDReceiver struct {
	pogsHandler func(batch []string, isLast bool) error
}

func (cl *ReadDirectoryCapnpClient) Release() {
	cl.capability.Release()
}
func NewReadDirectoryCapnpClient(cap hubapi.CapReadDirectory) directory.IReadDirectory {
	return &ReadDirectoryCapnpClient{
		capability: cap,
	}
}
