package adapter

import (
	"context"
	"net"
	"time"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/historystore/client"
	"github.com/hiveot/hub/pkg/historystore/mongohs"
)

// HistoryStoreCapnpAdapter is a capnproto adapter for the history store
// This implements the capnproto generated interface HistoryStore_Server
// See hub.capnp/go/hubapi/HistoryStore.capnp.go for the interface.
type HistoryStoreCapnpAdapter struct {
	srv client.IHistoryStore
}

func (adpt *HistoryStoreCapnpAdapter) AddAction(
	ctx context.Context, call hubapi.HistoryStore_addAction) error {
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

	err := adpt.srv.AddAction(ctx, actionValue)
	return err
}

func (adpt *HistoryStoreCapnpAdapter) AddEvent(
	ctx context.Context, call hubapi.HistoryStore_addEvent) error {
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
	err := adpt.srv.AddEvent(ctx, eventValue)
	return err
}
func (adpt *HistoryStoreCapnpAdapter) AddEvents(
	ctx context.Context, call hubapi.HistoryStore_addEvents) error {
	args := call.Args()
	capValues, _ := args.EventValues()
	eventValues := caphelp.CapnpToThingValueList(capValues)
	err := adpt.srv.AddEvents(ctx, eventValues)
	return err
}

func (adpt *HistoryStoreCapnpAdapter) GetActionHistory(
	ctx context.Context, call hubapi.HistoryStore_getActionHistory) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	name, _ := args.ActionName()
	after, _ := args.After()
	before, _ := args.Before()
	limit := args.Limit()
	hist, err := adpt.srv.GetActionHistory(ctx, thingID, name, after, before, int(limit))
	if err == nil {
		res, _ := call.AllocResults()
		valList := caphelp.ThingValueListToCapnp(hist)
		res.SetValues(valList)
	}
	return err
}

func (adpt *HistoryStoreCapnpAdapter) GetEventHistory(
	ctx context.Context, call hubapi.HistoryStore_getEventHistory) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	name, _ := args.EventName()
	after, _ := args.After()
	before, _ := args.Before()
	limit := args.Limit()

	hist, err := adpt.srv.GetEventHistory(ctx, thingID, name, after, before, int(limit))
	if err == nil {
		res, _ := call.AllocResults()
		valList := caphelp.ThingValueListToCapnp(hist)
		res.SetValues(valList)
	}
	return err
}

func (adpt *HistoryStoreCapnpAdapter) GetLatestEvents(
	ctx context.Context, call hubapi.HistoryStore_getLatestEvents) error {

	args := call.Args()
	thingID, _ := args.ThingID()

	hist, err := adpt.srv.GetLatestEvents(ctx, thingID)
	if err == nil {
		res, _ := call.AllocResults()
		capMap := caphelp.ThingValueMapToCapnp(hist)
		res.SetThingValueMap(capMap)
	}
	return err
}

func (adpt *HistoryStoreCapnpAdapter) Info(
	ctx context.Context, call hubapi.HistoryStore_info) (err error) {

	inf, err := adpt.srv.Info(ctx)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		storeInfo, _ := res.NewStatistics()
		storeInfo.SetNrActions(int64(inf.NrActions))
		storeInfo.SetNrEvents(int64(inf.NrEvents))
		storeInfo.SetEngine(inf.Engine)
		storeInfo.SetUptime(int64(inf.Uptime))
	}

	return err
}

// StartHistoryStoreCapnpAdapter starts the history store capnp protocol server
func StartHistoryStoreCapnpAdapter(listener net.Listener,
	srv *mongohs.MongoHistoryStoreServer) error {

	adpt := &HistoryStoreCapnpAdapter{
		srv: srv,
	}
	// Create the capnp client to receive requests
	main := hubapi.HistoryStore_ServerToClient(adpt)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)

	return caphelp.CapServe(ctx, listener, capnp.Client(main))
}
