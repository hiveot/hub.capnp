package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/resolver/client"
)

// AuthzCapnpServer provides the capnp RPC server for authorization services
// This implements the capnproto generated interface Authz_Server
// See hub.capnp/go/hubapi/Authz.capnp.go for the interface.
type AuthzCapnpServer struct {
	// the capability provider for use by the resolver
	capProvider *client.CapRegistrationServer
	svc         authz.IAuthz
}

func (capsrv *AuthzCapnpServer) CapClientAuthz(
	ctx context.Context, call hubapi.CapAuthz_capClientAuthz) error {

	clientID, _ := call.Args().ClientID()
	capClientAuthz := capsrv.svc.CapClientAuthz(ctx, clientID)
	capClientAuthzCapnp := &ClientAuthzCapnpServer{
		srv: capClientAuthz,
	}
	capability := hubapi.CapClientAuthz_ServerToClient(capClientAuthzCapnp)

	res, err := call.AllocResults()
	if err == nil {

		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *AuthzCapnpServer) CapManageAuthz(ctx context.Context, call hubapi.CapAuthz_capManageAuthz) error {
	manageAuthzCapSrv := &ManageAuthzCapnpServer{
		srv: capsrv.svc.CapManageAuthz(ctx),
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
		srv: capsrv.svc.CapVerifyAuthz(ctx),
	}
	capability := hubapi.CapVerifyAuthz_ServerToClient(verifyAuthzSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartAuthzCapnpServer starts the capnp protocol server for the authentication service
func StartAuthzCapnpServer(lis net.Listener, svc authz.IAuthz) error {

	logrus.Infof("Starting authz service capnp adapter on: %s", lis.Addr())
	srv := &AuthzCapnpServer{svc: svc}
	capProvider := client.NewCapRegistrationServer(authz.ServiceName, hubapi.CapAuthz_Methods(nil, srv))
	srv.capProvider = capProvider
	// register the methods available through getCapability
	capProvider.ExportCapability("capClientAuthz",
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeUser})
	capProvider.ExportCapability("capManageAuthz",
		[]string{hubapi.ClientTypeService})
	capProvider.ExportCapability("capVerifyAuthz",
		[]string{hubapi.ClientTypeService})
	_ = capProvider.Start("")

	main := hubapi.CapAuthz_ServerToClient(srv)
	err := caphelp.Serve(lis, capnp.Client(main), nil)
	return err
}
