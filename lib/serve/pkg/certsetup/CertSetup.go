// Package certsetup with server side creation of self signed certificate chain using ECDSA
// Credits: https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251 keys
package certsetup

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/certs"
	"github.com/wostzone/hub/lib/client/pkg/config"
)

// // Standard WoST client and server key/certificate filenames. All stored in PEM format.
// const (
// 	CaCertFile     = "caCert.pem" // CA that signed the server and client certificates
// 	CaKeyFile      = "caKey.pem"
// 	HubCertFile    = "hubCert.pem"
// 	HubKeyFile     = "hubKey.pem"
// 	PluginCertFile = "pluginCert.pem"
// 	PluginKeyFile  = "pluginKey.pem"
// 	// AdminCertFile = "adminCert.pem"
// 	// AdminKeyFile  = "adminKey.pem"
// )

// Organization Unit for client authorization are stored in the client certificate OU field
const (
	// Default OU with no API access permissions
	OUNone = ""

	// OUClient lets a client connect to the message bus
	OUClient = "client"

	// OUIoTDevice indicates the client is a IoT device that can connect to the message bus
	// perform discovery and request provisioning.
	// Provision API permissions: GetDirectory, ProvisionRequest, GetStatus
	OUIoTDevice = "iotdevice"

	//OUAdmin lets a client approve thing provisioning (postOOB), add and remove users
	// Provision API permissions: GetDirectory, ProvisionRequest, GetStatus, PostOOB
	OUAdmin = "admin"

	// OUPlugin marks a client as a plugin.
	// By default, plugins have full permission to all APIs
	// Provision API permissions: Any
	OUPlugin = "plugin"

	// OUService marks a certificate as that of a Hub service.
	// By default, services have full permission to all APIs
	// Provision API permissions: Any
	OUService = "service"
)

// Certificate organization name
const CertOrgName = "WoST"
const CertOrgLocality = "WoST zone"

// DefaultPluginClientID in the certificate is standard for all plugins
const DefaultPluginClientID = "plugin"

// const keySize = 2048 // 4096
const caDefaultValidityDuration = time.Hour * 24 * 364 * 20 // 20 years
// const caTemporaryValidityDuration = time.Hour * 24 * 3      // 3 days

const DefaultCertDurationDays = 365
const TempCertDurationDays = 1

// CreateCertificateBundle is a convenience function to create the Hub CA, server and (plugin) client
// certificates into the given folder.
//  * The CA certificate will only be created if missing
//  * The plugin keys and certificate will always be recreated
//  * The service keys and certificate will always be recreated
//
//  names contain the list of hostname and ip addresses the hub can be reached at. Used in hub cert.
//  certFolder where to create the certificates
func CreateCertificateBundle(names []string, certFolder string) error {
	var err error
	forcePluginCert := true // best to always created these certs
	forceHubCert := true
	var caCert *x509.Certificate
	var caKeys *ecdsa.PrivateKey

	// create the CA only if needed
	// TODO: How to handle CA expiry?
	// TODO: Handle CA revocation
	caCert, _ = certs.LoadX509CertFromPEM(path.Join(certFolder, config.DefaultCaCertFile))
	caKeys, _ = certs.LoadKeysFromPEM(path.Join(certFolder, config.DefaultCaKeyFile))
	if caCert == nil || caKeys == nil {
		logrus.Warningf("CreateCertificateBundle Generating a CA certificate in %s as none was found. Names: %s", certFolder, names)
		caCert, caKeys = CreateHubCA()
		err = certs.SaveKeysToPEM(caKeys, path.Join(certFolder, config.DefaultCaKeyFile))
		if err != nil {
			logrus.Errorf("CreateCertificateBundle CA failed writing. Unable to continue: %s", err)
			return err
		}
		err = certs.SaveX509CertToPEM(caCert, path.Join(certFolder, config.DefaultCaCertFile))
		if err != nil {
			return err
		}
	}

	// create the Hub server cert
	serverCertPath := path.Join(certFolder, config.DefaultServerCertFile)
	serverKeyPath := path.Join(certFolder, config.DefaultServerKeyFile)
	serverCert, _ := certs.LoadTLSCertFromPEM(serverCertPath, serverKeyPath)
	if serverCert == nil || forceHubCert {
		logrus.Infof("CreateCertificateBundle Refreshing Hub server certificate in %s", certFolder)
		serverCert, err = CreateHubServerCert(names, caCert, caKeys)
		if err != nil {
			logrus.Errorf("CreateCertificateBundle server failed: %s", err)
			return err
		}
		certs.SaveTLSCertToPEM(serverCert, serverCertPath, serverKeyPath)
	}

	// create the Plugin (client) certificate
	pluginCertPath := path.Join(certFolder, config.DefaultPluginCertFile)
	pluginKeyPath := path.Join(certFolder, config.DefaultPluginKeyFile)
	pluginTlsCert, _ := certs.LoadTLSCertFromPEM(pluginCertPath, pluginKeyPath)
	if pluginTlsCert == nil || forcePluginCert {
		logrus.Infof("CreateCertificateBundle Refreshing plugin server certificate in %s", certFolder)

		// The plugin client cert uses the fixed common name 'plugin'
		privKey := certs.CreateECDSAKeys()
		pluginCert, err := CreateHubClientCert(DefaultPluginClientID, OUPlugin,
			&privKey.PublicKey, caCert, caKeys, time.Now(), DefaultCertDurationDays)
		if err != nil {
			logrus.Fatalf("CreateCertificateBundle client failed: %s", err)
		}
		certs.SaveX509CertToPEM(pluginCert, pluginCertPath)
		certs.SaveKeysToPEM(privKey, pluginKeyPath)
	}
	return nil
}

