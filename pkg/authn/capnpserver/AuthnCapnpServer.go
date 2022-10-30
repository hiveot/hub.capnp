package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/authn"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

// AuthnCapnpServer provides the capnp RPC server for authentication services
// This implements the capnproto generated interface Authn_Server
// See hub.capnp/go/hubapi/Authn.capnp.go for the interface.
type AuthnCapnpServer struct {
	srv authn.IAuthn
}

func (capsrv *AuthnCapnpServer) CapUserAuthn(
	ctx context.Context, call hubapi.CapAuthn_capUserAuthn) error {

	clientID, _ := call.Args().ClientID()
	userAuthnCapSrv := &UserAuthnCapnpServer{
		svc: capsrv.srv.CapUserAuthn(ctx, clientID),
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
		svc: capsrv.srv.CapManageAuthn(ctx),
	}
	capability := hubapi.CapManageAuthn_ServerToClient(manageAuthnCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartAuthnCapnpServer starts the capnp protocol server for the authentication service
func StartAuthnCapnpServer(ctx context.Context, lis net.Listener, srv authn.IAuthn) error {

	logrus.Infof("Starting Authn service capnp adapter on: %s", lis.Addr())

	main := hubapi.CapAuthn_ServerToClient(&AuthnCapnpServer{
		srv: srv,
	})

	err := caphelp.CapServe(ctx, authn.ServiceName, lis, capnp.Client(main))
	return err
}
