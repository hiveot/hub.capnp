// Package certsetup with creation of self signed certificate chain
// Credits: https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251
package certsetup

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

const keySize = 2048 // 4096
const caDurationYears = 10
const certDurationYears = 10

// Standard client and server certificate filenames
const (
	CaCertFile     = "ca.crt"
	CaKeyFile      = "ca.key"
	ServerCertFile = "hub.crt"
	ServerKeyFile  = "hub.key"
	ClientCertFile = "client.crt"
	ClientKeyFile  = "client.key"
)

// func GenCARoot() (*x509.Certificate, []byte, *rsa.PrivateKey) {
// 	if _, err := os.Stat("someFile"); err == nil {
// 		//read PEM and cert from file
// 	}
// 	var rootTemplate = x509.Certificate{
// 		SerialNumber: big.NewInt(1),
// 		Subject: pkix.Name{
// 			Country:      []string{"SE"},
// 			Organization: []string{"Company Co."},
// 			CommonName:   "Root CA",
// 		},
// 		NotBefore:             time.Now().Add(-10 * time.Second),
// 		NotAfter:              time.Now().AddDate(10, 0, 0),
// 		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
// 		BasicConstraintsValid: true,
// 		IsCA:                  true,
// 		MaxPathLen:            2,
// 		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
// 	}
// 	priv, err := rsa.GenerateKey(rand.Reader, 2048)
// 	if err != nil {
// 		panic(err)
// 	}
// 	rootCert, rootPEM := genCert(&rootTemplate, &rootTemplate, &priv.PublicKey, priv)
// 	return rootCert, rootPEM, priv
// }

// CreateWoSTCA creates WoST CA and certificate and private key for signing server certificates
// Source: https://shaneutt.com/blog/golang-ca-and-signed-cert-go/
func CreateWoSTCA() (certPEM []byte, keyPEM []byte) {
	// set up our CA certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:  []string{"WoST Zone"},
			Country:       []string{"CA"},
			Province:      []string{"BC"},
			Locality:      []string{"Project"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "Root CA",
		},
		NotBefore:             time.Now().Add(-10 * time.Second),
		NotAfter:              time.Now().AddDate(caDurationYears, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		// MaxPathLen: 2,
		// 		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create the CA private key
	privKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		logrus.Errorf("CertSetup: Unable to create private key: %s", err)
		return nil, nil
	}

	// PEM encode private key
	privKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(privKeyPEMBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privKey.PublicKey, privKey)
	if err != nil {
		logrus.Errorf("CertSetup: Unable to create CA cert: %s", err)
		return nil, nil
	}

	// pem encode certificate
	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	return certPEMBuffer.Bytes(), privKeyPEMBuffer.Bytes()
}

// CreateHubCert creates Wost hub server key and certificate
func CreateHubCert(caCertPEM []byte, caKeyPEM []byte, hostname string) (pkPEM []byte, certPEM []byte, err error) {
	// We need the CA key and certificate
	caPrivKeyBlock, _ := pem.Decode(caKeyPEM)
	caPrivKey, err := x509.ParsePKCS1PrivateKey(caPrivKeyBlock.Bytes)
	certBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	// hostname = "localhost"
	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:  []string{"WoST Zone"},
			Country:       []string{"CA"},
			Province:      []string{"BC"},
			Locality:      []string{"WoST Hub"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    hostname,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(certDurationYears, 0, 0),
		// SubjectKeyId: []byte{1, 2, 3, 4, 6},   // WTF is this???
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// If an IP address is given, then allow localhost
	ipAddr := net.ParseIP(hostname)
	if ipAddr != nil {
		logrus.Infof("CreateHubCert: hostname %s is an IP address. Setting as SAN", hostname)
		cert.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback, ipAddr}
	}

	privKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &privKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	privKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(privKeyPEMBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	return certPEMBuffer.Bytes(), privKeyPEMBuffer.Bytes(), nil
}

// CreateClientCert creates a client side certificate, signed by the CA
func CreateClientCert(caCertPEM []byte, caKeyPEM []byte, hostname string) (pkPEM []byte, certPEM []byte, err error) {
	// We need the CA key and certificate
	caPrivKeyBlock, _ := pem.Decode(caKeyPEM)
	caPrivKey, err := x509.ParsePKCS1PrivateKey(caPrivKeyBlock.Bytes)
	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return nil, nil, err
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	// hostname = "localhost"
	// set up our server certificate
	clientCert := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:  []string{"WoST"},
			Country:       []string{"CA"},
			Province:      []string{"BC"},
			Locality:      []string{"WoST Client"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    hostname,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(certDurationYears, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	clientKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	clientCertBytes, err := x509.CreateCertificate(rand.Reader, clientCert, caCert, &clientKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	clientCertPEMBuffer := new(bytes.Buffer)
	pem.Encode(clientCertPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertBytes,
	})

	clientKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(clientKeyPEMBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
	})

	return clientCertPEMBuffer.Bytes(), clientKeyPEMBuffer.Bytes(), nil

}
