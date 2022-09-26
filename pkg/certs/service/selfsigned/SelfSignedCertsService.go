package selfsigned

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/certsclient"
)

// SelfSignedCertsService creates certificates for use by services, devices and admin users.
//
// This implements the ICerts, IDeviceCerts and IVerifyCert capabilities
//
// Note that this service does not support certificate revocation.
//   See also: https://www.imperialviolet.org/2014/04/19/revchecking.html
// Issued certificates are short-lived and must be renewed before they expire.
type SelfSignedCertsService struct {
	caCert     *x509.Certificate
	caCertPEM  string
	caCertPool *x509.CertPool
	caKey      *ecdsa.PrivateKey
}

// createClientCert is the internal function to create a client certificate
// for IoT devices, administrator
//
// The client ouRole is intended to for role based authorization. It is stored in the
// certificate OrganizationalUnit. See OUxxx
//
// This generates a TLS client certificate with keys
//  clientID used as the CommonName, eg pluginID or deviceID
//  ouRole with type of client: OUNone, OUAdmin, OUClient, OUIoTDevice
//  ownerPubKey the public key of the certificate holder
//  caCert CA's certificate for signing
//  caPrivKey CA's ECDSA key for signing
//  validityDays nr of days the certificate will be valid
// Returns the signed certificate with the corresponding CA used to sign, or an error
func (srv *SelfSignedCertsService) createClientCert(
	clientID string, ouRole string, ownerPubKey *ecdsa.PublicKey, validityDays int) (
	clientCert *x509.Certificate, err error) {

	var newCert *x509.Certificate

	if clientID == "" || ownerPubKey == nil {
		err := fmt.Errorf("missing clientID or client public key")
		logrus.Error(err)
		return nil, err
	}
	// firefox complains if serial is the same as that of the CA. So generate a unique one based on timestamp.
	serial := time.Now().Unix() - 2
	template := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{
			Country:            []string{"CA"},
			Province:           []string{"BC"},
			Locality:           []string{CertOrgLocality},
			Organization:       []string{CertOrgName},
			OrganizationalUnit: []string{ouRole},
			CommonName:         clientID,
			Names:              make([]pkix.AttributeTypeAndValue, 0),
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(0, 0, validityDays),

		//KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment,
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},

		BasicConstraintsValid: true,
		IsCA:                  false,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	// clientKey := certs.CreateECDSAKeys()
	certDer, err := x509.CreateCertificate(rand.Reader, template, srv.caCert, ownerPubKey, srv.caKey)
	if err == nil {
		newCert, err = x509.ParseCertificate(certDer)
	}

	// // combined them into a TLS certificate
	// tlscert := &tls.Certificate{}
	// tlscert.Certificate = append(tlscert.Certificate, certDer)
	// tlscert.PrivateKey = clientKey

	return newCert, err
}

// _createDeviceCert internal function to create a CA signed certificate for mutual authentication by IoT devices
func (srv *SelfSignedCertsService) _createDeviceCert(deviceID string, pubKey *ecdsa.PublicKey, validityDays int) (
	cert *x509.Certificate, err error) {

	cert, err = srv.createClientCert(
		deviceID,
		certsclient.OUIoTDevice,
		pubKey,
		validityDays)

	// TODO: send Thing event (services are things too)
	return cert, err
}

// CreateDeviceCert creates a CA signed certificate for mutual authentication by IoT devices in PEM format
func (srv *SelfSignedCertsService) CreateDeviceCert(deviceID string, pubKeyPEM string, durationDays int) (
	certPEM string, caCertPEM string, err error) {
	var cert *x509.Certificate

	logrus.Infof("deviceID='%s' pubKey='%s'", deviceID, pubKeyPEM)
	pubKey, err := certsclient.PublicKeyFromPEM(pubKeyPEM)
	if err == nil {
		cert, err = srv._createDeviceCert(deviceID, pubKey, durationDays)
	}
	if err == nil {
		certPEM = certsclient.X509CertToPEM(cert)
	}
	return certPEM, srv.caCertPEM, err
}

// createServiceCert internal function to create a CA signed service certificate for mutual authentication between services
func (srv *SelfSignedCertsService) _createServiceCert(
	serviceID string, servicePubKey *ecdsa.PublicKey, names []string, validityDays int) (
	cert *x509.Certificate, err error) {

	if serviceID == "" || servicePubKey == nil || names == nil {
		err := fmt.Errorf("missing argument serviceID, servicePubKey, or names")
		logrus.Error(err)
		return nil, err
	}

	logrus.Infof("Create service certificate for IP/name: %s", names)
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
		NotBefore: time.Now(),
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
func (srv *SelfSignedCertsService) CreateServiceCert(serviceID string, pubKeyPEM string, names []string, validityDays int) (
	certPEM string, caCertPEM string, err error) {
	var cert *x509.Certificate

	logrus.Infof("serviceID='%s' pubKey='%s', names='%s'", serviceID, pubKeyPEM, names)
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

// _createUserCert internal function to create a client certificate for end-users
func (srv *SelfSignedCertsService) _createUserCert(userID string, pubKey *ecdsa.PublicKey, validityDays int) (
	cert *x509.Certificate, err error) {

	cert, err = srv.createClientCert(
		userID,
		certsclient.OUUser,
		pubKey,
		validityDays)
	// TODO: send Thing event (services are things too)
	return cert, err
}

// CreateUserCert creates a client certificate for end-users
func (srv *SelfSignedCertsService) CreateUserCert(userID string, pubKeyPEM string, validityDays int) (
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

// VerifyCert verifies whether the given certificate is a valid client certificate
func (srv *SelfSignedCertsService) VerifyCert(clientID string, certPEM string) error {
	opts := x509.VerifyOptions{
		Roots:     srv.caCertPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	cert, err := certsclient.X509CertFromPEM(certPEM)
	if cert.Subject.CommonName != clientID {
		err = fmt.Errorf("client ID '%s' doesn't match certificate name '%s'", clientID, cert.Subject.CommonName)
	}
	//if err == nil {
	//	x509Cert, err := x509.ParseCertificate(clientCert.Certificate[0])
	//}
	if err == nil {
		// FIXME: TestCertAuth: certificate specifies incompatible key usage
		// why? Is the certpool invalid? Yet the test succeeds
		_, err = cert.Verify(opts)
	}
	return err
}

// NewSelfSignedCertsService returns a new instance of the selfsigned certificate service
//  caCert is the CA certificate used to created certificates
//  caKey is the CA private key used to created certificates
func NewSelfSignedCertsService(caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *SelfSignedCertsService {
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	service := &SelfSignedCertsService{
		caCert:     caCert,
		caKey:      caKey,
		caCertPEM:  certsclient.X509CertToPEM(caCert),
		caCertPool: caCertPool,
	}
	if caCert == nil || caKey == nil || caCert.PublicKey == nil {
		logrus.Panic("Missing CA certificate or key")
	}

	return service
}
