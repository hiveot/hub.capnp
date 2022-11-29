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
	caphelp.HiveOTServiceCapnpServer
	svc authz.IAuthz
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
func StartAuthzCapnpServer(ctx context.Context, lis net.Listener, svc authz.IAuthz) error {

	logrus.Infof("Starting authz service capnp adapter on: %s", lis.Addr())
	srv := &AuthzCapnpServer{
		HiveOTServiceCapnpServer: caphelp.NewHiveOTServiceCapnpServer(authz.ServiceName),
		svc:                      svc,
	}
	// register the methods available through getCapability
	srv.RegisterKnownMethods(hubapi.CapAuthz_Methods(nil, srv))
	srv.ExportCapability("capClientAuthz",
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeUser})
	srv.ExportCapability("capManageAuthz",
		[]string{hubapi.ClientTypeService})
	srv.ExportCapability("capVerifyAuthz",
		[]string{hubapi.ClientTypeService})

	main := hubapi.CapAuthz_ServerToClient(srv)
	err := caphelp.CapServe(ctx, authz.ServiceName, lis, capnp.Client(main))
	return err
}
