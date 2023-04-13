package gateway_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hiveot/hub/pkg/gateway"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/dummy"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/lib/testsvc"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
	"github.com/hiveot/hub/pkg/resolver"
	capnpserver2 "github.com/hiveot/hub/pkg/resolver/capnpserver"
	service2 "github.com/hiveot/hub/pkg/resolver/service"
)

const testSocketDir = "/tmp/test-gateway"
const testClientID = "client1"
const testUseCapnp = true
const testUseWS = false //true

var resolverSocketPath = path.Join(testSocketDir, resolver.ServiceName+".socket")
var testServiceSocketPath = path.Join(testSocketDir, "testService.socket")

var testGatewayURL = ""
var testCACert *x509.Certificate
var testCAKey *ecdsa.PrivateKey

var testServiceKeys *ecdsa.PrivateKey
var testServiceCert tls.Certificate
var testServiceCertPEM string
var testServicePubKeyPEM string
var testServicePrivKeyPEM string

var testClientKeys *ecdsa.PrivateKey
var testClientCert tls.Certificate
var testClientCertPEM string
var testClientPubKeyPEM string
var testClientPrivKeyPEM string

// CA, server and plugin test certificate
//var testCerts testenv.TestCerts

func createCerts() error {
	var err error
	var ctx = context.Background()
	// a CA is needed
	testCACert, testCAKey, _ = selfsigned.CreateHubCA(1)
	certSvc := selfsigned.NewSelfSignedCertsService(testCACert, testCAKey)
	capServiceCert, _ := certSvc.CapServiceCerts(ctx, testClientID)
	capDeviceCert, err := certSvc.CapDeviceCerts(ctx, testClientID)
	if err != nil {
		return err
	}
	// the server cert
	testServiceKeys = certsclient.CreateECDSAKeys()
	testServicePubKeyPEM, _ = certsclient.PublicKeyToPEM(&testServiceKeys.PublicKey)
	testServicePrivKeyPEM, _ = certsclient.PrivateKeyToPEM(testServiceKeys)
	testServiceCertPEM, _, err = capServiceCert.CreateServiceCert(
		ctx, testClientID, testServicePubKeyPEM, []string{"localhost", "127.0.0.1"}, 1)
	if err != nil {
		return err
	}
	testServiceCert, _ = tls.X509KeyPair([]byte(testServiceCertPEM), []byte(testServicePrivKeyPEM))

	// and a client cert, also signed by the CA
	testClientKeys = certsclient.CreateECDSAKeys()
	testClientPubKeyPEM, _ = certsclient.PublicKeyToPEM(&testClientKeys.PublicKey)
	testClientPrivKeyPEM, _ = certsclient.PrivateKeyToPEM(testClientKeys)
	testClientCertPEM, _, err = capDeviceCert.CreateDeviceCert(
		ctx, testClientID, testClientPubKeyPEM, 1)
	if err != nil {
		return err
	}
	testClientCert, err = tls.X509KeyPair([]byte(testClientCertPEM), []byte(testClientPrivKeyPEM))

	return err
}

// start resolver and register a dummy authn
func startResolver() (stopfn func()) {
	_ = os.RemoveAll(resolverSocketPath)
	svc := service2.NewResolverService(testSocketDir)
	err := svc.Start(context.Background())
	if err != nil {
		panic("can't start resolver: " + err.Error())
	}
	lis, err := net.Listen("unix", resolverSocketPath)
	if err != nil {
		panic("need resolver")
	}
	go capnpserver2.StartResolverServiceCapnpServer(svc, lis, svc.HandleUnknownMethod)

	return func() {
		_ = lis.Close()
		_ = svc.Stop()
	}
}

func startService(useCapnp bool) (gwSession gateway.IGatewaySession, stopFn func()) {
	_ = os.RemoveAll(testSocketDir)
	_ = os.MkdirAll(testSocketDir, 0700)
	err := createCerts()
	if err != nil {
		panic(err)
	}
	// authn service is needed for login
	var authnDummy authn.IAuthnService = dummy.NewDummyAuthnService()
	//
	stopResolver := startResolver()
	svc := service.NewGatewayService(resolverSocketPath, authnDummy)
	err = svc.Start()
	if err != nil {
		panic(err)
	}

	// optionally test with capnp RPC over TLS
	if useCapnp {
		wsPath := ""
		if testUseWS {
			// Can't register a path twice. Using "/" seems to work
			wsPath = "/ws" // TODO: use const
		}
		// --- start the capnpserver for the service
		// TLS only works on tcp sockets as address must match certificate name
		// This can be converted to a socket connection if wsPath is set
		srvListener, err2 := net.Listen("tcp", "127.0.0.1:0")

		if err2 != nil {
			logrus.Panicf("Unable to create a listener, can't run test: %s", err2)
		}
		srvListener = listener.CreateTLSListener(srvListener, &testServiceCert, testCACert)
		// TODO: cleanup the need for wsPath
		go capnpserver.StartGatewayCapnpServer(svc, srvListener, wsPath)

		time.Sleep(time.Millisecond)

		// --- connect the client to the server above, using the same service certificate
		testGatewayAddr := srvListener.Addr().String()
		testGatewayURL = fmt.Sprintf("tcp://%s/", testGatewayAddr)
		if testUseWS {
			testGatewayURL = fmt.Sprintf("wss://%s%s", testGatewayAddr, wsPath)
		}

		gwClient, err2 := capnpclient.ConnectToGateway(
			testGatewayURL, 0, &testClientCert, testCACert)
		if err2 != nil {
			panic("unable to connect the client to the gateway:" + err2.Error())
		}

		return gwClient, func() {
			gwClient.Release()
			_ = svc.Stop()
			err = srvListener.Close()
			time.Sleep(time.Millisecond)
			stopResolver()
		}
	}
	// unfortunately can't do this without capnp
	//return svc, func() error {
	return nil, func() {
		stopResolver()
		_ = svc.Stop()
	}
}
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	clientInfo, err := svc.Ping(ctx)
	assert.NoError(t, err)
	assert.Equal(t, hubapi.AuthTypeIotDevice, clientInfo.AuthType)
	stopFn()
	time.Sleep(time.Millisecond)
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
	assert.Equal(t, hubapi.AuthTypeUser, clientInfo.AuthType)
	assert.Equal(t, testClientID, clientInfo.ClientID)
}

