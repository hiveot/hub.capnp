package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/history"
)

// ReadHistoryCapnpServer is a capnproto RPC server for reading of the history store
type ReadHistoryCapnpServer struct {
	srv history.IReadHistory
}

func (capsrv *ReadHistoryCapnpServer) GetActionHistory(
	ctx context.Context, call hubapi.CapReadHistory_getActionHistory) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	name, _ := args.ActionName()
	after, _ := args.After()
	before, _ := args.Before()
	limit := args.Limit()
	hist, err := capsrv.srv.GetActionHistory(ctx, thingID, name, after, before, int(limit))
	if err == nil {
		res, _ := call.AllocResults()
		valList := caphelp.ThingValueListPOGS2Capnp(hist)
		res.SetValues(valList)
	}
	return err
}

func (capsrv *ReadHistoryCapnpServer) GetEventHistory(
	ctx context.Context, call hubapi.CapReadHistory_getEventHistory) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	name, _ := args.EventName()
	after, _ := args.After()
	before, _ := args.Before()
	limit := args.Limit()

	hist, err := capsrv.srv.GetEventHistory(ctx, thingID, name, after, before, int(limit))
	if err == nil {
		res, _ := call.AllocResults()
		valList := caphelp.ThingValueListPOGS2Capnp(hist)
		res.SetValues(valList)
	}
	return err
}

func (capsrv *ReadHistoryCapnpServer) GetLatestEvents(
	ctx context.Context, call hubapi.CapReadHistory_getLatestEvents) error {

	args := call.Args()
	thingID, _ := args.ThingID()

	hist, err := capsrv.srv.GetLatestEvents(ctx, thingID)
	if err == nil {
		res, _ := call.AllocResults()
		capMap := caphelp.ThingValueMapPOGS2ToCapnp(hist)
		res.SetThingValueMap(capMap)
	}
	return err
}

func (capsrv *ReadHistoryCapnpServer) Info(
	ctx context.Context, call hubapi.CapReadHistory_info) (err error) {

	inf, err := capsrv.srv.Info(ctx)
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

// NewReadHistoryCapnpServer creates an instance of the capnp server for reading history
func NewReadHistoryCapnpServer(srv history.IReadHistory) *ReadHistoryCapnpServer {
	capsrv := &ReadHistoryCapnpServer{srv: srv}
	return capsrv
}
