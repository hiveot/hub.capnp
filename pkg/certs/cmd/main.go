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
	"github.com/hiveot/hub/pkg/certs/capnpserver"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
)

const ServiceName = "certs"

// Start the certs service
//  commandline options:
//  --certs <certificate folder>
func main() {
	var caCert *x509.Certificate
	var caKey *ecdsa.PrivateKey
	var err error

	var certFolder = folders.GetFolders("").Certs
	flag.StringVar(&certFolder, "certs", certFolder, "Certificate folder.")

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

	// check commandline and create a listener
	lis := listener.CreateServiceListener(ServiceName)
	_ = lis

	logrus.Infof("CertServiceCapnpServer starting on %s", lis.Addr())
	svc := selfsigned.NewSelfSignedCertsService(caCert, caKey)
	_ = capnpserver.StartCertsCapnpServer(context.Background(), lis, svc)
}
