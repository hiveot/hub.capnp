package adapter

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/historystore/mongohs"
)

// HistoryStoreCapnpAdapter is a capnproto adapter for the history store
// This implements the capnproto generated interface HistoryStore_Server
// See hub.capnp/go/hubapi/History.capnp.go for the interface.
type HistoryStoreCapnpAdapter struct {
	srv *mongohs.MongoHistoryStoreServer
}

func (adpt *HistoryStoreCapnpAdapter) AddAction(
	ctx context.Context, call hubapi.HistoryStore_addAction) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	name, _ := args.Name()
	valueJSON, _ := args.ValueJSON()
	created, _ := args.Created()
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
	thingID, _ := args.ThingID()
	name, _ := args.Name()
	valueJSON, _ := args.ValueJSON()
	created, _ := args.Created()
	eventValue := thing.ThingValue{
		ThingID:   thingID,
		Name:      name,
		ValueJSON: valueJSON,
		Created:   created,
	}
	err := adpt.srv.AddEvent(ctx, eventValue)
	return err
}

func (adpt *HistoryStoreCapnpAdapter) GetActionHistory(
	ctx context.Context, call hubapi.HistoryStore_getActionHistory) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	name, _ := args.ActionName()
	after, _ := args.After()
	before, _ := args.Before()
	hist, err := adpt.srv.GetActionHistory(ctx, thingID, name, after, before)
	if err == nil {
		res, _ := call.AllocResults()
		valList := caphelp.ToCapnpValueList(hist)
		res.SetValues(valList)

		//valList, _ := res.Values()
		//for i := 0; i < len(hist); i++ {
		//	histValue := hist[i]
		//	capValue := valList.At(i)
		//	capValue.SetName(histValue.Name)
		//	capValue.SetValueJSON(histValue.ValueJSON)
		//	capValue.SetCreated(histValue.Created)
		//}
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

	hist, err := adpt.srv.GetEventHistory(ctx, thingID, name, after, before)
	if err == nil {
		res, _ := call.AllocResults()
		valList, _ := res.Values()
		for i := 0; i < len(hist); i++ {
			histValue := hist[i]
			capValue := valList.At(i)
			capValue.SetName(histValue.Name)
			capValue.SetValueJSON(histValue.ValueJSON)
			capValue.SetCreated(histValue.Created)
		}
	}
	return err
}

// TODO
func (adpt *HistoryStoreCapnpAdapter) Info(
	ctx context.Context, call hubapi.HistoryStore_info) error {
	//todo
	//err := adpt.srv.Info(ctx)
	return fmt.Errorf("Not implemented")
}

// StartHistoryStoreCapnpAdapter starts the history store capnp protocol server
func StartHistoryStoreCapnpAdapter(ctx context.Context,
	listener net.Listener,
	srv *mongohs.MongoHistoryStoreServer) error {

	adpt := &HistoryStoreCapnpAdapter{
		srv: srv,
	}
	// Create the capnp client to receive requests
	main := hubapi.HistoryStore_ServerToClient(adpt)

	return caphelp.CapServe(ctx, listener, capnp.Client(main))
}
