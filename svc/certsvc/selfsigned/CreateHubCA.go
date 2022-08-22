package selfsigned

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/wost-go/pkg/certsclient"
)

const caDefaultValidityDuration = time.Hour * 24 * 364 * 20 // 20 years
const CertOrgName = "WoST"
const CertOrgLocality = "WoST zone"

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
	// firefox complains if serial is the same as that of the CA. So generate a unique one based on timestamp.
	serial := time.Now().Unix() - 1 // prevent duplicate timestamp with server cert
	rootTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{
			Country:      []string{"CA"},
			Province:     []string{"BC"},
			Locality:     []string{CertOrgLocality},
			Organization: []string{CertOrgName},
			CommonName:   "Hub CA",
		},
		NotBefore: time.Now().Add(-10 * time.Second),
		NotAfter:  time.Now().Add(validity),
		// CA cert can be used to sign certificate and revocation lists
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment,

		// firefox seems to consider a CA invalid if extended key usage is combined with regular (critical) key usage???
		// certificate.Verify however fails if ext key usage is just the OCSPSigning.
		//ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageOCSPSigning},
		//ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageOCSPSigning},
		// https://github.com/hashicorp/vault/issues/846 suggests no ext key usage for CA's
		ExtKeyUsage: []x509.ExtKeyUsage{},

		// This hub cert is the only CA. Not using intermediate CAs
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	// Create the CA private key
	privKey := certsclient.CreateECDSAKeys()

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
