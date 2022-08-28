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

// CreateServiceCert creates a new Hub service certificate for mutual authentication between services.
//
// The certificate is valid for the given names either local domain name and IP addresses.
// The service must have a fixed IP.
//  serviceID identifies the service. Used in certificate CommonName.
//  names contains one or more domain names and/or IP addresses the Hub can be reached on, to add to the certificate
//  pubKeyPEM public key of the service to sign
//  caCert is the CA to sign the server certificate
//  caPrivKey is the CA private key to sign the server certificate
// returns the signed public certificate
func CreateServiceCert(
	serviceID string, names []string, publicKey *ecdsa.PublicKey,
	caCert *x509.Certificate, caPrivKey *ecdsa.PrivateKey,
	durationDays int) (cert *x509.Certificate, err error) {

	if caCert == nil || caPrivKey == nil || names == nil {
		err := fmt.Errorf("CreateServiceCert: missing argument")
		logrus.Error(err)
		return nil, err
	} else if caCert.PublicKey == nil {
		err := fmt.Errorf("CreateServiceCert: CA cert has no public key")
		logrus.Error(err)
		return nil, err
	}

	logrus.Infof("CertSetup.CreateServiceCert: Refresh server certificate for IP/name: %s", names)
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
		NotAfter:  time.Now().AddDate(0, 0, durationDays),
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
	certDer, err := x509.CreateCertificate(rand.Reader, template, caCert,
		publicKey, caPrivKey)
	if err != nil {
		return nil, err
	}
	newCert, err := x509.ParseCertificate(certDer)
	//// combined them into a TLS certificate
	//tlscert := &tls.Certificate{}
	//tlscert.Certificate = append(tlscert.Certificate, certDer)
	//tlscert.PrivateKey = certKey

	return newCert, err
}
