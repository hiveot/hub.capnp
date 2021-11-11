package testenv

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/certs"
)

// FIXME: These OU names must match that in certsetup
const OUPlugin = "plugin"
const OUDevice = "iotdevice"
const ServerAddress = "127.0.0.1"

// Test certificates for CA, server and plugin (client)
type TestCerts struct {
	CaCert    *x509.Certificate
	CaKey     *ecdsa.PrivateKey
	ServerKey *ecdsa.PrivateKey
	DeviceKey *ecdsa.PrivateKey
	PluginKey *ecdsa.PrivateKey

	ServerCert *tls.Certificate
	PluginCert *tls.Certificate
	DeviceCert *tls.Certificate
}

// CreateCertBundle creates new certificates for CA, Server, Plugin and Thing Device testing
// The server cert is valid for localhost only
//  this returns the x509 and tls certificates
func CreateCertBundle() TestCerts {
	testCerts := TestCerts{}
	testCerts.CaCert, testCerts.CaKey = CreateCA()
	testCerts.ServerKey = certs.CreateECDSAKeys()
	testCerts.PluginKey = certs.CreateECDSAKeys()
	testCerts.DeviceKey = certs.CreateECDSAKeys()
	testCerts.ServerCert = CreateTlsCert("Server", "wost", true,
		testCerts.ServerKey, testCerts.CaCert, testCerts.CaKey)
	testCerts.PluginCert = CreateTlsCert("Plugin", OUPlugin, false,
		testCerts.PluginKey, testCerts.CaCert, testCerts.CaKey)
	testCerts.DeviceCert = CreateTlsCert("Device", OUDevice, false,
		testCerts.DeviceKey, testCerts.CaCert, testCerts.CaKey)
	return testCerts
}

// CreateCA generates the CA keys with certificate for testing
// not intended for production
func CreateCA() (caCert *x509.Certificate, caKey *ecdsa.PrivateKey) {
	validity := time.Hour

	caKey = certs.CreateECDSAKeys()

	// set up our CA certificate
	// see also: https://superuser.com/questions/738612/openssl-ca-keyusage-extension
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Country:      []string{"CA"},
			Organization: []string{"Testing"},
			Province:     []string{"BC"},
			Locality:     []string{"wostzone"},
			CommonName:   "WoST CA",
		},
		NotBefore: time.Now().Add(-10 * time.Second),
		NotAfter:  time.Now().Add(validity),
		// CA cert can be used to sign certificate and revocation lists
		KeyUsage:    x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},

		// This hub cert is the only CA. No intermediate CAs
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	// create the CA
	certDerBytes, _ := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
	// certPEMBuffer := new(bytes.Buffer)
	// pem.Encode(certPEMBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes})

	caCert, _ = x509.ParseCertificate(certDerBytes)
	return caCert, caKey
}

// CreateTlsCert generates the certificate with keys, signed by the CA, valid for 127.0.0.1
// intended for testing, not for production
//  cn is the certificate common name, usually the client ID or server hostname
//  ou the organization
//  isServer if set allow key usage of ServerAuth instead of ClientAuth
//  clientKey is the client's private key for this certificate
//  caCert and caKey is the signing CA
func CreateTlsCert(cn string, ou string, isServer bool, clientKey *ecdsa.PrivateKey,
	caCert *x509.Certificate, caKey *ecdsa.PrivateKey) (tlscert *tls.Certificate) {

	_, derBytes, err := CreateX509Cert(cn, ou, isServer, &clientKey.PublicKey, caCert, caKey)
	if err == nil {
		// A TLS certificate is a wrapper around x509 with private key
		tlscert = &tls.Certificate{}
		tlscert.Certificate = append(tlscert.Certificate, derBytes)
		tlscert.PrivateKey = clientKey
	}
	if err != nil {
		logrus.Errorf("CreateSignedCert. Failed creating cert: %s", err)
		return nil
	}
	return tlscert
}

// CreateX509Cert generates a x509 certificate with keys, signed by the CA, valid for 127.0.0.1
// intended for testing, not for production
//  cn is the certificate common name, usually the client ID or server hostname
//  ou the organization
//  isServer if set allow key usage of ServerAuth instead of ClientAuth
//  pubKey is the owner public key for this certificate
//  caCert and caKey is the signing CA
func CreateX509Cert(cn string, ou string, isServer bool, pubKey *ecdsa.PublicKey,
	caCert *x509.Certificate, caKey *ecdsa.PrivateKey) (cert *x509.Certificate, derBytes []byte, err error) {
	validity := time.Hour

	extkeyUsage := x509.ExtKeyUsageClientAuth
	keyUsage := x509.KeyUsageDigitalSignature
	if isServer {
		extkeyUsage = x509.ExtKeyUsageServerAuth
		keyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Country:            []string{"CA"},
			Organization:       []string{"Testing"},
			Province:           []string{"BC"},
			Locality:           []string{"wostzone"},
			CommonName:         cn,
			OrganizationalUnit: []string{ou},
			Names:              make([]pkix.AttributeTypeAndValue, 0),
		},
		NotBefore:   time.Now().Add(-10 * time.Second),
		NotAfter:    time.Now().Add(validity),
		KeyUsage:    keyUsage,
		ExtKeyUsage: []x509.ExtKeyUsage{extkeyUsage},

		BasicConstraintsValid: true,
		IsCA:                  false,
		IPAddresses:           []net.IP{net.ParseIP(ServerAddress)},
	}

	// Not for production. Ignore all but the first error. Testing would fail if this fails.
	certDerBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, pubKey, caKey)
	certPEMBuffer := new(bytes.Buffer)
	_ = pem.Encode(certPEMBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes})
	cert, _ = x509.ParseCertificate(certDerBytes)
	if err != nil {
		logrus.Panicf("CreateSignedCert. Failed creating cert: %s", err)
	}
	return cert, certDerBytes, err
}
