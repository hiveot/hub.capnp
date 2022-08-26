package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"flag"
	"log"
	"path"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/wostzone/hub/internal/folders"
	"github.com/wostzone/hub/internal/listener"
	"github.com/wostzone/hub/pkg/svc/certsvc/certconfig"
	"github.com/wostzone/hub/pkg/svc/certsvc/selfsigned"
	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost.grpc/go/svc"
)

const ServiceName = "certsvc"

// DefaultCaCertPath is the path to the CA Certificate in PEM format
const DefaultCaCertPath = "config/cacert.pem"

// DefaultCaKeyPath is the path to the CA Certificate Private key
const DefaultCaKeyPath = "config/cakey.pem"

// Start the service using gRPC
// This service issues certificates signed by the CA.
func main() {
	var caCert *x509.Certificate
	var caKey *ecdsa.PrivateKey
	var err error

	var certFolder = folders.GetFolders("").Certs
	flag.StringVar(&certFolder, "certs", certFolder, "Certificate folder.")

	// handle commandline to create a listener
	lis := listener.CreateServiceListener(ServiceName)

	caCertPath := path.Join(certFolder, certconfig.DefaultCaCertFile)
	caKeyPath := path.Join(certFolder, certconfig.DefaultCaKeyFile)

	// This service needs the CA certificate and key to operate
	logrus.Infof("Loading CA certificate and key from %s", certFolder)
	caCert, err = certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Fatalf("Error loading CA certificate from '%s': %v", caCertPath, err)
	}
	caKey, err = certsclient.LoadKeysFromPEM(caKeyPath)
	if err != nil {
		logrus.Fatalf("Error loading CA key from '%s': %v", caKeyPath, err)
	}

	//
	svcConfig := certconfig.CertSvcConfig{
		CaCert: caCert,
		CaKey:  caKey,
	}
	s := grpc.NewServer()
	service := selfsigned.NewSelfSignedServer(svcConfig)
	svc.RegisterCertServiceServer(s, service)

	// exit the service when signal is received and close the listener
	listener.ExitOnSignal(lis, func() {
		logrus.Infof("Shutting down '%s'", ServiceName)
	})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Service '%s; exited: %v", ServiceName, err)
	}
}
