// Package capnpserver with the capnproto server for the CapCerts API
package capnpserver

import (
	"context"
	"log"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/certs"
)

// CertsCapnpServer provides the capnpr RPC server for interface hubapi.CapCerts_Server
// See hub.capnp/go/hubapi/Certs.capnp.go for the interface
type CertsCapnpServer struct {
	caphelp.HiveOTServiceCapnpServer
	svc certs.ICerts
}

// CapDeviceCerts provides the device certificate capability
// TBD, auth for handing out this capability
// TODO: option to restrict capability to a single device
func (capsrv *CertsCapnpServer) CapDeviceCerts(
	ctx context.Context, call hubapi.CapCerts_capDeviceCerts) error {

	// Create the capnp proxy that provides the capability to create device certificates
	// TODO: use context to identify caller and include as part of restrictions?
	deviceCertsSrv := &DeviceCertsCapnpServer{
		srv: capsrv.svc.CapDeviceCerts(ctx),
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
	serviceCertsSrv := &ServiceCertsCapnpServer{
		srv: capsrv.svc.CapServiceCerts(ctx),
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
	userCertsSrv := &UserCertsCapnpServer{
		srv: capsrv.svc.CapUserCerts(ctx),
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
	verifyCertsSrv := &VerifyCertsCapnpServer{
		srv: capsrv.svc.CapVerifyCerts(ctx),
	}

	capability := hubapi.CapVerifyCerts_ServerToClient(verifyCertsSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartCertsCapnpServer starts the capnp protocol server for the certificates service
func StartCertsCapnpServer(_ context.Context, lis net.Listener, svc certs.ICerts) error {

	log.Printf("Starting certs service capnp server on: %s", lis.Addr())

	srv := &CertsCapnpServer{
		HiveOTServiceCapnpServer: caphelp.NewHiveOTServiceCapnpServer(certs.ServiceName),
		svc:                      svc,
	}
	// register the methods available through getCapability
	srv.RegisterKnownMethods(hubapi.CapCerts_Methods(nil, srv))
	srv.ExportCapability("capDeviceCerts", []string{hubapi.ClientTypeService})
	srv.ExportCapability("capServiceCerts", []string{hubapi.ClientTypeService})
	srv.ExportCapability("capUserCerts", []string{hubapi.ClientTypeService})
	srv.ExportCapability("capVerifyCerts", []string{
		hubapi.ClientTypeService,
		hubapi.ClientTypeIotDevice,
		hubapi.ClientTypeUser,
		hubapi.ClientTypeUnauthenticated,
	})

	// Create the capnp server proxy that provides the certificate capability
	main := hubapi.CapCerts_ServerToClient(srv)
	//err := caphelp.ServeCapnp(lis, capnp.Client(main))
	err := rpc.Serve(lis, capnp.Client(main))

	log.Printf("Certs service capnp server stopped")
	return err
}
