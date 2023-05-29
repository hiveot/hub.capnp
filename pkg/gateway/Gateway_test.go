package gateway_test

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"
	"fmt"
	"github.com/hiveot/hub/pkg/gateway"
	capnpclient2 "github.com/hiveot/hub/pkg/resolver/capnpclient"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/dummy"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/lib/testenv"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
	"github.com/hiveot/hub/pkg/resolver"
	capnpserver2 "github.com/hiveot/hub/pkg/resolver/capnpserver"
	service2 "github.com/hiveot/hub/pkg/resolver/service"
)

const testSocketDir = "/tmp/test-gateway"
const testClientID = "client1"
const testUseCapnp = true // can't disable. capnp is needed
const testUseWS = false

var resolverSocketPath = path.Join(testSocketDir, resolver.ServiceName+".socket")
var testServiceSocketPath = path.Join(testSocketDir, "testService.socket")

var testGatewayURL = ""

// CA, server and plugin test certificate
var testCerts = testenv.CreateCertBundle()

// start resolver and register a dummy authn
func startResolver() (resolverClient *capnpclient2.ResolverCapnpClient, stopfn func()) {
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
	boot, _ := hubclient.ConnectWithCapnpUDS("", resolverSocketPath)
	capclient := capnpclient2.NewResolverCapnpClient(boot)

	return capclient, func() {
		_ = lis.Close()
		_ = svc.Stop()
	}
}

