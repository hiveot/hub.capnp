package certs_test

import (
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/certs"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
)

// ! This uses the TestMain of Keys_test.go !

// removeTestCerts easy cleanup for existing keys and certificate
func removeTestCerts() {
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(certFolder, "*.pem")).Output()
}

func TestSaveLoadX509Cert(t *testing.T) {
	// hostnames := []string{"localhost"}
	caPemFile := path.Join(certFolder, "caCert.pem")

	testCerts := testenv.CreateCertBundle()
	removeTestCerts()

	// save the test x509 cert
	err := certs.SaveX509CertToPEM(testCerts.CaCert, caPemFile)
	assert.NoError(t, err)

	cert, err := certs.LoadX509CertFromPEM(caPemFile)
	assert.NoError(t, err)
	assert.NotNil(t, cert)
}

func TestSaveLoadTLSCert(t *testing.T) {
	// hostnames := []string{"localhost"}
	certFile := path.Join(certFolder, "tlscert.pem")
	keyFile := path.Join(certFolder, "tlskey.pem")

	testCerts := testenv.CreateCertBundle()
	removeTestCerts()

	// save the test x509 part of the TLS cert
	err := certs.SaveTLSCertToPEM(testCerts.DeviceCert, certFile, keyFile)
	assert.NoError(t, err)

	// load back the x509 part of the TLS cert
	cert, err := certs.LoadTLSCertFromPEM(certFile, keyFile)
	assert.NoError(t, err)
	assert.NotNil(t, cert)
}

func TestSaveLoadCertNoFile(t *testing.T) {
	certFile := "/root/notavalidcert.pem"
	keyFile := "/root/notavalidkey.pem"
	testCerts := testenv.CreateCertBundle()
	// save the test x509 cert
	err := certs.SaveX509CertToPEM(testCerts.CaCert, certFile)
	assert.Error(t, err)

	_, err = certs.LoadX509CertFromPEM(certFile)
	assert.Error(t, err)

	// save the test x509 part of the TLS cert
	err = certs.SaveTLSCertToPEM(testCerts.DeviceCert, certFile, keyFile)
	assert.Error(t, err)

	// load back the x509 part of the TLS cert
	_, err = certs.LoadTLSCertFromPEM(certFile, keyFile)
	assert.Error(t, err)

}
