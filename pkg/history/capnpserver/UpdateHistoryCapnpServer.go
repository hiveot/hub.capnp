package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/history"
)

// UpdateHistoryCapnpServer provides the capnp RPC server for adding to the history
type UpdateHistoryCapnpServer struct {
	srv history.IUpdateHistory
	// TODO: restrict to a specific device publisher
}

func (capsrv *UpdateHistoryCapnpServer) AddAction(
	ctx context.Context, call hubapi.CapUpdateHistory_addAction) error {

	args := call.Args()
	capValue, _ := args.ActionValue()
	thingID, _ := capValue.ThingID()
	name, _ := capValue.Name()
	valueJSON, _ := capValue.ValueJSON()
	created, _ := capValue.Created()
	actionValue := thing.ThingValue{
		ThingID:   thingID,
		Name:      name,
		ValueJSON: valueJSON,
		Created:   created,
	}

	err := capsrv.srv.AddAction(ctx, actionValue)
	return err
}

func (capsrv *UpdateHistoryCapnpServer) AddEvent(
	ctx context.Context, call hubapi.CapUpdateHistory_addEvent) error {

	args := call.Args()
	capValue, _ := args.EventValue()
	thingID, _ := capValue.ThingID()
	name, _ := capValue.Name()
	valueJSON, _ := capValue.ValueJSON()
	created, _ := capValue.Created()
	eventValue := thing.ThingValue{
		ThingID:   thingID,
		Name:      name,
		ValueJSON: valueJSON,
		Created:   created,
	}
	err := capsrv.srv.AddEvent(ctx, eventValue)
	//call.Ack()
	return err
}
func (capsrv *UpdateHistoryCapnpServer) AddEvents(
	ctx context.Context, call hubapi.CapUpdateHistory_addEvents) error {

	args := call.Args()
	capValues, _ := args.EventValues()
	eventValues := caphelp.ThingValueListCapnp2POGS(capValues)
	err := capsrv.srv.AddEvents(ctx, eventValues)
	return err
}
