// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// UpdateDirectoryCapnpClient is the POGO client to updating a directory
// It can only be obtained from the DirectoryCapnpClient
type UpdateDirectoryCapnpClient struct {
	capability hubapi.CapUpdateDirectory
}

// Release the capability and allocated resources
func (cl *UpdateDirectoryCapnpClient) Release() {
	cl.capability.Release()
}

// RemoveTD removes a TD document from the store
func (cl *UpdateDirectoryCapnpClient) RemoveTD(ctx context.Context, thingAddr string) (err error) {
	method, release := cl.capability.RemoveTD(ctx,
		func(params hubapi.CapUpdateDirectory_removeTD_Params) error {
			params.SetThingAddr(thingAddr)
			return nil
		})
	defer release()
	_, err = method.Struct()
	return err
}

// UpdateTD updates the TD document in the directory
// If the TD with the given ID doesn't exist it will be added.
func (cl *UpdateDirectoryCapnpClient) UpdateTD(ctx context.Context, thingAddr string, tdDoc []byte) (err error) {
	method, release := cl.capability.UpdateTD(ctx,
		func(params hubapi.CapUpdateDirectory_updateTD_Params) error {
			params.SetThingAddr(thingAddr)
			err = params.SetTdDoc(tdDoc)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// NewUpdateDirectoryCapnpClient returns a directory update client using the capnp protocol
func NewUpdateDirectoryCapnpClient(cap hubapi.CapUpdateDirectory) *UpdateDirectoryCapnpClient {
	cl := &UpdateDirectoryCapnpClient{capability: cap}
	return cl
}