func StartTestGateway(useCapnp bool) (gw gateway.IGatewayService, stopFn func()) {
	_ = os.RemoveAll(testSocketDir)
	_ = os.MkdirAll(testSocketDir, 0700)

	//if err != nil {
	//	panic(err)
	//}
	// authn service is needed for login
	var authnDummy authn.IAuthnService = dummy.NewDummyAuthnService()
	userAuthn, _ := authnDummy.CapUserAuthn(context.Background(), gateway.ServiceName)
	//
	resolverSvc, stopResolver := startResolver()
	time.Sleep(time.Second)
	svc := service.NewGatewayService(resolverSvc, userAuthn)
	err := svc.Start()
	if err != nil {
		panic(err)
	}

	// optionally test with capnp RPC over TLS
	if useCapnp {
		var capClient capnp.Client

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
		srvListener = listener.CreateTLSListener(srvListener, testCerts.ServerCert, testCerts.CaCert)
		// TODO: cleanup the need for wsPath
		go capnpserver.StartGatewayServiceCapnpServer(svc, srvListener, wsPath)

		time.Sleep(time.Millisecond)

		// --- connect the client to the server above, using the same service certificate
		testGatewayAddr := srvListener.Addr().String()
		if testUseWS {
			testGatewayURL = fmt.Sprintf("wss://%s%s", testGatewayAddr, wsPath)
			capClient, err = hubclient.ConnectWithCapnpWebsockets(testGatewayURL, testCerts.DeviceCert, testCerts.CaCert)
		} else {
			testGatewayURL = fmt.Sprintf("tcp://%s/", testGatewayAddr)
			capClient, err = hubclient.ConnectWithCapnpTCP(testGatewayURL, testCerts.DeviceCert, testCerts.CaCert)
		}

		gwClient := capnpclient.NewGatewayServiceCapnpClient(capClient)
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
	svc, stopFn := StartTestGateway(testUseCapnp)

	token := svc.AuthNoAuth("itsme")
	assert.NotEmpty(t, token)
	stopFn()
	time.Sleep(time.Millisecond)
}

func TestLogin(t *testing.T) {
	svc, stopFn := StartTestGateway(testUseCapnp)
	defer stopFn()

	authToken := svc.AuthWithPassword(testClientID, "password")
	assert.NotEmpty(t, authToken)

	session, err := svc.NewSession(authToken)
	require.NoError(t, err)
	assert.NotEmpty(t, session)

	session.Release()
}

func TestRefresh(t *testing.T) {
	svc, stopFn := StartTestGateway(testUseCapnp)
	defer stopFn()

	authToken := svc.AuthWithPassword(testClientID, "password")
	assert.NotEmpty(t, authToken)

	newToken := svc.AuthRefresh(testClientID, authToken)
	assert.NotEmpty(t, newToken)
}
func TestGetInfo(t *testing.T) {

	ctx := context.Background()
	svc, stopFn := StartTestGateway(testUseCapnp)
	defer stopFn()
	_ = svc
	_ = ctx

	// use test service to get the capability
	ts := testenv.NewTestService()
	err := ts.Start(testServiceSocketPath)
	assert.NoError(t, err)
	// give the resolver time to discover the test service
	time.Sleep(time.Millisecond * 100)

	// the client is logged in with a client cert and should be recognized as a service
	// the given client type is ignored
	authToken := svc.AuthWithCert()
	assert.NotEmpty(t, authToken)
	session, err := svc.NewSession(authToken)
	assert.NoError(t, err)

	capList, err := session.ListCapabilities(ctx)
	assert.NoError(t, err)
	require.Equal(t, 1, len(capList))

	ts.Stop()
	time.Sleep(time.Millisecond)
}

func TestGetCapability(t *testing.T) {
	const serviceID1 = "service1"
	var rpcConn *rpc.Conn
	var gwClient2 capnp.Client
	ctx := context.Background()

	// Phase 1 - setup environment with a resolver and test service
	svc, stopFn := StartTestGateway(testUseCapnp)
	defer stopFn()
	authToken := svc.AuthWithCert()

	// register a test service
	ts := testenv.NewTestService()
	err := ts.Start(testServiceSocketPath)
	assert.NoError(t, err)

	// give the resolver time to discover the test service
	time.Sleep(time.Millisecond * 10)

	// Phase 2 - obtain the test service capability from the gateway/resolver
	session, err := svc.NewSession(authToken)
	assert.NoError(t, err)
	caps, err := session.ListCapabilities(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(caps))

	// Phase 3 - obtain the test service capability for method1 using the gateway connection
	// use the gateway as a proxy for the test service
	if testUseWS {
		gwClient2, err = hubclient.ConnectWithCapnpWebsockets(testGatewayURL, testCerts.ServerCert, testCerts.CaCert)

	} else {
		gwClient2, err = hubclient.ConnectWithCapnpTCP(testGatewayURL, testCerts.ServerCert, testCerts.CaCert)
	}
	assert.NoError(t, err)
	_ = rpcConn
	capability := testenv.CapTestService(gwClient2)

	method1, release1 := capability.CapMethod1(ctx,
		func(params testenv.CapTestService_capMethod1_Params) error {
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
	method2, release2 := capMethod1.Method1(ctx, nil)
	// get the result
	resp, err3 := method2.Struct()
	assert.NoError(t, err3)

	msg1, _ := resp.ForYou()
	assert.NotEmpty(t, msg1)
	t.Logf("Received method1 response from testserver: %s", msg1)
	// release the method and capability
	release2()
	capMethod1.Release()

	// Phase 4 - stop the test service. Its capabilities should disappear
	ts.Stop()
	// remote side needs time to discover disconnect
	time.Sleep(time.Millisecond * 1)
	caps, err = session.ListCapabilities(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(caps))

	// Phase 5 - capabilities should no longer resolve
	// fixme: can we do without the boilerplate please?
	// when the service disconnects the capabilities should disappear
	method3, release3 := capability.CapMethod1(ctx,
		func(params testenv.CapTestService_capMethod1_Params) error {
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

	// get the result should fail as method4 is released
	time.Sleep(time.Millisecond)
	_, err4 := method4.Struct()
	assert.Error(t, err4)
}

//
//func TestGetCapabilityNotExists(t *testing.T) {
//	const authType = hubapi.AuthTypeService
//	ctx := context.Background()
//
//	svc, stopFn := StartTestGateway(testUseCapnp)
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
