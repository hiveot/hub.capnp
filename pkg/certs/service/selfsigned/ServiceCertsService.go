package selfsigned

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/pkg/certs"
)

// ServiceCertsService creates certificates for use by services.
// This implements the IServiceCerts interface
// Issued certificates are short-lived and must be renewed before they expire.
type ServiceCertsService struct {
	caCert    *x509.Certificate
	caCertPEM string
	//caCertPool *x509.CertPool
	caKey *ecdsa.PrivateKey
}

// createServiceCert internal function to create a CA signed service certificate for mutual authentication between services
func (srv *ServiceCertsService) _createServiceCert(
	serviceID string, servicePubKey *ecdsa.PublicKey, names []string, validityDays int) (
	cert *x509.Certificate, err error) {

	if serviceID == "" || servicePubKey == nil || names == nil {
		err := fmt.Errorf("missing argument serviceID, servicePubKey, or names")
		logrus.Error(err)
		return nil, err
	}
	if validityDays == 0 {
		validityDays = certs.DefaultServiceCertValidityDays
	}

	// firefox complains if serial is the same as that of the CA. So generate a unique one based on timestamp.
	serial := time.Now().Unix() - 3
	template := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{
			Country:            []string{"CA"},
			Province:           []string{"BC"},
			Locality:           []string{CertOrgLocality},
			Organization:       []string{CertOrgName},
			OrganizationalUnit: []string{certsclient.OUService},
			CommonName:         serviceID,
		},
		NotBefore: time.Now().Add(-time.Second),
		NotAfter:  time.Now().AddDate(0, 0, validityDays),
		//NotBefore: time.Now(),
		//NotAfter:  time.Now().AddDate(0, 0, config.DefaultServiceCertDurationDays),

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		//ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:           false,
		MaxPathLenZero: true,
		// BasicConstraintsValid: true,
		// IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		IPAddresses: []net.IP{},
	}
	// determine the hosts for this hub
	for _, h := range names {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}
	// Create the service private key
	//certKey := certsclient.CreateECDSAKeys()
	// and the certificate itself
	certDer, err := x509.CreateCertificate(rand.Reader, template,
		srv.caCert, servicePubKey, srv.caKey)
	if err == nil {
		cert, err = x509.ParseCertificate(certDer)
	}

	// TODO: send Thing event (services are things too)
	return cert, err
}

// CreateServiceCert creates a CA signed service certificate for mutual authentication between services
func (srv *ServiceCertsService) CreateServiceCert(
	_ context.Context, serviceID string, pubKeyPEM string, names []string, validityDays int) (
	certPEM string, caCertPEM string, err error) {
	var cert *x509.Certificate

	logrus.Infof("Creating service certificate: serviceID='%s', names='%s'", serviceID, names)
	pubKey, err := certsclient.PublicKeyFromPEM(pubKeyPEM)
	if err == nil {
		cert, err = srv._createServiceCert(
			serviceID,
			pubKey,
			names,
			validityDays,
		)
	}
	if err == nil {
		certPEM = certsclient.X509CertToPEM(cert)
	}
	// TODO: send Thing event (services are things too)
	return certPEM, srv.caCertPEM, err
}

// Release the provided capability and release resources
func (srv *ServiceCertsService) Release() {
	// nothing to do here
}

// NewServiceCertsService returns a new instance of the selfsigned service certificate service
//
//	caCert is the CA certificate used to create certificates
//	caKey is the CA private key used to create certificates
func NewServiceCertsService(caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *ServiceCertsService {
	if caCert == nil || caKey == nil || caCert.PublicKey == nil {
		logrus.Fatal("Missing CA certificate or key")
	}

	service := &ServiceCertsService{
		caCert:    caCert,
		caKey:     caKey,
		caCertPEM: certsclient.X509CertToPEM(caCert),
		//caCertPool: caCertPool,
	}

	return service
}
