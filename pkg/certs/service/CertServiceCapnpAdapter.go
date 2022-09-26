// Package service with the capnproto adapter
package service

import (
	"context"
	"net"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/certs"
)

// CertServiceCapnpAdapter implements the capnproto generated interface ICertsService_Server
// See hub.capnp/go/hubapi/Cert.capnp.go for the interface
type CertServiceCapnpAdapter struct {
	srv certs.ICerts
}

func (adr *CertServiceCapnpAdapter) CreateDeviceCert(
	_ context.Context, call hubapi.CapDeviceCert_createDeviceCert) error {
	deviceID, _ := call.Args().DeviceID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	validityDays := call.Args().ValidityDays()
	if validityDays == 0 {
		validityDays = hubapi.DefaultClientCertValidityDays
	}
	certPEM, caCertPEM, err := adr.srv.CreateDeviceCert(deviceID, pubKeyPEM, int(validityDays))
	if err == nil {
		//logrus.Infof("Created device cert for %s", deviceID)
		res, err2 := call.AllocResults()
		res.SetCertPEM(certPEM)
		res.SetCaCertPEM(caCertPEM)
		err = err2
	}
	return err
}

func (adaptr *CertServiceCapnpAdapter) CreateServiceCert(
	_ context.Context, call hubapi.CapServiceCert_createServiceCert) error {
	clientID, _ := call.Args().ServiceID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	namesList, _ := call.Args().Names()
	validityDays := call.Args().ValidityDays()
	if validityDays == 0 {
		validityDays = hubapi.DefaultServiceCertValidityDays
	}
	names := []string{}
	for i := 0; i < namesList.Len(); i++ {
		name, _ := namesList.At(i)
		names = append(names, name)
	}
	certPEM, caCertPEM, err := adaptr.srv.CreateServiceCert(clientID, pubKeyPEM, names, int(validityDays))
	if err == nil {
		//logrus.Infof("Created device cert for %s", clientID)
		res, err2 := call.AllocResults()
		res.SetCertPEM(certPEM)
		res.SetCaCertPEM(caCertPEM)
		err = err2
	}
	return err
}

func (adaptr *CertServiceCapnpAdapter) CreateUserCert(
	ctx context.Context, call hubapi.CapUserCert_createUserCert) error {

	clientID, _ := call.Args().ClientID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	validityDays := call.Args().ValidityDays()
	if validityDays == 0 {
		validityDays = hubapi.DefaultClientCertValidityDays
	}
	certPEM, caCertPEM, err := adaptr.srv.CreateUserCert(clientID, pubKeyPEM, int(validityDays))
	if err == nil {
		//logrus.Infof("Created client cert for %s", clientID)
		res, err2 := call.AllocResults()
		res.SetCertPEM(certPEM)
		res.SetCaCertPEM(caCertPEM)
		err = err2
	}
	return err
}

func (adaptr *CertServiceCapnpAdapter) VerifyCert(
	ctx context.Context, call hubapi.CapVerifyCert_verifyCert) error {

	clientID, _ := call.Args().ClientID()
	certPEM, _ := call.Args().CertPEM()
	err := adaptr.srv.VerifyCert(clientID, certPEM)
	return err
}

// GetDeviceCertCapability returns the device certificate capability
func (adaptr *CertServiceCapnpAdapter) GetDeviceCertCapability(
	_ context.Context, call hubapi.CapCertService_getDeviceCertCapability) error {

	// Create the capnp proxy that provides the capability to create device certificates
	capability := hubapi.CapDeviceCert_ServerToClient(&CertServiceCapnpAdapter{
		srv: adaptr.srv,
	})
	res, err := call.AllocResults()
	res.SetCap(capability)

	return err
}

// GetServiceCertCapability returns the service certificate capability
func (adaptr *CertServiceCapnpAdapter) GetServiceCertCapability(
	_ context.Context, call hubapi.CapCertService_getServiceCertCapability) error {

	// Create the capnp proxy that provides the capability to create certificates
	capability := hubapi.CapServiceCert_ServerToClient(&CertServiceCapnpAdapter{
		srv: adaptr.srv,
	})
	res, err := call.AllocResults()
	res.SetCap(capability)

	return err
}

// GetUserCertCapability returns the service certificate capability
func (adaptr *CertServiceCapnpAdapter) GetUserCertCapability(
	_ context.Context, call hubapi.CapCertService_getUserCertCapability) error {

	// Create the capnp proxy that provides the capability to create certificates
	capability := hubapi.CapUserCert_ServerToClient(&CertServiceCapnpAdapter{
		srv: adaptr.srv,
	})
	res, err := call.AllocResults()
	res.SetCap(capability)

	return err
}

// GetVerifyCertCapability returns the certificate verification capability
func (adaptr *CertServiceCapnpAdapter) GetVerifyCertCapability(
	_ context.Context, call hubapi.CapCertService_getVerifyCertCapability) error {

	// Create the capnp proxy that provides the capability to create certificates
	capability := hubapi.CapVerifyCert_ServerToClient(&CertServiceCapnpAdapter{
		srv: adaptr.srv,
	})
	res, err := call.AllocResults()
	res.SetCap(capability)

	return err
}

// StartCertServiceCapnpAdapter starts the certificate service capnp protocol server
func StartCertServiceCapnpAdapter(lis net.Listener, srv certs.ICerts) error {

	logrus.Infof("Starting cert service capnp adapter on: %s", lis.Addr())

	// Create the capnp proxy that provides the certificate capability
	main := hubapi.CapCertService_ServerToClient(&CertServiceCapnpAdapter{
		srv: srv,
	})
	//
	//// Create the capnp proxy that provides the capability to create device certificates
	//deviceCap := hubapi.CapDeviceCert_ServerToClient(&CertServiceCapnpAdapter{
	//	srv: srv,
	//})

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
	// serve the requests
	err := caphelp.CapServe(ctx, lis, capnp.Client(main))
	ctxCancel()
	return err
}