func TestRefresh(t *testing.T) {
	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)
	defer stopFn()

	authToken, refreshToken, err := svc.Login(ctx, testClientID, "password")
	require.NoError(t, err)
	assert.NotEmpty(t, authToken)
	assert.NotEmpty(t, refreshToken)

	newAuthToken, newRefreshToken, err := svc.Refresh(ctx, testClientID, refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAuthToken)
	assert.NotEmpty(t, newRefreshToken)
}
func TestGetInfo(t *testing.T) {

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)
	defer stopFn()
	_ = svc
	_ = ctx

	// use test service to get the capability
	ts := testsvc.NewTestService()
	err := ts.Start(testServiceSocketPath)
	assert.NoError(t, err)
	// give the resolver time to discover the test service
	time.Sleep(time.Millisecond * 100)

	// the client is logged in with a client cert and should be recognized as a service
	// the given client type is ignored
	assert.NoError(t, err)
	capList, err := svc.ListCapabilities(ctx)
	assert.NoError(t, err)
	require.Equal(t, 1, len(capList))

	ts.Stop()
	time.Sleep(time.Millisecond)
}

func TestGetCapability(t *testing.T) {
	const serviceID1 = "service1"
	ctx := context.Background()

	// Phase 1 - setup environment with a resolver and test service
	gwClient, stopFn := startService(testUseCapnp)
	defer stopFn()

	// register a test service
	ts := testsvc.NewTestService()
	err := ts.Start(testServiceSocketPath)
	assert.NoError(t, err)

	// give the resolver time to discover the test service
	time.Sleep(time.Millisecond * 10)

	// Phase 2 - obtain the test service capability from the gateway/resolver
	caps, err := gwClient.ListCapabilities(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(caps))

	// Phase 3 - obtain the test service capability for method1 using the gateway connection
	// use the gateway as a proxy for the test service
	// gwClient2, err := capnpclient.ConnectToGatewayProxyClient(
	// 	"tcp", testGatewayAddr, &testServiceCert, testCACert)
	rpcConn, gwClient2, err := hubclient.ConnectToHubClient(
		testGatewayURL, 0, &testServiceCert, testCACert)
	require.NoError(t, err)
	_ = rpcConn
	capability := testsvc.CapTestService(gwClient2)

	method1, release1 := capability.CapMethod1(ctx,
		func(params testsvc.CapTestService_capMethod1_Params) error {
			err2 := params.SetClientID(serviceID1)
			assert.NoError(t, err2)
			_ = params.SetAuthType(hubapi.AuthTypeService)
			return err2
		})
	defer release1()
	m1s, err := method1.Struct()
	require.NoError(t, err)

	capMethod1 := m1s.Capabilit()
	// invoke method 1
	method2, release := capMethod1.Method1(ctx, nil)
	// get the result
	resp, err3 := method2.Struct()
	assert.NoError(t, err3)

	msg1, _ := resp.ForYou()
	assert.NotEmpty(t, msg1)
	t.Logf("Received method1 response from testserver: %s", msg1)
	// release the method and capability
	release()
	capMethod1.Release()

	// Phase 4 - stop the test service. Its capabilities should disappear
	ts.Stop()
	// remote side needs time to discover disconnect
	time.Sleep(time.Millisecond * 1)
	caps, err = gwClient.ListCapabilities(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(caps))

	// Phase 5 - capabilities should no longer resolve
	// fixme: can we do without the boilerplate please?
	// when the service disconnects the capabilities should disappear
	method3, release3 := capability.CapMethod1(ctx,
		func(params testsvc.CapTestService_capMethod1_Params) error {
			err2 := params.SetClientID(serviceID1)
			assert.NoError(t, err2)
			_ = params.SetAuthType(hubapi.AuthTypeService)
			return err2
		})
	defer release3()
	capMethod3 := method3.Capabilit()

	// getting capability should fail
	method4, release4 := capMethod3.Method1(ctx, nil)
	defer release4()
	// get the result
	_, err4 := method4.Struct()

	assert.Error(t, err4)

	release()
}

//
//func TestGetCapabilityNotExists(t *testing.T) {
//	const authType = hubapi.AuthTypeService
//	ctx := context.Background()
//
//	svc, stopFn := startService(testUseCapnp)
//	defer stopFn()
//
//	// get capability that doesn't exist
//	capability, err := svc.GetCapability(ctx, testClientID, authType, "notacapability", nil)
//	if err == nil {
//		err = capability.Resolve(ctx)
//	}
//	assert.False(t, capability.IsValid())
//	//assert.Error(t, err)
//}
