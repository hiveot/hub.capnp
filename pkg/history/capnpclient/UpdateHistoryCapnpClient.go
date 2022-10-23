// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
)

// UpdateHistoryCapnpClient provides a POGS wrapper around the capnp client API
// This implements the IUpdateHistory interface
type UpdateHistoryCapnpClient struct {
	capability hubapi.CapUpdateHistory // capnp client
}

// AddAction adds a Thing action with the given name and value to the action history
// TODO: split this into get capability and add action
func (cl *UpdateHistoryCapnpClient) AddAction(ctx context.Context, actionValue thing.ThingValue) error {

	// next add the action
	method, release := cl.capability.AddAction(ctx,
		func(params hubapi.CapUpdateHistory_addAction_Params) error {
			capValue := caphelp.MarshalThingValue(actionValue)
			err2 := params.SetActionValue(capValue)
			return err2
		})
	defer release()
	_, err := method.Struct()
	return err
}

// AddEvent adds an event to the event history
func (cl *UpdateHistoryCapnpClient) AddEvent(
	ctx context.Context, eventValue thing.ThingValue) error {

	method, release := cl.capability.AddEvent(ctx,
		func(params hubapi.CapUpdateHistory_addEvent_Params) error {
			capValue := caphelp.MarshalThingValue(eventValue)
			err2 := params.SetEventValue(capValue)
			return err2
		})
	defer release()
	_, err := method.Struct()
	return err
}

func (cl *UpdateHistoryCapnpClient) AddEvents(
	ctx context.Context, events []thing.ThingValue) error {

	method, release := cl.capability.AddEvents(ctx,
		func(params hubapi.CapUpdateHistory_addEvents_Params) error {
			// suspect that this conversion is slow
			capValues := caphelp.MarshalThingValueList(events)
			err2 := params.SetEventValues(capValues)
			return err2
		})
	defer release()
	_, err := method.Struct()
	return err
}

// NewUpdateHistoryCapnpClient returns an update-history client using the capnp protocol
// Intended for internal use.
func NewUpdateHistoryCapnpClient(cap hubapi.CapUpdateHistory) *UpdateHistoryCapnpClient {
	cl := &UpdateHistoryCapnpClient{capability: cap}
	return cl
}
