package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/authz"
)

// AuthzCapnpServer provides the capnp RPC server for authorization services
// This implements the capnproto generated interface Authz_Server
// See hub.capnp/go/hubapi/Authz.capnp.go for the interface.
type AuthzCapnpServer struct {
	srv authz.IAuthz
}

func (capsrv *AuthzCapnpServer) CapClientAuthz(
	ctx context.Context, call hubapi.CapAuthz_capClientAuthz) error {

	clientID, _ := call.Args().ClientID()
	clientAuthzCapSrv := &ClientAuthzCapnpServer{
		srv: capsrv.srv.CapClientAuthz(ctx, clientID),
	}
	capability := hubapi.CapClientAuthz_ServerToClient(clientAuthzCapSrv)

	res, err := call.AllocResults()
	if err == nil {

		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *AuthzCapnpServer) CapManageAuthz(ctx context.Context, call hubapi.CapAuthz_capManageAuthz) error {
	manageAuthzCapSrv := &ManageAuthzCapnpServer{
		srv: capsrv.srv.CapManageAuthz(ctx),
	}
	capability := hubapi.CapManageAuthz_ServerToClient(manageAuthzCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *AuthzCapnpServer) CapVerifyAuthz(ctx context.Context, call hubapi.CapAuthz_capVerifyAuthz) error {
	verifyAuthzSrv := &VerifyAuthzCapnpServer{
		srv: capsrv.srv.CapVerifyAuthz(ctx),
	}
	capability := hubapi.CapVerifyAuthz_ServerToClient(verifyAuthzSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartAuthzCapnpServer starts the capnp protocol server for the authentication service
func StartAuthzCapnpServer(ctx context.Context, lis net.Listener, srv authz.IAuthz) error {

	logrus.Infof("Starting authz service capnp adapter on: %s", lis.Addr())

	main := hubapi.CapAuthz_ServerToClient(&AuthzCapnpServer{
		srv: srv,
	})

	err := caphelp.CapServe(ctx, authz.ServiceName, lis, capnp.Client(main))
	return err
}
