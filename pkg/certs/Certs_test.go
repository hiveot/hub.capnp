package certs_test

import (
	"context"
	"crypto/x509"
	"net"
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/certs/capnpclient"
	"github.com/hiveot/hub/pkg/certs/capnpserver"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const useCapnp = true

var testFolder = path.Join(os.TempDir(), "test-certs")
var testSocket = path.Join(testFolder, "certs.socket")

// removeCerts easy cleanup for existing device certificate
//func removeServerCerts() {
//	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(certFolder, "*.pem")).Output()
//}

// Factory for creating service instance. Currently the only implementation is selfsigned.
func NewService() (svc certs.ICerts, closeFunc func()) {
	// use selfsigned to create a new CA for these tests
	//ctx, cancelFunc := context.WithCancel(context.Background())
	caCert, caKey, _ := selfsigned.CreateHubCA(1)
	certSvc := selfsigned.NewSelfSignedCertsService(caCert, caKey)
	ctx := context.Background()
	// when using capnp, return a client instance instead the svc
	if useCapnp {
		// remove stale handle
		_ = syscall.Unlink(testSocket)
		srvListener, _ := net.Listen("unix", testSocket)
		go capnpserver.StartCertsCapnpServer(ctx, srvListener, certSvc)
		// connect the client to the server above
		clConn, _ := net.Dial("unix", testSocket)
		capClient, _ := capnpclient.NewCertServiceCapnpClient(clConn)
		return capClient, func() {
			capClient.Release()
			certSvc.Stop()
		}
	}
	return certSvc, func() {
		certSvc.Stop()
	}
}

// TestMain clears the certs folder for clean testing
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	// clean start
	os.RemoveAll(testFolder)
	_ = os.MkdirAll(testFolder, 0700)
	logging.SetLogging("info", "")
	//removeServerCerts()

	res := m.Run()
	if res == 0 {
		//os.RemoveAll(tempFolder)
	}
	os.Exit(res)
}

func TestCreateService(t *testing.T) {
	svc, cancelFunc := NewService()
	defer cancelFunc()
	require.NotNil(t, svc)
}

func TestCreateDeviceCert(t *testing.T) {
	deviceID := "device1"
	ctx := context.Background()

	svc, cancelFunc := NewService()
	defer cancelFunc()
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)

	deviceCertsSvc := svc.CapDeviceCerts(ctx)
	defer deviceCertsSvc.Release()
	deviceCertPEM, caCertPEM, err := deviceCertsSvc.CreateDeviceCert(
		ctx, deviceID, pubKeyPEM, 1)
	require.NoError(t, err)

	deviceCert, err := certsclient.X509CertFromPEM(deviceCertPEM)
	require.NoError(t, err)
	require.NotNil(t, deviceCert)
	caCert2, err := certsclient.X509CertFromPEM(caCertPEM)
	require.NoError(t, err)
	require.NotNil(t, caCert2)

	// verify certificate
	verifyCertsSvc := svc.CapVerifyCerts(ctx)
	defer verifyCertsSvc.Release()
	err = verifyCertsSvc.VerifyCert(ctx, deviceID, deviceCertPEM)
	assert.NoError(t, err)
	err = verifyCertsSvc.VerifyCert(ctx, "notanid", deviceCertPEM)
	assert.Error(t, err)

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
	ctx := context.Background()

	// test creating hub certificate
	svc, cancelFunc := NewService()
	defer cancelFunc()
	deviceCertsSvc := svc.CapDeviceCerts(ctx)
	defer deviceCertsSvc.Release()

	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)

	// missing device ID
	certPEM, _, err := deviceCertsSvc.CreateDeviceCert(ctx, "", pubKeyPEM, 0)
	require.Error(t, err)
	assert.Empty(t, certPEM)

	// missing public key
	certPEM, _, err = deviceCertsSvc.CreateDeviceCert(ctx, deviceID, "", 1)
	require.Error(t, err)
	assert.Empty(t, certPEM)

}

func TestCreateServiceCert(t *testing.T) {
	// test creating hub certificate
	const serviceID = "testService"
	names := []string{"127.0.0.1", "localhost"}
	ctx := context.Background()

	svc, cancelFunc := NewService()
	defer cancelFunc()
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)
	serviceCertsSvc := svc.CapServiceCerts(ctx)
	defer serviceCertsSvc.Release()

	serviceCertPEM, caCertPEM, err := serviceCertsSvc.CreateServiceCert(
		ctx, serviceID, pubKeyPEM, names, 0)
	require.NoError(t, err)
	serviceCert, err := certsclient.X509CertFromPEM(serviceCertPEM)
	require.NoError(t, err)
	caCert2, err := certsclient.X509CertFromPEM(caCertPEM)
	require.NoError(t, err)

	// verify service certificate against CA
	verifyCertsSvc := svc.CapVerifyCerts(ctx)
	defer verifyCertsSvc.Release()

	err = verifyCertsSvc.VerifyCert(ctx, serviceID, serviceCertPEM)
	assert.NoError(t, err)

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
	ctx := context.Background()

	caCert, caKey, _ := selfsigned.CreateHubCA(1)
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)

	// Bad CA certificate
	badCa := x509.Certificate{}
	assert.Panics(t, func() {
		selfsigned.NewSelfSignedCertsService(&badCa, caKey)
	})

	// missing CA private key
	assert.Panics(t, func() {
		selfsigned.NewSelfSignedCertsService(caCert, nil)
	})

	// missing service ID
	svc := selfsigned.NewSelfSignedCertsService(caCert, caKey)
	capServiceCerts := svc.CapServiceCerts(ctx)
	defer capServiceCerts.Release()
	serviceCertPEM, _, err := capServiceCerts.CreateServiceCert(
		ctx, "", pubKeyPEM, hostnames, 1)

	require.Error(t, err)
	require.Empty(t, serviceCertPEM)

	// missing public key
	serviceCertPEM, _, err = capServiceCerts.CreateServiceCert(
		ctx, serviceID, "", hostnames, 1)
	require.Error(t, err)
	require.Empty(t, serviceCertPEM)

}

func TestCreateUserCert(t *testing.T) {
	ctx := context.Background()
	userID := "bob"
	// test creating hub certificate
	svc, cancelFunc := NewService()
	defer cancelFunc()
	keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&keys.PublicKey)

	capUserCert := svc.CapUserCerts(ctx)
	defer capUserCert.Release()
	userCertPEM, caCertPEM, err := svc.CapUserCerts(ctx).CreateUserCert(
		ctx, userID, pubKeyPEM, 0)
	require.NoError(t, err)

	userCert, err := certsclient.X509CertFromPEM(userCertPEM)
	require.NoError(t, err)
	require.NotNil(t, userCert)
	caCert2, err := certsclient.X509CertFromPEM(caCertPEM)
	require.NoError(t, err)
	require.NotNil(t, caCert2)

	// verify service certificate against CA
	capVerifyCerts := svc.CapVerifyCerts(ctx)
	defer capVerifyCerts.Release()
	err = capVerifyCerts.VerifyCert(ctx, userID, userCertPEM)
	assert.NoError(t, err)

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
