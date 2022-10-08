package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"flag"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/certs/capnpserver"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
)

// Start the certs service
//  commandline options:
//  --certs <certificate folder>
func main() {
	var caCert *x509.Certificate
	var caKey *ecdsa.PrivateKey
	var err error

	logging.SetLogging("info", "")

	// this is a service so go 2 levels up
	// FIXME: import the folder structure instead of hard coding it
	homeFolder := filepath.Join(filepath.Dir(os.Args[0]), "../..")
	f := folders.GetFolders(homeFolder, false)
	flag.StringVar(&f.Certs, "certs", f.Certs, "Certificate folder.")
	flag.Parse()

	// This service needs the CA certificate and key to operate
	caCertPath := path.Join(f.Certs, hubapi.DefaultCaCertFile)
	caKeyPath := path.Join(f.Certs, hubapi.DefaultCaKeyFile)

	logrus.Infof("Loading CA certificate and key from %s", f.Certs)
	caCert, err = certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Fatalf("Error loading CA certificate from '%s': %v", caCertPath, err)
	}
	caKey, err = certsclient.LoadKeysFromPEM(caKeyPath)
	if err != nil {
		logrus.Fatalf("Error loading CA key from '%s': %v", caKeyPath, err)
	}

	// check commandline and create a listener
	srvListener := listener.CreateServiceListener(f.Run, certs.ServiceName)

	logrus.Infof("CertServiceCapnpServer starting on: %s", srvListener.Addr())
	svc := selfsigned.NewSelfSignedCertsService(caCert, caKey)
	_ = capnpserver.StartCertsCapnpServer(context.Background(), srvListener, svc)
}
