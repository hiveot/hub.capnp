package selfsigned

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/pkg/certs"
)

// DeviceCertsService creates device certificates for use by IoT devices.
// This implements the IDeviceCerts interface
// Issued certificates are short-lived and must be renewed before they expire.
type DeviceCertsService struct {
	caCert    *x509.Certificate
	caCertPEM string
	//caCertPool *x509.CertPool
	caKey *ecdsa.PrivateKey
}

// _createDeviceCert internal function to create a CA signed certificate for mutual authentication by IoT devices
func (srv *DeviceCertsService) _createDeviceCert(
	deviceID string, pubKey *ecdsa.PublicKey, validityDays int) (
	cert *x509.Certificate, err error) {
	if validityDays == 0 {
		validityDays = certs.DefaultDeviceCertValidityDays
	}

	cert, err = createClientCert(
		deviceID,
		certsclient.OUIoTDevice,
		pubKey,
		srv.caCert,
		srv.caKey,
		validityDays)

	// TODO: send Thing event (services are things too)
	return cert, err
}

// CreateDeviceCert creates a CA signed certificate for mutual authentication by IoT devices in PEM format
func (srv *DeviceCertsService) CreateDeviceCert(
	_ context.Context, deviceID string, pubKeyPEM string, durationDays int) (
	certPEM string, caCertPEM string, err error) {
	var cert *x509.Certificate

	logrus.Infof("deviceID='%s' pubKey='%s'", deviceID, pubKeyPEM)
	pubKey, err := certsclient.PublicKeyFromPEM(pubKeyPEM)
	if err != nil {
		err = fmt.Errorf("public key for '%s' is invalid: %s", deviceID, err)
	} else {
		cert, err = srv._createDeviceCert(deviceID, pubKey, durationDays)
	}
	if err == nil {
		certPEM = certsclient.X509CertToPEM(cert)
	}
	return certPEM, srv.caCertPEM, err
}

// Release the provided capability and release resources
func (srv *DeviceCertsService) Release() {
	// nothing to do here
}

// NewDeviceCertsService returns a new instance of the selfsigned device certificate service
//
//	caCert is the CA certificate used to created certificates
//	caKey is the CA private key used to created certificates
func NewDeviceCertsService(caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *DeviceCertsService {
	//caCertPool := x509.NewCertPool()
	//caCertPool.AddCert(caCert)

	service := &DeviceCertsService{
		caCert:    caCert,
		caKey:     caKey,
		caCertPEM: certsclient.X509CertToPEM(caCert),
		//caCertPool: caCertPool,
	}
	if caCert == nil || caKey == nil || caCert.PublicKey == nil {
		logrus.Panic("Missing CA certificate or key")
	}

	return service
}
