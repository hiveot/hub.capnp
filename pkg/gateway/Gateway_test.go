package gateway_test

import (
	"context"
	"crypto/ecdsa"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub.go/pkg/testenv"
	"github.com/hiveot/hub/internal/captest"
	"github.com/hiveot/hub/internal/dummy"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/certs"
	capnpserverCerts "github.com/hiveot/hub/pkg/certs/capnpserver"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
	capnpserver2 "github.com/hiveot/hub/pkg/resolver/capnpserver"
)

const testSocketDir = "/tmp/test-gateway"
const testClientID = "client1"
const testUseCapnp = true

var certsSocketPath = path.Join(testSocketDir, certs.ServiceName+".socket")
var testSocketPath = path.Join(testSocketDir, gateway.ServiceName+".socket")
var testAddress = "127.0.0.1:0"
var testCACert string
var testClientKeys *ecdsa.PrivateKey
var testClientCertPem string
var testClientPubKeyPem string

// CA, server and plugin test certificate
var testCerts testenv.TestCerts

// this creates a test instance of the certificate service
func startCertService(ctx context.Context) certs.ICerts {
	var err error
	caCert, caKey, _ := selfsigned.CreateHubCA(1)
	certCap := selfsigned.NewSelfSignedCertsService(caCert, caKey)

	srvListener, _ := net.Listen("unix", certsSocketPath)

	logrus.Infof("CertServiceCapnpServer starting on: %s", srvListener.Addr())
	svc := selfsigned.NewSelfSignedCertsService(caCert, caKey)
	go capnpserverCerts.StartCertsCapnpServer(srvListener, svc)

	cuc := certCap.CapUserCerts(ctx)

	testclientKeys := certsclient.CreateECDSAKeys()
	testClientPubKeyPem, _ = certsclient.PublicKeyToPEM(&testclientKeys.PublicKey)
	testClientCertPem, testCACert, err = cuc.CreateUserCert(ctx, testClientID, testClientPubKeyPem, 1)
	if err != nil {
		panic(err)
	}
	cuc.Release()

	return certCap
}

func startService(useCapnp bool) (gateway.IGatewaySession, func() error) {
	_ = os.RemoveAll(testSocketDir)
	_ = os.MkdirAll(testSocketDir, 0700)

	// test needs a resolver with authn service
	resolverStop, err := capnpserver2.StartResolver(testSocketPath)
	if err != nil {
		panic("unable to start the resolver")
	}
	//authnService := service3.NewAuthnService(config.AuthnConfig{
	//	PasswordFile: "",
	//})
	authnService := dummy.NewDummyAuthnService()
	svc := service.NewGatewayService(testSocketPath, authnService)
	// optionally test with capnp RPC over TLS
	if useCapnp {
		// start the capnpserver for the service
		// TLS only works on sockets as address must match certificate name
		//_ = syscall.Unlink(testSocket)
		srvListener, err2 := net.Listen("tcp", testAddress)
		if err2 != nil {
			logrus.Panicf("Unable to create a listener, can't run test: %s", err2)
		}

		// wrap the connection in TLS
		srvListener = listener.CreateTLSListener(
			srvListener, testCerts.ServerCert, testCerts.CaCert)

		go capnpserver.StartGatewayCapnpServer(srvListener, svc)

		// connect the client to the server above
		addr := srvListener.Addr().String()
		gwClient, err := capnpclient.ConnectToGatewayTLS(
			"", addr, testCerts.PluginCert, testCerts.CaCert)

		return gwClient, func() error {
			gwClient.Release()
			time.Sleep(time.Millisecond)
			srvListener.Close()
			time.Sleep(time.Millisecond)
			resolverStop()
			return err
		}
	}
	// unfortunately can't do this without capnp
	//return svc, func() error {
	return nil, func() error {
		resolverStop()
		return err
	}
}
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	testCerts = testenv.CreateCertBundle()

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	clientInfo, err := svc.Ping(ctx)
	assert.NoError(t, err)
	assert.Equal(t, hubapi.ClientTypeUnauthenticated, clientInfo.ClientType)

	err = stopFn()
	assert.NoError(t, err)
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)
	defer stopFn()

	authToken, refreshToken, err := svc.Login(ctx, testClientID, "password")
	require.NoError(t, err)
	assert.NotEmpty(t, authToken)
	assert.NotEmpty(t, refreshToken)

	clientInfo, err := svc.Ping(ctx)
	assert.NoError(t, err)
	assert.Equal(t, hubapi.ClientTypeUser, clientInfo.ClientType)
	assert.Equal(t, testClientID, clientInfo.ClientID)
}

func TestGetInfo(t *testing.T) {

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)
	_ = svc
	_ = ctx

	// use test service to get the capability
	ts := captest.NewTestService()
	err := ts.Start(testSocketPath)
	assert.NoError(t, err)

	ts.Stop()
	time.Sleep(time.Millisecond)
	err = stopFn()
	assert.NoError(t, err)
}

func TestGetCapability(t *testing.T) {
	const clientType = hubapi.ClientTypeService
	ctx := context.Background()

	svc, stopFn := startService(testUseCapnp)

	ts := captest.NewTestService()
	ts.Start(testSocketPath)

	// list capabilities
	capList, err := svc.ListCapabilities(ctx)
	assert.NoError(t, err)
	require.Equal(t, 1, len(capList))

	// the connection method determines the client type. In this case service.
	capability, err := svc.GetCapability(ctx,
		testClientID, clientType, capList[0].CapabilityName, nil)
	require.NoError(t, err)
	assert.NotNil(t, capability)

	// invoke the test method
	method1Capability := captest.CapMethod1Service(capability)
	method1, release1 := method1Capability.Method1(ctx, nil)
	assert.NoError(t, err)
	resp, err := method1.Struct()
	assert.NoError(t, err)
	fy, _ := resp.ForYou()
	assert.NotEmpty(t, fy)
	release1()

	ts.Stop()
	err = stopFn()
	assert.NoError(t, err)
}

func TestGetCapabilityNotExists(t *testing.T) {
	const clientType = hubapi.ClientTypeService
	ctx := context.Background()

	svc, stopFn := startService(testUseCapnp)

	// get capability that doesn't exist
	capability, err := svc.GetCapability(ctx, testClientID, clientType, "notacapability", nil)
	if err == nil {
		err = capability.Resolve(ctx)
	}
	assert.False(t, capability.IsValid())
	//assert.Error(t, err)

	err = stopFn()
	assert.NoError(t, err)
}
