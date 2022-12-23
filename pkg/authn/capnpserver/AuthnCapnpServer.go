package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/resolver/client"
)

// AuthnCapnpServer provides the capnp RPC server for authentication services
// This implements the capnproto generated interface Authn_Server
// See hub.capnp/go/hubapi/Authn.capnp.go for the interface.
type AuthnCapnpServer struct {
	capRegSrv *client.CapRegistrationServer
	svc       authn.IAuthnService
}

func (capsrv *AuthnCapnpServer) CapUserAuthn(
	ctx context.Context, call hubapi.CapAuthn_capUserAuthn) error {

	clientID, _ := call.Args().ClientID()
	userAuthnCapSrv := &UserAuthnCapnpServer{
		svc: capsrv.svc.CapUserAuthn(ctx, clientID),
	}
	capability := hubapi.CapUserAuthn_ServerToClient(userAuthnCapSrv)

	res, err := call.AllocResults()
	if err == nil {

		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *AuthnCapnpServer) CapManageAuthn(ctx context.Context, call hubapi.CapAuthn_capManageAuthn) error {
	manageAuthnCapSrv := &ManageAuthnCapnpServer{
		svc: capsrv.svc.CapManageAuthn(ctx),
	}
	capability := hubapi.CapManageAuthn_ServerToClient(manageAuthnCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartAuthnCapnpServer starts the capnp protocol server for the authentication service
// lis is optional if the service needs to be listening on its own endpoint instead of using the resolver.
func StartAuthnCapnpServer(lis net.Listener, svc authn.IAuthnService) error {

	srv := &AuthnCapnpServer{
		svc: svc,
	}
	// this server will handle capability registration for us.
	capRegSrv := client.NewCapRegistrationServer(
		authn.ServiceName,
		hubapi.CapAuthn_Methods(nil, srv))

	// register the methods available through getCapability
	capRegSrv.ExportCapability("capUserAuthn",
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeUser, hubapi.ClientTypeUnauthenticated})
	capRegSrv.ExportCapability("capManageAuthn",
		[]string{hubapi.ClientTypeService})

	err := capRegSrv.Start("")
	if err != nil {
		logrus.Warningf("unable to connect to the resolver service: %s", err)
	}

	// also listen, although that isn't needed if the resolver works.
	if lis != nil {
		logrus.Infof("Starting Authn service capnp adapter listening on: %s", lis.Addr())
		main := hubapi.CapAuthn_ServerToClient(srv)
		err = caphelp.Serve(lis, capnp.Client(main), nil)
	}
	return err
}
