package provisioning_test

import (
	"context"
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/provisioning/capnpclient"
	"github.com/hiveot/hub/pkg/provisioning/capnpserver"
	"github.com/hiveot/hub/pkg/provisioning/service"
)

// when testing using the capnp RPC
const testAddress = "/tmp/provisioning_test.socket"
const useTestCapnp = true

// provide the capability to create and verify device certificates
// this creates a test instance of the certificate service
func getCertCap() certs.ICerts {
	caCert, caKey, _ := selfsigned.CreateHubCA(1)
	certCap := selfsigned.NewSelfSignedCertsService(caCert, caKey)
	return certCap
}

func newServer(useCapnp bool) (svc provisioning.IProvisioning, closeFn func()) {
	certCap := getCertCap()
	ctx, cancelFunc := context.WithCancel(context.Background())

	provSvc := service.NewProvisioningService(ctx, certCap.CapDeviceCerts(), certCap.CapVerifyCerts())

	// optionally test with capnp RPC
	if useCapnp {
		_ = syscall.Unlink(testAddress)
		lis, _ := net.Listen("unix", testAddress)
		go capnpserver.StartProvisioningCapnpServer(ctx, lis, provSvc)
		// connect the client to the server above
		clConn, _ := net.Dial("unix", testAddress)
		cl, err := capnpclient.NewProvisioningCapnpClient(ctx, clConn)
		if err != nil {
			logrus.Fatalf("Failed starting capnp client: %s", err)
		}
		return cl, cancelFunc
	}
	return provSvc, cancelFunc
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

// Test starting the provisioning service
func TestStartStop(t *testing.T) {
	// this needs a certificate service capability
	provServer, closeFn := newServer(useTestCapnp)
	defer closeFn()
	assert.NotNil(t, provServer)
}

func TestAutomaticProvisioning(t *testing.T) {
	const device1ID = "device1"
	const secret1 = "secret1"
	device1Keys := certsclient.CreateECDSAKeys()
	ctx := context.Background()

	secrets := make([]provisioning.OOBSecret, 2)
	secrets[0] = provisioning.OOBSecret{DeviceID: device1ID, OobSecret: secret1}
	secrets[1] = provisioning.OOBSecret{DeviceID: "device2", OobSecret: "secret2"}
	provServer, closeFn := newServer(useTestCapnp)
	defer closeFn()

	capManage := provServer.CapManageProvisioning(ctx)
	defer capManage.Release()
	capProv := provServer.CapRequestProvisioning(ctx)
	defer capProv.Release()

	err := capManage.AddOOBSecrets(ctx, secrets)
	assert.NoError(t, err)

	// next, provisioning should succeed
	secret1md5 := fmt.Sprint(md5.Sum([]byte(secret1)))
	pubKeyPEM, err := certsclient.PublicKeyToPEM(&device1Keys.PublicKey)
	assert.NoError(t, err)
	status, err := capProv.SubmitProvisioningRequest(
		ctx, "device1", secret1md5, pubKeyPEM)
	require.NoError(t, err)
	assert.Equal(t, device1ID, status.DeviceID)
	assert.NotEmpty(t, status.ClientCertPEM)
	assert.NotEmpty(t, status.CaCertPEM)
	assert.False(t, status.Pending)
	assert.NotEmpty(t, status.RequestTime)

	// provisioned device should show up in the list of approved devices
	approved, err := capManage.GetApprovedRequests(ctx)
	assert.NoError(t, err)
	require.True(t, len(approved) > 0)
	assert.Equal(t, device1ID, approved[0].DeviceID)
}

func TestAutomaticProvisioningBadParameters(t *testing.T) {
	const device1ID = "device1"
	const secret1 = "secret1"
	ctx := context.Background()
	device1Keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&device1Keys.PublicKey)
	secrets := make([]provisioning.OOBSecret, 1)
	secrets[0] = provisioning.OOBSecret{DeviceID: device1ID, OobSecret: secret1}

	provServer, closeFn := newServer(useTestCapnp)
	defer closeFn()
	capProv := provServer.CapRequestProvisioning(ctx)
	defer capProv.Release()
	capManage := provServer.CapManageProvisioning(ctx)
	defer capManage.Release()

	// add a secret for testing
	err := capManage.AddOOBSecrets(context.Background(), secrets)
	assert.NoError(t, err)

	// test missing deviceID
	_, err = capProv.SubmitProvisioningRequest(
		ctx, "", "", pubKeyPEM)
	require.Error(t, err)

	// test missing public key
	_, err = capProv.SubmitProvisioningRequest(
		ctx, device1ID, "", "")
	require.Error(t, err)

	// test bad public key
	_, err = capProv.SubmitProvisioningRequest(
		ctx, device1ID, "", "badpubkey")
	require.Error(t, err)

	// test bad secret. This should return an error and pending status
	status, err := capProv.SubmitProvisioningRequest(
		ctx, device1ID, "badsecret", pubKeyPEM)
	require.NoError(t, err)
	require.True(t, status.Pending)
}

