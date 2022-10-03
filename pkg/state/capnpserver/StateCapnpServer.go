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
	srv state.IState
}

func (capsrv *StateCapnpServer) Get(
	ctx context.Context, call hubapi.CapState_get) error {
	args := call.Args()
	key, _ := args.Key()
	value, err := capsrv.srv.Get(ctx, key)
	if err == nil {
		res, err := call.AllocResults()
		if err == nil {
			err = res.SetValue(value)
		}
	}
	return err
}

func (capsrv *StateCapnpServer) Set(
	ctx context.Context, call hubapi.CapState_set) error {
	args := call.Args()
	key, _ := args.Key()
	value, _ := args.Value()
	err := capsrv.srv.Set(ctx, key, value)
	return err
}

// StartStateCapnpServer starts the capnp protocol server for the state store
func StartStateCapnpServer(ctx context.Context, lis net.Listener, srv state.IState) error {

	logrus.Infof("Starting state store capnp adapter on: %s", lis.Addr())

	main := hubapi.CapState_ServerToClient(&StateCapnpServer{
		srv: srv,
	})

	err := caphelp.CapServe(ctx, lis, capnp.Client(main))
	return err
}
