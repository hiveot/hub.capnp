package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/history"
)

// AddHistoryCapnpServer provides the capnp RPC server for adding to the history
type AddHistoryCapnpServer struct {
	svc history.IAddHistory
	// TODO: restrict to a specific device publisher
}

func (capsrv *AddHistoryCapnpServer) AddAction(
	ctx context.Context, call hubapi.CapAddHistory_addAction) error {

	args := call.Args()
	capValue, _ := args.Tv()
	thingAddr, _ := capValue.ThingAddr()
	name, _ := capValue.Name()
	valueJSON, _ := capValue.ValueJSON()
	created, _ := capValue.Created()
	actionValue := &thing.ThingValue{
		ThingAddr: thingAddr,
		Name:      name,
		ValueJSON: valueJSON,
		Created:   created,
	}

	err := capsrv.svc.AddAction(ctx, actionValue)
	return err
}

func (capsrv *AddHistoryCapnpServer) AddEvent(
	ctx context.Context, call hubapi.CapAddHistory_addEvent) error {

	args := call.Args()
	capValue, _ := args.Tv()
	thingAddr, _ := capValue.ThingAddr()
	name, _ := capValue.Name()
	valueJSON, _ := capValue.ValueJSON()
	created, _ := capValue.Created()
	eventValue := &thing.ThingValue{
		ThingAddr: thingAddr,
		Name:      name,
		ValueJSON: valueJSON,
		Created:   created,
	}
	err := capsrv.svc.AddEvent(ctx, eventValue)
	//call.Ack()
	return err
}
func (capsrv *AddHistoryCapnpServer) AddEvents(
	ctx context.Context, call hubapi.CapAddHistory_addEvents) error {

	args := call.Args()
	capValues, _ := args.Tv()
	eventValues := caphelp.UnmarshalThingValueList(capValues)
	err := capsrv.svc.AddEvents(ctx, eventValues)
	return err
}

func (capsrv *AddHistoryCapnpServer) Shutdown() {
	// Release on the client calls capnp release
	// Pass this to the server to cleanup
	capsrv.svc.Release()
}