// CreateHubCA creates WoST Hub Root CA certificate and private key for signing server certificates
// Source: https://shaneutt.com/blog/golang-ca-and-signed-cert-go/
// This creates a CA certificate used for signing client and server certificates.
// CA is valid for 'caDurationYears'
//
//  temporary set to generate a temporary CA for one-off signing
func CreateHubCA() (cert *x509.Certificate, key *ecdsa.PrivateKey) {
	validity := caDefaultValidityDuration

	// set up our CA certificate
	// see also: https://superuser.com/questions/738612/openssl-ca-keyusage-extension
	rootTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Country:      []string{"CA"},
			Organization: []string{CertOrgName},
			Province:     []string{"BC"},
			Locality:     []string{CertOrgLocality},
			CommonName:   "Hub CA",
		},
		NotBefore: time.Now().Add(-10 * time.Second),
		NotAfter:  time.Now().Add(validity),
		// CA cert can be used to sign certificate and revocation lists
		KeyUsage:    x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign | x509.KeyUsageDataEncipherment|x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageOCSPSigning},

		// This hub cert is the only CA. Not using intermediate CAs
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	// Create the CA private key
	privKey := certs.CreateECDSAKeys()

	// create the CA
	caCertDer, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &privKey.PublicKey, privKey)
	if err != nil {
		// normally this never happens
		logrus.Errorf("CertSetup.CreateHubCA: Unable to create WoST Hub CA cert: %s", err)
		return nil, nil
	}
	caCert, _ := x509.ParseCertificate(caCertDer)
	return caCert, privKey
}

// CreateHubClientCert creates a hub client certificate for mutual authentication from client's public key
// The client role is intended to for role based authorization. It is stored in the
// certificate OrganizationalUnit. See OUxxx
//
// This generates a TLS client certificate with keys
//  clientID used as the CommonName, eg pluginID or deviceID
//  ou of the client role, eg OUNone, OUClient, OUPlugin
//  ownerPubKey the public key of the certificate holder
//  caCert CA's certificate for signing
//  caPrivKey CA's ECDSA key for signing
//  start time the certificate is first valid. Intended for testing. Use time.now()
//  durationDays nr of days the certificate will be valid
// Returns the signed TLS certificate or error
func CreateHubClientCert(clientID string, ou string,
	ownerPubKey *ecdsa.PublicKey, caCert *x509.Certificate, caPrivKey *ecdsa.PrivateKey,
	start time.Time, durationDays int) (clientCert *x509.Certificate, err error) {

	if caCert == nil || caPrivKey == nil {
		err := fmt.Errorf("CreateHubClientCert: missing CA cert or key")
		logrus.Error(err)
		return nil, err
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:       []string{CertOrgName},
			Locality:           []string{CertOrgLocality},
			OrganizationalUnit: []string{ou},
			Names:              make([]pkix.AttributeTypeAndValue, 0),
			CommonName:         clientID,
		},
		NotBefore: start,
		NotAfter:  start.AddDate(0, 0, durationDays),

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth,x509.ExtKeyUsageServerAuth},

		IsCA:                  false,
		BasicConstraintsValid: true,
	}
	// clientKey := certs.CreateECDSAKeys()
	certDer, err := x509.CreateCertificate(rand.Reader, template, caCert, ownerPubKey, caPrivKey)
	if err != nil {
		logrus.Errorf("CertSetup.CreateHubClientCert: Unable to create WoST Hub client cert: %s", err)
		return nil, err
	}
	newCert, err := x509.ParseCertificate(certDer)

	// // combined them into a TLS certificate
	// tlscert := &tls.Certificate{}
	// tlscert.Certificate = append(tlscert.Certificate, certDer)
	// tlscert.PrivateKey = clientKey

	return newCert, err
}

// CreateHubServerCert creates a new Hub service certificate and private key
// The certificate is valid for the given names either local domain name and IP addresses.
// The server must have a fixed IP.
//  names contains one or more domain names and/or IP addresses the Hub can be reached on, to add to the certificate
//  caCert is the CA to sign the server certificate
//  caPrivKey is the CA private key to sign the server certificate
// returns the signed Server TLS certificate
func CreateHubServerCert(names []string, caCert *x509.Certificate, caPrivKey *ecdsa.PrivateKey) (cert *tls.Certificate, err error) {
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

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:       []string{CertOrgName},
			Country:            []string{"CA"},
			Province:           []string{"BC"},
			Locality:           []string{CertOrgLocality},
			OrganizationalUnit: []string{OUAdmin},
			CommonName:         "Hub Server",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(0, 0, DefaultCertDurationDays),

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign | x509.KeyUsageDataEncipherment|x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageOCSPSigning},
		// ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
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
	// Create the server private key
	certKey := certs.CreateECDSAKeys()
	// and the certificate itself
	certDer, err := x509.CreateCertificate(rand.Reader, template, caCert,
		&certKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}
	// combined them into a TLS certificate
	tlscert := &tls.Certificate{}
	tlscert.Certificate = append(tlscert.Certificate, certDer)
	tlscert.PrivateKey = certKey

	return tlscert, nil
}
