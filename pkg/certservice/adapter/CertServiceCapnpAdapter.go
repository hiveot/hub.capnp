// Package adapter with the capnproto adapter
package adapter

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/certservice/selfsigned"
)

// CertServiceCapnpAdapter implements the capnproto generated interface CertsService_Server
// See hub.capnp/go/hubapi/Cert.capnp.go for the interface
type CertServiceCapnpAdapter struct {
	srv *selfsigned.SelfSignedCertService
}

func (adr *CertServiceCapnpAdapter) CreateClientCert(
	ctx context.Context, call hubapi.CertService_createClientCert) error {

	clientID, _ := call.Args().ClientID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	validityDays := call.Args().ValidityDays()
	if validityDays == 0 {
		validityDays = hubapi.DefaultClientCertValidityDays
	}
	certPEM, caCertPEM, err := adr.srv.CreateClientCert(clientID, pubKeyPEM, int(validityDays))
	if err == nil {
		logrus.Infof("Created client cert for %s", clientID)
		res, err2 := call.AllocResults()
		res.SetCertPEM(certPEM)
		res.SetCaCertPEM(caCertPEM)
		err = err2
	}
	return err
}

func (adr *CertServiceCapnpAdapter) CreateDeviceCert(
	_ context.Context, call hubapi.CertService_createDeviceCert) error {
	clientID, _ := call.Args().DeviceID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	validityDays := call.Args().ValidityDays()
	if validityDays == 0 {
		validityDays = hubapi.DefaultClientCertValidityDays
	}
	certPEM, caCertPEM, err := adr.srv.CreateDeviceCert(clientID, pubKeyPEM, int(validityDays))
	if err == nil {
		logrus.Infof("Created device cert for %s", clientID)
		res, err2 := call.AllocResults()
		res.SetCertPEM(certPEM)
		res.SetCaCertPEM(caCertPEM)
		err = err2
	}
	return err
}

func (adr *CertServiceCapnpAdapter) CreateServiceCert(
	_ context.Context, call hubapi.CertService_createServiceCert) error {
	clientID, _ := call.Args().ServiceID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	namesList, _ := call.Args().Names()
	names := []string{}
	for i := 0; i < namesList.Len(); i++ {
		name, _ := namesList.At(i)
		names = append(names, name)
	}
	certPEM, caCertPEM, err := adr.srv.CreateServiceCert(clientID, pubKeyPEM, names, int(hubapi.DefaultServiceCertValidityDays))
	if err == nil {
		logrus.Infof("Created device cert for %s", clientID)
		res, err2 := call.AllocResults()
		res.SetCertPEM(certPEM)
		res.SetCaCertPEM(caCertPEM)
		err = err2
	}
	return err
}

// StartCertServiceCapnpAdapter starts the certificate service capnp protocol server
func StartCertServiceCapnpAdapter(ctx context.Context,
	lis net.Listener,
	srv *selfsigned.SelfSignedCertService) error {

	// Create the capnp client to receive requests
	main := hubapi.CertService_ServerToClient(&CertServiceCapnpAdapter{
		srv: srv,
	})
	return caphelp.CapServe(ctx, lis, capnp.Client(main))
}
