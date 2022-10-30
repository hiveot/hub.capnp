package selfsigned

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub/pkg/certs"
)

// UserCertsService creates certificates for use by end-users
// This implements the IUserCerts interface
// Issued certificates are short-lived and must be renewed before they expire.
type UserCertsService struct {
	caCert    *x509.Certificate
	caCertPEM string
	caKey     *ecdsa.PrivateKey
}

// _createUserCert internal function to create a client certificate for end-users
func (srv *UserCertsService) _createUserCert(userID string, pubKey *ecdsa.PublicKey, validityDays int) (
	cert *x509.Certificate, err error) {
	if validityDays == 0 {
		validityDays = certs.DefaultUserCertValidityDays
	}

	cert, err = createClientCert(
		userID,
		certsclient.OUUser,
		pubKey,
		srv.caCert,
		srv.caKey,
		validityDays)
	// TODO: send Thing event (services are things too)
	return cert, err
}

// CreateUserCert creates a client certificate for end-users
func (srv *UserCertsService) CreateUserCert(
	_ context.Context, userID string, pubKeyPEM string, validityDays int) (
	certPEM string, caCertPEM string, err error) {
	var cert *x509.Certificate

	logrus.Infof("userID='%s' pubKey='%s'", userID, pubKeyPEM)
	pubKey, err := certsclient.PublicKeyFromPEM(pubKeyPEM)
	if err == nil {

		cert, err = srv._createUserCert(
			userID,
			pubKey,
			validityDays)
	}
	if err == nil {
		certPEM = certsclient.X509CertToPEM(cert)
	}

	// TODO: send Thing event (services are things too)
	return certPEM, srv.caCertPEM, err
}

// Release the provided capability and release resources
func (srv *UserCertsService) Release() {
	// nothing to do here
}

// NewUserCertsService returns a new instance of the selfsigned user certificate management service
//  caCert is the CA certificate used to created certificates
//  caKey is the CA private key used to created certificates
func NewUserCertsService(caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *UserCertsService {
	service := &UserCertsService{
		caCert:    caCert,
		caKey:     caKey,
		caCertPEM: certsclient.X509CertToPEM(caCert),
	}
	if caCert == nil || caKey == nil || caCert.PublicKey == nil {
		logrus.Panic("Missing CA certificate or key")
	}

	return service
}
