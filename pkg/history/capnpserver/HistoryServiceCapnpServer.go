package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/resolver/client"
)

// HistoryServiceCapnpServer is a capnproto adapter for the history store
// This implements the capnproto generated interface History_Server
// See hub.capnp/go/hubapi/HistoryStore.capnp.go for the interface.
type HistoryServiceCapnpServer struct {
	capRegSrv *client.CapRegistrationServer
	svc       history.IHistoryService
}

func (capsrv *HistoryServiceCapnpServer) CapAddHistory(
	ctx context.Context, call hubapi.CapHistoryService_capAddHistory) error {
	// create a client instance for adding history
	args := call.Args()
	thingAddr, _ := args.ThingAddr()
	ahCapSrv := &AddHistoryCapnpServer{
		svc: capsrv.svc.CapAddHistory(ctx, thingAddr),
	}

	capnpAddHistory := hubapi.CapAddHistory_ServerToClient(ahCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capnpAddHistory)
	}
	return err
}

func (capsrv *HistoryServiceCapnpServer) CapAddAnyThing(
	ctx context.Context, call hubapi.CapHistoryService_capAddAnyThing) error {
	// create a client instance for adding history
	ahCapSrv := &AddHistoryCapnpServer{
		svc: capsrv.svc.CapAddAnyThing(ctx),
	}
	// reuse the add history marshalling
	capnpAddHistory := hubapi.CapAddHistory_ServerToClient(ahCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capnpAddHistory)
	}
	return err
}

func (capsrv *HistoryServiceCapnpServer) CapReadHistory(
	ctx context.Context, call hubapi.CapHistoryService_capReadHistory) error {

	// create a client instance for reading the history
	args := call.Args()
	thingAddr, _ := args.ThingAddr()
	readSrv := &ReadHistoryCapnpServer{
		svc: capsrv.svc.CapReadHistory(ctx, thingAddr),
	}
	capnpReadHistory := hubapi.CapReadHistory_ServerToClient(readSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capnpReadHistory)
	}
	return err
}

//func (capsrv *HistoryServiceCapnpServer) Info(
//	ctx context.Context, call hubapi.CapHistoryService_info) (err error) {
//
//	inf, err := capsrv.svc.Info(ctx)
//	if err == nil {
//		res, err2 := call.AllocResults()
//		err = err2
//		storeInfo, _ := res.NewStatistics()
//		storeInfo.SetNrActions(int64(inf.NrActions))
//		storeInfo.SetNrEvents(int64(inf.NrEvents))
//		storeInfo.SetEngine(inf.Engine)
//		storeInfo.SetUptime(int64(inf.Uptime))
//	}
//
//	return err
//}

// StartHistoryServiceCapnpServer returns the capnp protocol server for the history store
func StartHistoryServiceCapnpServer(_ context.Context, listener net.Listener, svc history.IHistoryService) error {

	capsrv := &HistoryServiceCapnpServer{
		svc: svc,
	}
	capRegSrv := client.NewCapRegistrationServer(
		history.ServiceName, hubapi.CapHistoryService_Methods(nil, capsrv))
	capRegSrv.ExportCapability("capAddHistory", []string{hubapi.ClientTypeService})
	capRegSrv.ExportCapability("capAddAnyThing", []string{hubapi.ClientTypeService})
	capRegSrv.ExportCapability("capReadHistory",
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeUser})
	capsrv.capRegSrv = capRegSrv

	// Create the capnp handler to receive requests
	main := hubapi.CapHistoryService_ServerToClient(capsrv)
	err := caphelp.Serve(listener, capnp.Client(main), nil)
	return err
}
