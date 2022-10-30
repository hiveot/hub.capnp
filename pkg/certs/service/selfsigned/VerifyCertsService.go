package selfsigned

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/certsclient"
)

// VerifyCertsService creates certificates for use by services, devices and admin users.
// This implements the IVerifyCerts interface
type VerifyCertsService struct {
	caCert     *x509.Certificate
	caCertPEM  string
	caCertPool *x509.CertPool
}

// VerifyCert verifies whether the given certificate is a valid client certificate
func (srv *VerifyCertsService) VerifyCert(
	_ context.Context, clientID string, certPEM string) error {

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

// Release the provided capability and release resources
func (srv *VerifyCertsService) Release() {
	// nothing to do here
}

// NewVerifyCertsService returns a new instance of the certificate verification service
//  caCert is the CA certificate to verify against
func NewVerifyCertsService(caCert *x509.Certificate) *VerifyCertsService {
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	service := &VerifyCertsService{
		caCert:     caCert,
		caCertPEM:  certsclient.X509CertToPEM(caCert),
		caCertPool: caCertPool,
	}
	if caCert == nil || caCert.PublicKey == nil {
		logrus.Panic("Missing CA certificate or key")
	}

	return service
}
