package gateway_test

import (
	"context"
	"crypto/ecdsa"
	"net"
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/certs"
	capnpclient2 "github.com/hiveot/hub/pkg/certs/capnpclient"
	capnpserverCerts "github.com/hiveot/hub/pkg/certs/capnpserver"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
)

const testSocketDir = "/tmp/test-gateway"
const testClientID = "client1"
const testUseCapnp = true

var testSocket = path.Join(testSocketDir, gateway.ServiceName+".socket")
var testCACert string
var testClientKeys *ecdsa.PrivateKey
var testClientCertPem string
var testClientPubKeyPem string

// this creates a test instance of the certificate service
func startCertService(ctx context.Context) certs.ICerts {
	var err error
	caCert, caKey, _ := selfsigned.CreateHubCA(1)
	certCap := selfsigned.NewSelfSignedCertsService(caCert, caKey)
	srvListener := listener.CreateServiceListener(testSocketDir, certs.ServiceName)

	logrus.Infof("CertServiceCapnpServer starting on: %s", srvListener.Addr())
	svc := selfsigned.NewSelfSignedCertsService(caCert, caKey)
	go capnpserverCerts.StartCertsCapnpServer(ctx, srvListener, svc)

	cuc := certCap.CapUserCerts()

	testclientKeys := certsclient.CreateECDSAKeys()
	testClientPubKeyPem, _ = certsclient.PublicKeyToPEM(&testclientKeys.PublicKey)
	testClientCertPem, testCACert, err = cuc.CreateUserCert(ctx, testClientID, testClientPubKeyPem, 1)
	if err != nil {
		panic(err)
	}
	cuc.Release()

	return certCap
}

func startService(useCapnp bool) (gateway.IGatewayService, func() error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	_ = os.RemoveAll(testSocketDir)
	err := os.MkdirAll(testSocketDir, 0700)

	// start with the authn service capabilities
	certSvc := startCertService(ctx)
	_ = certSvc

	//
	svc := service.NewGatewayService(testSocketDir)
	err = svc.Start(ctx)
	if err != nil {
		logrus.Panicf("Failed to start with socket dir %s", testSocketDir)
	}
	// optionally test with capnp RPC
	if useCapnp {
		// start the capnpserver for the service
		_ = syscall.Unlink(testSocket)
		srvListener, err := net.Listen("unix", testSocket)
		if err != nil {
			logrus.Panic("Unable to create a listener, can't run test")
		}
		go capnpserver.StartGatewayServiceCapnpServer(ctx, srvListener, svc, testSocketDir)

		// connect the client to the server above
		clConn, _ := net.Dial("unix", testSocket)
		capClient, err := capnpclient.NewGatewayServiceCapnpClient(ctx, clConn)
		return capClient, func() error {
			cancelFunc()
			//_ = capClient.Stop(ctx)
			err = svc.Stop(ctx)
			//_ = certSvc.Stop(ctx)
			return err
		}
	}
	return svc, func() error {
		cancelFunc()
		err = svc.Stop(ctx)
		//_ = certSvc.Stop(ctx)
		return err
	}
}
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	//ctx, stopFn := context.WithCancel(context.Background())
	//err := os.MkdirAll(testSocketDir, 0700)
	//require.NoError(t, err)
	//svc := service.NewGatewayService(testSocketDir)
	//err = svc.Start(ctx)
	//assert.NoError(t, err)
	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	pong, err := svc.Ping(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "pong", pong)

	err = svc.Stop(ctx)
	assert.NoError(t, err)
	stopFn()
}

//
//func TestExtractCapConfig(t *testing.T) {
//	const service1 = "service1"
//	ctx, stopFn := context.WithCancel(context.Background())
//	err := os.MkdirAll(testSocketDir, 0700)
//	require.NoError(t, err)
//
//	svc := service.NewGatewayService(testSocketDir)
//	err = svc.Start(ctx)
//	assert.NoError(t, err)
//	// extract capabilities
//	caps := svc.ExtractCapabilitiesFromConfig(service1, []byte(config1))
//	assert.Equal(t, 1, len(caps))
//	assert.Equal(t, 3, len(caps[0].ClientType))
//	err = svc.Stop(ctx)
//	assert.NoError(t, err)
//	stopFn()
//}

func TestGetInfo(t *testing.T) {

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	// use gateway service itself to get its capability
	caps, err := svc.ListCapabilities(ctx, hubapi.ClientTypeService)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(caps))

	err = stopFn()
	assert.NoError(t, err)
}

func TestGetCapability(t *testing.T) {
	const clientType = hubapi.ClientTypeService
	ctx := context.Background()

	svc, stopFn := startService(testUseCapnp)

	// list capabilites, get the service capability and invoke 'Ping'
	capList, err := svc.ListCapabilities(ctx, clientType)
	assert.NoError(t, err)
	require.Equal(t, 4, len(capList))
	//assert.Equal(t, "capVerifyCerts", capList[0].CapabilityName)

	// the connection method determines the client type. In this case service.
	capability, err := svc.GetCapability(ctx, testClientID, clientType, "capVerifyCerts", nil)
	require.NoError(t, err)
	assert.NotNil(t, capability)

	// cast the capability to that of the gateway.
	// Just using the gateway service itself for testing
	verifyCapabilityCapnp := hubapi.CapVerifyCerts(capability)
	verifyCapability := capnpclient2.NewVerifyCertsCapnpClient(verifyCapabilityCapnp)
	err = verifyCapability.VerifyCert(ctx, testClientID, testClientCertPem)
	assert.NoError(t, err)

	// get capability that doesn't exist
	capability, err = svc.GetCapability(ctx, testClientID, clientType, "notacapability", nil)
	assert.Error(t, err)

	err = stopFn()
	assert.NoError(t, err)
}

func TestLogin(t *testing.T) {
	//TODO
}
