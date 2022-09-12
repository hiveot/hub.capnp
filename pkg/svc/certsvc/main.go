package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"flag"
	"log"
	"path"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.grpc/go/svc"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/svc/certsvc/selfsigned"
	"github.com/hiveot/hub/pkg/svc/certsvc/service"
)

const ServiceName = "certsvc"

// DefaultCaCertPath is the path to the CA Certificate in PEM format
const DefaultCaCertPath = "config/cacert.pem"

// DefaultCaKeyPath is the path to the CA Certificate Private key
const DefaultCaKeyPath = "config/cakey.pem"

// Start the cert service using gRPC (or capnproto)
// This service issues certificates signed by the CA.
func main() {
	var caCert *x509.Certificate
	var caKey *ecdsa.PrivateKey
	var useCapnproto = true // experiment with capnproto for rpc
	var err error

	var certFolder = folders.GetFolders("").Certs
	flag.StringVar(&certFolder, "certs", certFolder, "Certificate folder.")

	// handle commandline to create a listener
	lis := listener.CreateServiceListener(ServiceName)
	_ = lis

	// This service needs the CA certificate and key to operate
	caCertPath := path.Join(certFolder, service.DefaultCaCertFile)
	caKeyPath := path.Join(certFolder, service.DefaultCaKeyFile)

	logrus.Infof("Loading CA certificate and key from %s", certFolder)
	caCert, err = certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Fatalf("Error loading CA certificate from '%s': %v", caCertPath, err)
	}
	caKey, err = certsclient.LoadKeysFromPEM(caKeyPath)
	if err != nil {
		logrus.Fatalf("Error loading CA key from '%s': %v", caKeyPath, err)
	}

	if useCapnproto {
		// test capnproto
		logrus.Infof("ServeCertServiceCapnpAdapter started")
		err = ServeCertServiceCapnpAdapter(
			context.Background(),
			selfsigned.NewSelfSignedServer(caCert, caKey),
			lis)

		if err != nil {
			logrus.Fatalf("ServeCertServiceCapnpAdapter failed: %s", err)
		} else {
			logrus.Infof("ServeCertServiceCapnpAdapter ended")
		}

	} else {
		// use grpc service adapter to handle grpc requests
		s := grpc.NewServer()
		grpcAdapter := &CertServerGRPCAdapter{
			srv: selfsigned.NewSelfSignedServer(caCert, caKey),
		}
		svc.RegisterCertServiceServer(s, grpcAdapter)

		// exit the service when signal is received and close the listener
		listener.ExitOnSignal(lis, func() {
			logrus.Infof("Shutting down '%s'", ServiceName)
		})

		logrus.Infof("listening on %s", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Service '%s; exited: %v", ServiceName, err)
		}
	}

}
