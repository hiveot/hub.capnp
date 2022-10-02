// Package capnpserver with the capnproto server for the CapCerts API
package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/certs"
)

// CertsCapnpServer provides the capnpr RPC server for interface hubapi.CapCerts_Server
// See hub.capnp/go/hubapi/Certs.capnp.go for the interface
type CertsCapnpServer struct {
	srv certs.ICerts
}

// CapDeviceCerts provides the device certificate capability
// TBD, auth for handing out this capability
// TODO: option to restrict capability to a single device
func (capsrv *CertsCapnpServer) CapDeviceCerts(
	ctx context.Context, call hubapi.CapCerts_capDeviceCerts) error {

	// Create the capnp proxy that provides the capability to create device certificates
	// TODO: use context to identify caller and include as part of restrictions?
	deviceCertsSrv := NewDeviceCertsCapnpServer(capsrv.srv.CapDeviceCerts())
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
	_ context.Context, call hubapi.CapCerts_capServiceCerts) error {

	// Create the capnp proxy that provides the capability to create certificates
	serviceCertsSrv := NewServiceCertsCapnpServer(capsrv.srv.CapServiceCerts())
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
	_ context.Context, call hubapi.CapCerts_capUserCerts) error {

	// Create the capnp proxy that provides the capability to create certificates
	userCertsSrv := NewUserCertsCapnpServer(capsrv.srv.CapUserCerts())
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
	verifyCertsSrv := NewVerifyCertsCapnpServer(capsrv.srv.CapVerifyCerts())
	capability := hubapi.CapVerifyCerts_ServerToClient(verifyCertsSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartCertsCapnpServer starts the capnp protocol server for the certificates service
func StartCertsCapnpServer(ctx context.Context, lis net.Listener, srv certs.ICerts) error {

	logrus.Infof("Starting certs service capnp server on: %s", lis.Addr())

	// Create the capnp proxy that provides the certificate capability
	main := hubapi.CapCerts_ServerToClient(&CertsCapnpServer{
		srv: srv,
	})

	// serve the requests by creating client instances of main (I think)
	err := caphelp.CapServe(ctx, lis, capnp.Client(main))
	return err
}
