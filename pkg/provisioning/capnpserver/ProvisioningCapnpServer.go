package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/provisioning"
)

// ProvisioningCapnpServer provides the capnproto RPC server for IOT device provisioning.
// This implements the capnproto generated interface Provisioning_Server
// See hub.capnp/go/hubapi/Provisioning.capnp.go for the interface.
type ProvisioningCapnpServer struct {
	// the plain-old-go-object provisioning server
	srv provisioning.IProvisioning
}

func (capsrv *ProvisioningCapnpServer) CapManageProvisioning(
	ctx context.Context, call hubapi.CapProvisioning_capManageProvisioning) error {

	// create the service instance for this request
	mngCapSrv := &ManageProvisioningCapnpServer{
		pogosrv: capsrv.srv.CapManageProvisioning(ctx),
	}

	// wrap it with a capnp proxy
	capability := hubapi.CapManageProvisioning_ServerToClient(mngCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		// return the proxy
		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *ProvisioningCapnpServer) CapRefreshProvisioning(
	ctx context.Context, call hubapi.CapProvisioning_capRefreshProvisioning) error {

	// create the service instance for this request
	// TODO: restrict it to the deviceID of the caller
	refreshCapSrv := &RefreshProvisioningCapnpServer{
		pogosrv: capsrv.srv.CapRefreshProvisioning(ctx),
	}

	// wrap it with a capnp proxy
	capability := hubapi.CapRefreshProvisioning_ServerToClient(refreshCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		// return the proxy
		err = res.SetCap(capability)
	}
	return err
}
func (capsrv *ProvisioningCapnpServer) CapRequestProvisioning(
	ctx context.Context, call hubapi.CapProvisioning_capRequestProvisioning) error {
	// create the service instance for this request
	reqCapSrv := &RequestProvisioningCapnpServer{
		pogosrv: capsrv.srv.CapRequestProvisioning(ctx),
	}

	// wrap it with a capnp proxy
	capability := hubapi.CapRequestProvisioning_ServerToClient(reqCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartProvisioningCapnpServer starts the capnp server for the provisioning service
func StartProvisioningCapnpServer(
	ctx context.Context, lis net.Listener, srv provisioning.IProvisioning) error {

	logrus.Infof("Starting provisioning service capnp adapter on: %s", lis.Addr())

	main := hubapi.CapProvisioning_ServerToClient(&ProvisioningCapnpServer{
		srv: srv,
	})

	return caphelp.CapServe(ctx, provisioning.ServiceName, lis, capnp.Client(main))
}
