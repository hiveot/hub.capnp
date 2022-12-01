package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"path"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/certs/capnpserver"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
)

// Start the certs service
//
//	commandline options:
//	--certs <certificate folder>
func main() {
	var caCert *x509.Certificate
	var caKey *ecdsa.PrivateKey
	var err error

	logging.SetLogging("info", "")

	// Determine the folder layout and handle commandline options
	f := svcconfig.LoadServiceConfig(certs.ServiceName, false, nil)

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
	_ = capnpserver.StartCertsCapnpServer(srvListener, svc)
}
