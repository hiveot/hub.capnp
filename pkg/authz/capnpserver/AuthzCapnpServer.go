package capnpserver

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/resolver/capprovider"
)

// AuthzCapnpServer provides the capnp RPC server for authorization services
// This implements the capnproto generated interface Authz_Server
// See hub.capnp/go/hubapi/Authz.capnp.go for the interface.
type AuthzCapnpServer struct {
	svc authz.IAuthz
}

func (capsrv *AuthzCapnpServer) CapClientAuthz(
	ctx context.Context, call hubapi.CapAuthz_capClientAuthz) error {

	clientID, _ := call.Args().ClientID()
	capClientAuthz, _ := capsrv.svc.CapClientAuthz(ctx, clientID)
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

	clientID, _ := call.Args().ClientID()
	capManageAuthz, _ := capsrv.svc.CapManageAuthz(ctx, clientID)
	manageAuthzCapSrv := &ManageAuthzCapnpServer{
		srv: capManageAuthz,
	}
	capability := hubapi.CapManageAuthz_ServerToClient(manageAuthzCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *AuthzCapnpServer) CapVerifyAuthz(ctx context.Context, call hubapi.CapAuthz_capVerifyAuthz) error {

	clientID, _ := call.Args().ClientID()
	capVerifyAuthz, _ := capsrv.svc.CapVerifyAuthz(ctx, clientID)
	verifyAuthzSrv := &VerifyAuthzCapnpServer{
		srv: capVerifyAuthz,
	}
	capability := hubapi.CapVerifyAuthz_ServerToClient(verifyAuthzSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartAuthzCapnpServer starts the capnp protocol server for the authentication service
//
//	svc is the service implementation
//	lis is the cap provider listening endpoint
func StartAuthzCapnpServer(svc authz.IAuthz, lis net.Listener) (err error) {

	srv := &AuthzCapnpServer{
		svc: svc,
	}

	// the provider serves the exported capabilities
	capProv := capprovider.NewCapServer(
		authz.ServiceName,
		hubapi.CapAuthz_Methods(nil, srv))

	// register the methods available through getCapability
	capProv.ExportCapability(hubapi.CapNameClientAuthz,
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeUser})

	capProv.ExportCapability(hubapi.CapNameManageAuthz,
		[]string{hubapi.ClientTypeService})

	capProv.ExportCapability(hubapi.CapNameVerifyAuthz,
		[]string{hubapi.ClientTypeService})

	// listen for direct connections
	logrus.Infof("Starting 'authz' service capnp adapter listening on: %s", lis.Addr())
	err = capProv.Start(lis)

	return err
}
