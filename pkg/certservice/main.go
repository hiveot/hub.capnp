package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"flag"
	"path"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/certsclient"

	"github.com/hiveot/hub/pkg/certservice/adapter"
	"github.com/hiveot/hub/pkg/certservice/selfsigned"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
)

const ServiceName = "certs"

// DefaultCaCertPath is the path to the CA Certificate in PEM format
const DefaultCaCertPath = "config/caCert.pem"

// DefaultCaKeyPath is the path to the CA Certificate Private key
const DefaultCaKeyPath = "config/caKey.pem"

// Start the certs service
//  commandline options:
//  --certs <certificate folder>
func main() {
	var caCert *x509.Certificate
	var caKey *ecdsa.PrivateKey
	var err error

	var certFolder = folders.GetFolders("").Certs
	flag.StringVar(&certFolder, "certs", certFolder, "Certificate folder.")

	// handle commandline to create a listener
	lis := listener.CreateServiceListener(ServiceName)
	_ = lis

	// This service needs the CA certificate and key to operate
	caCertPath := path.Join(certFolder, hubapi.DefaultCaCertFile)
	caKeyPath := path.Join(certFolder, hubapi.DefaultCaKeyFile)

	logrus.Infof("Loading CA certificate and key from %s", certFolder)
	caCert, err = certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Fatalf("Error loading CA certificate from '%s': %v", caCertPath, err)
	}
	caKey, err = certsclient.LoadKeysFromPEM(caKeyPath)
	if err != nil {
		logrus.Fatalf("Error loading CA key from '%s': %v", caKeyPath, err)
	}

	logrus.Infof("CertServiceCapnpAdapter starting")
	service := selfsigned.NewSelfSignedServer(caCert, caKey)
	adapter.StartCertServiceCapnpAdapter(context.Background(), lis, service)
}
