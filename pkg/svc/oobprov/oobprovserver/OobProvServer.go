// Package oobprov with out-of-band provisioning service
package oobprovserver

import (
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"

	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost.grpc/go/svc"
)

// OobProvServer implements the svc.Provisioning interface using out-of-band provisiong
// Provisioning secrets are kept in-memory until the service is restarted.
type OobProvServer struct {
	svc.UnimplementedProvisioningServer
	caCert *x509.Certificate
	caKey  *ecdsa.PrivateKey
}

// NewOobProvServer creates a service instance to automatically provision IoT devices
func NewOobProvServer(caCertPath string, caKeyPath string) (*OobProvServer, error) {
	var caKey *ecdsa.PrivateKey
	fmt.Println("Loading cert from ", caCertPath)
	caCert, err := certsclient.LoadX509CertFromPEM(caCertPath)
	if err == nil {
		caKey, err = certsclient.LoadKeysFromPEM(caKeyPath)
	}

	srv := &OobProvServer{
		caCert: caCert,
		caKey:  caKey,
	}
	return srv, err
}
