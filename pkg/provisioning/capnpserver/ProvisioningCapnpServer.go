package capnpserver

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/resolver/capprovider"
)

// ProvisioningCapnpServer provides the capnproto RPC server for IOT device provisioning.
// This implements the capnproto generated interface Provisioning_Server
// See hub.capnp/go/hubapi/Provisioning.capnp.go for the interface.
type ProvisioningCapnpServer struct {
	// the plain-old-go-object provisioning server
	svc provisioning.IProvisioning
}

func (capsrv *ProvisioningCapnpServer) CapManageProvisioning(
	ctx context.Context, call hubapi.CapProvisioning_capManageProvisioning) error {

	clientID, _ := call.Args().ClientID()
	// create the service instance for this request
	mngCapSrv := &ManageProvisioningCapnpServer{
		pogosrv: capsrv.svc.CapManageProvisioning(ctx, clientID),
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

	clientID, _ := call.Args().ClientID()
	// create the service instance for this request
	// TODO: restrict it to the deviceID of the caller
	refreshCapSrv := &RefreshProvisioningCapnpServer{
		pogosrv: capsrv.svc.CapRefreshProvisioning(ctx, clientID),
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

	clientID, _ := call.Args().ClientID()
	// create the service instance for this request
	reqCapSrv := &RequestProvisioningCapnpServer{
		pogosrv: capsrv.svc.CapRequestProvisioning(ctx, clientID),
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
func StartProvisioningCapnpServer(svc provisioning.IProvisioning, lis net.Listener) error {
	serviceName := provisioning.ServiceName

	srv := &ProvisioningCapnpServer{
		svc: svc,
	}
	capProv := capprovider.NewCapServer(
		serviceName, hubapi.CapProvisioning_Methods(nil, srv))

	capProv.ExportCapability("capManageProvisioning",
		[]string{hubapi.ClientTypeService})

	capProv.ExportCapability("capRequestProvisioning",
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeIotDevice})

	capProv.ExportCapability("capRefreshProvisioning",
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeIotDevice})

	logrus.Infof("Starting provisioning service capnp adapter on: %s", lis.Addr())
	err := capProv.Start(lis)
	return err
}
