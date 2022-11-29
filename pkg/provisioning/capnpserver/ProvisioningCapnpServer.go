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
	caphelp.HiveOTServiceCapnpServer
	// the plain-old-go-object provisioning server
	svc provisioning.IProvisioning
}

func (capsrv *ProvisioningCapnpServer) CapManageProvisioning(
	ctx context.Context, call hubapi.CapProvisioning_capManageProvisioning) error {

	// create the service instance for this request
	mngCapSrv := &ManageProvisioningCapnpServer{
		pogosrv: capsrv.svc.CapManageProvisioning(ctx),
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
		pogosrv: capsrv.svc.CapRefreshProvisioning(ctx),
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
		pogosrv: capsrv.svc.CapRequestProvisioning(ctx),
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
	ctx context.Context, lis net.Listener, svc provisioning.IProvisioning) error {

	logrus.Infof("Starting provisioning service capnp adapter on: %s", lis.Addr())

	srv := &ProvisioningCapnpServer{
		HiveOTServiceCapnpServer: caphelp.NewHiveOTServiceCapnpServer(provisioning.ServiceName),
		svc:                      svc,
	}
	srv.RegisterKnownMethods(hubapi.CapProvisioning_Methods(nil, srv))
	srv.ExportCapability("capManageProvisioning", []string{hubapi.ClientTypeService})
	srv.ExportCapability("capRequestProvisioning", []string{hubapi.ClientTypeService, hubapi.ClientTypeIotDevice})
	srv.ExportCapability("capRefreshProvisioning", []string{hubapi.ClientTypeService, hubapi.ClientTypeIotDevice})

	//
	main := hubapi.CapProvisioning_ServerToClient(srv)
	err := caphelp.CapServe(ctx, provisioning.ServiceName, lis, capnp.Client(main))
	return err
}
