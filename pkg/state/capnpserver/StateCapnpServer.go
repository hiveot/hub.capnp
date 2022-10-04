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

// StateCapnpServer provides the capnp RPC server for state store
// This implements the capnproto generated interface State_Server
// See hub.capnp/go/hubapi/State.capnp.go for the interface.
type StateCapnpServer struct {
	pogo state.IState
}

// CapClientState returns a capnp server instance for accessing client state
// this wraps the POGS server with a capnp binding for client application state access
func (capsrv *StateCapnpServer) CapClientState(
	ctx context.Context, call hubapi.CapState_capClientState) error {

	// first create the instance of the POGS server for this client application
	args := call.Args()
	clientID, _ := args.ClientID()
	appID, _ := args.AppID()
	pogoClientStateServer := capsrv.pogo.CapClientState(ctx, clientID, appID)
	// second, wrap it in a capnp binding which implements the capnp generated API
	capnpClientStateServer := &ClientStateCapnpServer{srv: pogoClientStateServer}

	// last, create the capnp RPC server for this capability
	capability := hubapi.CapClientState_ServerToClient(capnpClientStateServer)

	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartStateCapnpServer starts the capnp protocol server for the state store
func StartStateCapnpServer(ctx context.Context, lis net.Listener, srv state.IState) error {

	logrus.Infof("Starting state store capnp adapter on: %s", lis.Addr())

	main := hubapi.CapState_ServerToClient(&StateCapnpServer{
		pogo: srv,
	})

	err := caphelp.CapServe(ctx, lis, capnp.Client(main))
	return err
}
