package selfsigned

import (
	"crypto/ecdsa"
	"crypto/x509"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub/pkg/svc/certsvc/service"
)

// SelfSignedCertService creates certificates for use by services, devices and admin users.
// Note that this service does not support certificate revocation.
//   See also: https://www.imperialviolet.org/2014/04/19/revchecking.html
// Instead the issued certificates are short lived and must be renewed before they expire.
type SelfSignedCertService struct {
	caCert *x509.Certificate
	caKey  *ecdsa.PrivateKey
}

// CreateClientCert creates a CA signed certificate for mutual authentication by consumers
func (srv *SelfSignedCertService) CreateClientCert(clientID string, pubKeyPEM string) (
	certPEM string, caCertPEM string, err error) {

	pubKey, err := certsclient.PublicKeyFromPEM(pubKeyPEM)
	if err != nil {
		return "", "", err
	}

	cert, err := CreateClientCert(
		clientID,
		certsclient.OUClient,
		pubKey,
		srv.caCert,
		srv.caKey,
		service.DefaultClientCertDurationDays)

	caCertPEM = certsclient.X509CertToPEM(srv.caCert)
	certPEM = certsclient.X509CertToPEM(cert)
	return certPEM, caCertPEM, err
}

// CreateDeviceCert creates a CA signed certificate for mutual authentication by IoT devices
func (srv *SelfSignedCertService) CreateDeviceCert(clientID string, pubKeyPEM string) (
	certPEM string, caCertPEM string, err error) {

	var cert *x509.Certificate
	pubKey, err := certsclient.PublicKeyFromPEM(pubKeyPEM)
	if err == nil {
		cert, err = CreateClientCert(
			clientID,
			certsclient.OUIoTDevice,
			pubKey,
			srv.caCert,
			srv.caKey,
			service.DefaultDeviceCertDurationDays)

		caCertPEM = certsclient.X509CertToPEM(srv.caCert)
		certPEM = certsclient.X509CertToPEM(cert)
	}
	return certPEM, caCertPEM, err
}

// CreateServiceCert creates a CA signed service certificate for mutual authentication between services
func (srv *SelfSignedCertService) CreateServiceCert(serviceID string, pubKeyPEM string, names []string) (
	certPEM string, caCertPEM string, err error) {
	var cert *x509.Certificate

	pubKey, err := certsclient.PublicKeyFromPEM(pubKeyPEM)
	if err == nil {
		cert, err = CreateServiceCert(
			serviceID,
			names,
			pubKey,
			srv.caCert,
			srv.caKey,
			service.DefaultServiceCertDurationDays,
		)

		caCertPEM = certsclient.X509CertToPEM(srv.caCert)
		certPEM = certsclient.X509CertToPEM(cert)
	}
	return certPEM, caCertPEM, err
}

func NewSelfSignedServer(caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *SelfSignedCertService {
	service := &SelfSignedCertService{
		caCert: caCert,
		caKey:  caKey,
	}
	return service
}
