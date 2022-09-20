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
)

// CreateClientCert creates a hub client certificate for mutual authentication from client's public key
// The client role is intended to for role based authorization. It is stored in the
// certificate OrganizationalUnit. See OUxxx
//
// This generates a TLS client certificate with keys
//  clientID used as the CommonName, eg pluginID or deviceID
//  ou with type of client: OUNone, OUAdmin, OUClient, OUIoTDevice
//  ownerPubKey the public key of the certificate holder
//  caCert CA's certificate for signing
//  caPrivKey CA's ECDSA key for signing
//  durationDays nr of days the certificate will be valid
// Returns the signed TLS certificate or error
func CreateClientCert(clientID string,
	ou string,
	ownerPubKey *ecdsa.PublicKey,
	caCert *x509.Certificate,
	caPrivKey *ecdsa.PrivateKey,
	validityDays int) (clientCert *x509.Certificate, err error) {

	if caCert == nil || caPrivKey == nil {
		err := fmt.Errorf("CreateHubClientCert: missing CA cert or key")
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
			OrganizationalUnit: []string{ou},
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
	certDer, err := x509.CreateCertificate(rand.Reader, template, caCert, ownerPubKey, caPrivKey)
	if err != nil {
		logrus.Errorf("CertSetup.CreateHubClientCert: Unable to create HiveOT Hub client cert: %s", err)
		return nil, err
	}
	newCert, err := x509.ParseCertificate(certDer)

	// // combined them into a TLS certificate
	// tlscert := &tls.Certificate{}
	// tlscert.Certificate = append(tlscert.Certificate, certDer)
	// tlscert.PrivateKey = clientKey

	return newCert, err
}
