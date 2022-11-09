package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/state"
)

// StateStoreCapnpServer provides the capnp RPC server for state store
// This implements the capnproto generated interface State_Server
// See hub.capnp/go/hubapi/State.capnp.go for the interface.
type StateStoreCapnpServer struct {
	svc state.IStateService
}

// CapClientState returns a capnp server instance for accessing client state
// this wraps the POGS server with a capnp binding for client application state access
func (capsrv *StateStoreCapnpServer) CapClientState(
	ctx context.Context, call hubapi.CapState_capClientState) error {

	// first create the instance of the POGS server for this client application
	args := call.Args()
	clientID, _ := args.ClientID()
	appID, _ := args.AppID()
	pogoClientStateServer, err := capsrv.svc.CapClientBucket(ctx, clientID, appID)
	if err == nil {
		// second, wrap it in a capnp binding which implements the capnp generated API
		capnpClientStateServer := &ClientStateCapnpServer{srv: pogoClientStateServer}

		// last, create the capnp RPC server for this capability
		capability := hubapi.CapClientState_ServerToClient(capnpClientStateServer)

		res, err := call.AllocResults()
		if err == nil {
			err = res.SetCap(capability)
		}
	}
	return err
}

func (capsrv *StateStoreCapnpServer) Shutdown() {
	// Release on the client calls capnp Shutdown ... or does it?
	logrus.Infof("shutting down service")
	capsrv.svc.Stop()
}

// StartStateCapnpServer starts the capnp protocol server for the state store
// The capnp server will release the service on shutdown.
func StartStateCapnpServer(ctx context.Context, lis net.Listener, svc state.IStateService) error {

	logrus.Infof("Starting state store capnp adapter on: %s", lis.Addr())

	main := hubapi.CapState_ServerToClient(&StateStoreCapnpServer{
		svc: svc,
	})

	err := caphelp.CapServe(ctx, state.ServiceName, lis, capnp.Client(main))
	return err
}
