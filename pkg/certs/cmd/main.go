package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"net"
	"path"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/certs/capnpserver"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
)

// Connect the certs service
//
//	commandline options:
//	--certs <certificate folder>
func main() {
	var caCert *x509.Certificate
	var caKey *ecdsa.PrivateKey
	var err error

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

	svc := selfsigned.NewSelfSignedCertsService(caCert, caKey)

	listener.RunService(certs.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err = svc.Start()
			if err == nil {
				err = capnpserver.StartCertsCapnpServer(svc, lis)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}
