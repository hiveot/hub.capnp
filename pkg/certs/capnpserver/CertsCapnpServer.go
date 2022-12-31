// Package capnpserver with the capnproto server for the CapCerts API
package capnpserver

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/resolver/capprovider"
)

// CertsCapnpServer provides the capnpr RPC server for interface hubapi.CapCerts_Server
// See hub.capnp/go/hubapi/Certs.capnp.go for the interface
type CertsCapnpServer struct {
	svc certs.ICerts
}

// CapDeviceCerts provides the device certificate capability
// TBD, auth for handing out this capability
// TODO: option to restrict capability to a single device
func (capsrv *CertsCapnpServer) CapDeviceCerts(
	ctx context.Context, call hubapi.CapCerts_capDeviceCerts) error {

	// Create the capnp proxy that provides the capability to create device certificates
	clientID, _ := call.Args().ClientID()
	capDeviceCerts, _ := capsrv.svc.CapDeviceCerts(ctx, clientID)
	deviceCertsSrv := &DeviceCertsCapnpServer{
		srv: capDeviceCerts,
	}

	capability := hubapi.CapDeviceCerts_ServerToClient(deviceCertsSrv)
	//capability := hubapi.CapDeviceCerts_ServerToClient(capsrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// CapServiceCerts provides the service certificate capability
// TBD, auth for handing out this capability
func (capsrv *CertsCapnpServer) CapServiceCerts(
	ctx context.Context, call hubapi.CapCerts_capServiceCerts) error {

	// Create the capnp proxy that provides the capability to create certificates
	clientID, _ := call.Args().ClientID()
	capServiceCerts, _ := capsrv.svc.CapServiceCerts(ctx, clientID)
	serviceCertsSrv := &ServiceCertsCapnpServer{
		srv: capServiceCerts,
	}

	capability := hubapi.CapServiceCerts_ServerToClient(serviceCertsSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// CapUserCerts provides the service certificate capability
// TBD, auth for handing out this capability
func (capsrv *CertsCapnpServer) CapUserCerts(
	ctx context.Context, call hubapi.CapCerts_capUserCerts) error {

	// Create the capnp proxy that provides the capability to create certificates
	clientID, _ := call.Args().ClientID()
	capUserCerts, _ := capsrv.svc.CapUserCerts(ctx, clientID)
	userCertsSrv := &UserCertsCapnpServer{
		srv: capUserCerts,
	}

	capability := hubapi.CapUserCerts_ServerToClient(userCertsSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// CapVerifyCerts provides the certificate verification capability
// TBD, auth for handing out this capability
func (capsrv *CertsCapnpServer) CapVerifyCerts(
	ctx context.Context, call hubapi.CapCerts_capVerifyCerts) error {

	// Create the capnp proxy that provides the capability to create certificates
	clientID, _ := call.Args().ClientID()
	capVerifyCerts, _ := capsrv.svc.CapVerifyCerts(ctx, clientID)
	verifyCertsSrv := &VerifyCertsCapnpServer{
		srv: capVerifyCerts,
	}

	capability := hubapi.CapVerifyCerts_ServerToClient(verifyCertsSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartCertsCapnpServer starts the capnp protocol server for the certificates service
//
//	svc is the service implementation
//	lis is the service listning endpoint for direct connections
//	resolverSocket is the UDS socket of the resolver used to register the service capabilities. "" to not register.
func StartCertsCapnpServer(svc certs.ICerts, lis net.Listener) (err error) {

	serviceName := certs.ServiceName
	srv := &CertsCapnpServer{
		svc: svc,
	}

	// the provider serves the exported capabilities
	capProv := capprovider.NewCapServer(
		serviceName,
		hubapi.CapCerts_Methods(nil, srv))

	// register the methods available through getCapability
	capProv.ExportCapability("capDeviceCerts", []string{hubapi.ClientTypeService})
	capProv.ExportCapability("capServiceCerts", []string{hubapi.ClientTypeService})
	capProv.ExportCapability("capUserCerts", []string{hubapi.ClientTypeService})
	capProv.ExportCapability("capVerifyCerts", []string{
		hubapi.ClientTypeService,
		hubapi.ClientTypeIotDevice,
		hubapi.ClientTypeUser,
		hubapi.ClientTypeUnauthenticated,
	})

	logrus.Infof("Starting '%s' service capnp adapter listening on: %s", serviceName, lis.Addr())
	err = capProv.Start(lis)
	return err
}
