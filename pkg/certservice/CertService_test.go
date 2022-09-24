package main_test

import (
	"crypto/x509"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/certservice/adapter"
	"github.com/hiveot/hub/pkg/certservice/client"
	"github.com/hiveot/hub/pkg/certservice/selfsigned"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//var tempFolder string
//var certFolder string

const TempCertDurationDays = 1
const useCapnp = true
const testAddress = "/tmp/certservice_test.socket"

// removeCerts easy cleanup for existing device certificate
//func removeServerCerts() {
//	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(certFolder, "*.pem")).Output()
//}

// Factory for creating service instance. Currently the only implementation is selfsigned.
func NewService() client.ICertService {
	// use selfsigned to create a new CA for these tests
	caCert, caKey, _ := selfsigned.CreateHubCA(1)
	service := selfsigned.NewSelfSignedService(caCert, caKey)
	// when using capnp, return a client instance instead the service
	if useCapnp {
		// remove stale handle
		_ = syscall.Unlink(testAddress)
		lis, _ := net.Listen("unix", testAddress)
		go adapter.StartCertServiceCapnpAdapter(lis, service)
		capClient, _ := client.NewCertServiceCapnpClient(testAddress, true)
		return capClient
	}
	return service
}

// TestMain clears the certs folder for clean testing
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	//tempFolder := path.Join(os.TempDir(), "hiveot-certs-test")
	// clean start
	//os.RemoveAll(tempFolder)
	//certFolder = path.Join(tempFolder, "certs")
	//_ = os.MkdirAll(certFolder, 0700)
	logging.SetLogging("info", "")
	//removeServerCerts()

	res := m.Run()
	if res == 0 {
		//os.RemoveAll(tempFolder)
	}
	os.Exit(res)
}

func TestCreateService(t *testing.T) {
	svc := NewService()
	require.NotNil(t, svc)
}

func TestCreateDeviceCert(t *testing.T) {
	deviceID := "device1"

	service := NewService()
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)

	deviceCertPEM, caCertPEM, err := service.CreateDeviceCert(deviceID, pubKeyPEM, 1)
	require.NoError(t, err)

	deviceCert, err := certsclient.X509CertFromPEM(deviceCertPEM)
	require.NoError(t, err)
	require.NotNil(t, deviceCert)
	caCert2, err := certsclient.X509CertFromPEM(caCertPEM)
	require.NoError(t, err)
	require.NotNil(t, caCert2)

	// verify certificate against CA
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert2)
	opts := x509.VerifyOptions{
		Roots:     caCertPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	_, err = deviceCert.Verify(opts)
	assert.NoError(t, err)
}

// test device cert with bad parameters
func TestDeviceCertBadParms(t *testing.T) {
	deviceID := "device1"

	// test creating hub certificate
	service := NewService()
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)

	// missing device ID
	certPEM, _, err := service.CreateDeviceCert("", pubKeyPEM, 0)
	require.Error(t, err)
	assert.Empty(t, certPEM)

	// missing public key
	certPEM, _, err = service.CreateDeviceCert(deviceID, "", 1)
	require.Error(t, err)
	assert.Empty(t, certPEM)

}

func TestCreateServiceCert(t *testing.T) {
	// test creating hub certificate
	const serviceID = "testService"
	names := []string{"127.0.0.1", "localhost"}

	svc := NewService()
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)
	serviceCertPEM, caCertPEM, err := svc.CreateServiceCert(serviceID, pubKeyPEM, names, 0)
	require.NoError(t, err)
	serviceCert, err := certsclient.X509CertFromPEM(serviceCertPEM)
	require.NoError(t, err)
	caCert2, err := certsclient.X509CertFromPEM(caCertPEM)
	require.NoError(t, err)

	// verify service certificate against CA
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert2)
	opts := x509.VerifyOptions{
		Roots:     caCertPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	_, err = serviceCert.Verify(opts)
	assert.NoError(t, err)
}

// test with bad parameters
func TestServiceCertBadParms(t *testing.T) {
	const serviceID = "testService"
	hostnames := []string{"127.0.0.1"}

	caCert, caKey, _ := selfsigned.CreateHubCA(1)
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)

	// Bad CA certificate
	badCa := x509.Certificate{}
	assert.Panics(t, func() {
		selfsigned.NewSelfSignedService(&badCa, caKey)
	})

	// missing CA private key
	assert.Panics(t, func() {
		selfsigned.NewSelfSignedService(caCert, nil)
	})

	// missing service ID
	service := selfsigned.NewSelfSignedService(caCert, caKey)
	serviceCertPEM, _, err := service.CreateServiceCert(
		"", pubKeyPEM, hostnames, 1)
	require.Error(t, err)
	require.Empty(t, serviceCertPEM)

	// missing public key
	serviceCertPEM, _, err = service.CreateServiceCert(
		serviceID, "", hostnames, 1)
	require.Error(t, err)
	require.Empty(t, serviceCertPEM)

}

func TestCreateUserCert(t *testing.T) {
	userID := "bob"
	// test creating hub certificate
	service := NewService()
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)

	userCertPEM, caCertPEM, err := service.CreateUserCert(userID, pubKeyPEM, 0)
	require.NoError(t, err)

	userCert, err := certsclient.X509CertFromPEM(userCertPEM)
	require.NoError(t, err)
	require.NotNil(t, userCert)
	caCert2, err := certsclient.X509CertFromPEM(caCertPEM)
	require.NoError(t, err)
	require.NotNil(t, caCert2)

	// verify client certificate against CA
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert2)
	opts := x509.VerifyOptions{
		Roots:     caCertPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	_, err = userCert.Verify(opts)
	assert.NoError(t, err)
}