func TestManualProvisioning(t *testing.T) {
	const device1ID = "device1"

	// setup
	device1Keys := certsclient.CreateECDSAKeys()
	ctx := context.Background()
	provServer, closeFn := newServer(useTestCapnp)
	defer closeFn()
	capProv := provServer.CapRequestProvisioning(ctx)
	defer capProv.Release()
	capManage := provServer.CapManageProvisioning(ctx)
	defer capManage.Release()

	// Stage 1: request provisioning without a secret.
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&device1Keys.PublicKey)
	status, err := capProv.SubmitProvisioningRequest(
		ctx, device1ID, "", pubKeyPEM)
	// This should return a 'pending' status
	require.NoError(t, err)
	assert.Equal(t, device1ID, status.DeviceID)
	assert.Empty(t, status.ClientCertPEM)
	//assert.NotEmpty(t, status.CaCertPEM)
	assert.True(t, status.Pending)
	assert.NotEmpty(t, status.RequestTime)

	// provisioned device should be added to the list of pending devices
	pendingList, err := capManage.GetPendingRequests(ctx)
	require.True(t, len(pendingList) > 0)
	assert.Equal(t, device1ID, pendingList[0].DeviceID)
	approvedList, err := capManage.GetApprovedRequests(ctx)
	assert.NoError(t, err)
	assert.True(t, len(approvedList) == 0)

	// Stage 2: approve the request
	err = capManage.ApproveRequest(ctx, device1ID)
	assert.NoError(t, err)

	// provisioning request should now succeed
	status, err = capProv.SubmitProvisioningRequest(
		ctx, "device1", "", pubKeyPEM)
	// This should now succeed
	require.NoError(t, err)
	require.False(t, status.Pending)
	require.NotEmpty(t, status.ClientCertPEM)
	require.NotEmpty(t, status.CaCertPEM)

	// provisioned device should now show up in the list of approved devices
	approvedList, err = capManage.GetApprovedRequests(ctx)
	assert.NoError(t, err)
	require.True(t, len(approvedList) > 0)
	assert.Equal(t, device1ID, approvedList[0].DeviceID)

	pendingList, err = capManage.GetPendingRequests(ctx)
	require.True(t, len(pendingList) == 0)
}

func TestRefreshProvisioning(t *testing.T) {

	const device1ID = "device1"
	const secret1 = "secret1"
	//setup and generate a certificate
	device1Keys := certsclient.CreateECDSAKeys()
	pubKeyPEM, _ := certsclient.PublicKeyToPEM(&device1Keys.PublicKey)
	secrets := make([]provisioning.OOBSecret, 1)
	secrets[0] = provisioning.OOBSecret{DeviceID: device1ID, OobSecret: secret1}
	ctx := context.Background()

	// request provisioning with a valid secret.
	provServer, closeFn := newServer(useTestCapnp)
	defer closeFn()
	capProv := provServer.CapRequestProvisioning(ctx)
	defer capProv.Release()
	capRefresh := provServer.CapRefreshProvisioning(ctx)
	defer capRefresh.Release()
	capManage := provServer.CapManageProvisioning(ctx)
	defer capManage.Release()

	// obtain a certificate
	err := capManage.AddOOBSecrets(context.Background(), secrets)
	assert.NoError(t, err)
	secret1md5 := fmt.Sprint(md5.Sum([]byte(secret1)))
	status, err := capProv.SubmitProvisioningRequest(
		context.Background(), device1ID, secret1md5, pubKeyPEM)
	require.NoError(t, err)
	assert.NotEmpty(t, status.ClientCertPEM)

	// refresh
	status2, err := capRefresh.RefreshDeviceCert(
		context.Background(), status.ClientCertPEM)
	// This should succeed
	require.NoError(t, err)
	require.False(t, status2.Pending)
	require.NotEmpty(t, status2.ClientCertPEM)
	require.NotEmpty(t, status2.CaCertPEM)

	// refresh with bad certificate should fail
	_, err = capRefresh.RefreshDeviceCert(
		context.Background(), "bad certificate")
	require.Error(t, err)
}
