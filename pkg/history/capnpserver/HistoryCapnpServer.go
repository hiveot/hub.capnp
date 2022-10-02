package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/history"
)

// HistoryCapnpServer is a capnproto adapter for the history store
// This implements the capnproto generated interface History_Server
// See hub.capnp/go/hubapi/HistoryStore.capnp.go for the interface.
type HistoryCapnpServer struct {
	srv history.IHistory
}

func (capsrv *HistoryCapnpServer) CapReadHistory(
	ctx context.Context, call hubapi.CapHistory_capReadHistory) error {
	// create a client instance for reading the history
	readHistoryCapSrv := NewReadHistoryCapnpServer(capsrv.srv.CapReadHistory())
	cap := hubapi.CapReadHistory_ServerToClient(readHistoryCapSrv)
	res, err := call.AllocResults()
	res.SetCap(cap)
	return err
}
func (capsrv *HistoryCapnpServer) CapUpdateHistory(
	ctx context.Context, call hubapi.CapHistory_capUpdateHistory) error {
	// create a client instance for updating the history
	updateHistoryCapSrv := NewUpdateHistoryCapnpServer(capsrv.srv.CapUpdateHistory())
	cap := hubapi.CapUpdateHistory_ServerToClient(updateHistoryCapSrv)
	res, err := call.AllocResults()
	res.SetCap(cap)
	return err
}

// StartHistoryCapnpServer starts the capnp protocol server for the history store
func StartHistoryCapnpServer(ctx context.Context, listener net.Listener, srv history.IHistory) error {

	adpt := &HistoryCapnpServer{
		srv: srv,
	}
	// Create the capnp client to receive requests
	main := hubapi.CapHistory_ServerToClient(adpt)

	return caphelp.CapServe(ctx, listener, capnp.Client(main))
}
