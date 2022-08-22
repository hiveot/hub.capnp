package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"flag"
	"log"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost.grpc/go/svc"
	"svc/certsvc/config"
	"svc/certsvc/selfsigned"
	"svc/internal/listener"
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
	caCertPath := DefaultCaCertPath
	caKeyPath := DefaultCaKeyPath

	// Add commandline option '--cacert  with CA certificate path
	flag.StringVar(&caCertPath, "cacert", caCertPath, "Path to CA certificate")
	// Add commandline option '--cakey with CA private key for issuing new certificates
	flag.StringVar(&caKeyPath, "c", caKeyPath, "Path to CA private key")

	// handle commandline to create a listener
	lis := listener.CreateServiceListener(ServiceName)

	// This service needs the CA certificate and key to operate
	caCert, err = certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Fatalf("Error loading CA certificate from '%s': %v", caCertPath, err)
	}
	caKey, err = certsclient.LoadKeysFromPEM(caKeyPath)
	if err != nil {
		logrus.Fatalf("Error loading CA key from '%s': %v", caKeyPath, err)
	}

	//
	svcConfig := config.CertSvcConfig{
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
