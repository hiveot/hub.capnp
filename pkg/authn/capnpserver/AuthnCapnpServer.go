package capnpserver

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/resolver/capprovider"
)

// AuthnCapnpServer provides the capnp RPC server for authentication services
// This implements the capnproto generated interface Authn_Server
// See hub.capnp/go/hubapi/Authn.capnp.go for the interface.
type AuthnCapnpServer struct {
	svc authn.IAuthnService
}

func (capsrv *AuthnCapnpServer) CapUserAuthn(
	ctx context.Context, call hubapi.CapAuthn_capUserAuthn) error {

	clientID, _ := call.Args().ClientID()
	userAuthInstance, err := capsrv.svc.CapUserAuthn(ctx, clientID)
	if err != nil {
		return err
	}
	userAuthnCapSrv := &UserAuthnCapnpServer{
		svc: userAuthInstance,
	}
	capability := hubapi.CapUserAuthn_ServerToClient(userAuthnCapSrv)

	res, err := call.AllocResults()
	if err == nil {

		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *AuthnCapnpServer) CapManageAuthn(ctx context.Context, call hubapi.CapAuthn_capManageAuthn) error {
	clientID, _ := call.Args().ClientID()
	manageAuthInstance, err := capsrv.svc.CapManageAuthn(ctx, clientID)
	if err != nil {
		return err
	}
	manageAuthnCapSrv := &ManageAuthnCapnpServer{
		svc: manageAuthInstance,
	}
	capability := hubapi.CapManageAuthn_ServerToClient(manageAuthnCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartAuthnCapnpServer starts the capnp protocol server for the authentication service
// The starts the cap-provider server on the listener
//
//	svc is the service implementation
//	lis is the cap provider listening endpoint
func StartAuthnCapnpServer(svc authn.IAuthnService, lis net.Listener) (err error) {
	serviceName := authn.ServiceName

	srv := &AuthnCapnpServer{
		svc: svc,
	}
	// the provider serves the exported capabilities
	capProv := capprovider.NewCapServer(
		serviceName, hubapi.CapAuthn_Methods(nil, srv))

	capProv.ExportCapability(hubapi.CapNameUserAuthn,
		[]string{hubapi.AuthTypeService, hubapi.AuthTypeUser, hubapi.AuthTypeUnauthenticated})

	capProv.ExportCapability(hubapi.CapNameManageAuthn,
		[]string{hubapi.AuthTypeService})

	logrus.Infof("Starting '%s' service capnp adapter listening on: %s", serviceName, lis.Addr())
	err = capProv.Start(lis)

	return err
}
