// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
)

// DirectoryCursorCapnpClient provides a POGS wrapper around the capnp client API
// This implements the IDirectoryCursor interface
type DirectoryCursorCapnpClient struct {
	capability hubapi.CapDirectoryCursor // capnp client
}

// First positions the cursor at the first key in the ordered list
func (cl *DirectoryCursorCapnpClient) First() (thingValue *thing.ThingValue, valid bool) {
	ctx := context.Background()
	method, release := cl.capability.First(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		tvCapnp, _ := resp.Tv()
		valid = resp.Valid()
		thingValue = caphelp.UnmarshalThingValue(tvCapnp)
	}
	return thingValue, valid
}

// Next moves the cursor to the next key from the current cursor
func (cl *DirectoryCursorCapnpClient) Next() (thingValue *thing.ThingValue, valid bool) {
	ctx := context.Background()
	method, release := cl.capability.Next(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		tvCapnp, _ := resp.Tv()
		valid = resp.Valid()
		thingValue = caphelp.UnmarshalThingValue(tvCapnp)
	}
	return thingValue, valid
}

// NextN moves the cursor to the next N steps from the current cursor
func (cl *DirectoryCursorCapnpClient) NextN(steps uint) (batch []*thing.ThingValue, valid bool) {
	ctx := context.Background()

	method, release := cl.capability.NextN(ctx,
		func(params hubapi.CapDirectoryCursor_nextN_Params) error {
			params.SetSteps(uint32(steps))
			return nil
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		valid = resp.Valid()
		thingValueListCap, _ := resp.Batch()
		batch = caphelp.UnmarshalThingValueList(thingValueListCap)
	}
	return batch, valid
}

// Release the cursor capability
func (cl *DirectoryCursorCapnpClient) Release() {
	logrus.Infof("releasing cursor")
	cl.capability.Release()
}

// NewDirectoryCursorCapnpClient returns a read directory client using the capnp protocol
// Intended for internal use.
func NewDirectoryCursorCapnpClient(cursorCapnp hubapi.CapDirectoryCursor) *DirectoryCursorCapnpClient {
	cl := &DirectoryCursorCapnpClient{capability: cursorCapnp}
	return cl
}
