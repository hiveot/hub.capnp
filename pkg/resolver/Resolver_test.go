package resolver_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/lib/testsvc"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpserver"
	"github.com/hiveot/hub/pkg/resolver/service"
)

const testResolverSocket = "/tmp/test-resolver.socket"
const testSocketDir = "/tmp/test-resolver"
const testServiceSocket = testSocketDir + "/testservice.socket"
const testUseCapnp = true

func startResolverAndClient(useCapnp bool) (resolver.IResolverService, func()) {

	ctx, cancelFunc := context.WithCancel(context.Background())
	_ = os.RemoveAll(testResolverSocket)

	svc := service.NewResolverService(testSocketDir)
	err := svc.Start(ctx)
	if err != nil {
		logrus.Panicf("Failed to start with socket dir %s", testResolverSocket)
	}
	// optionally test with capnp
	if useCapnp {
		// start the capnpserver on the socket
		srvListener, err2 := net.Listen("unix", testResolverSocket)
		if err2 != nil {
			logrus.Panicf("Unable to create a listener, can't run test: %s", err2)
		}
		go capnpserver.StartResolverServiceCapnpServer(svc, srvListener, svc.HandleUnknownMethod)
		time.Sleep(time.Millisecond)

		// connect the client to the server above
		conn, _ := net.Dial("unix", testResolverSocket)
		capClient, err := capnpclient.NewResolverServiceCapnpClient(ctx, conn)
		if err != nil {
			panic("")
		}
		return capClient, func() {
			capClient.Release()
			_ = srvListener.Close()
			time.Sleep(time.Millisecond)
			_ = svc.Stop()
			cancelFunc()
		}
	}

	return svc, func() {
		cancelFunc()
		_ = svc.Stop()
	}
}
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	svc, stopFn := startResolverAndClient(testUseCapnp)
	assert.NotNil(t, svc)

	stopFn()
}

// test that the server detects clients disconnecting
func TestConnectDisconnectClients(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()
	svc, stopFn := startResolverAndClient(testUseCapnp)
	defer stopFn()

	require.NotNil(t, svc)
	capInfo1, err := svc.ListCapabilities(ctx, hubapi.AuthTypeService)
	assert.NoError(t, err)
	assert.NotNil(t, capInfo1)

	// second connection
	conn2, err := net.Dial("unix", testResolverSocket)
	require.NoError(t, err)
	cl2, err := capnpclient.NewResolverServiceCapnpClient(ctx, conn2)
	assert.NoError(t, err)
	capInfo2, err := cl2.ListCapabilities(ctx, hubapi.AuthTypeService)
	assert.NoError(t, err)
	assert.NotNil(t, capInfo2)

	// after releasing the client this should fail
	cl2.Release()
	_, err = cl2.ListCapabilities(ctx, hubapi.AuthTypeService)
	assert.Error(t, err)
	time.Sleep(time.Millisecond)

}

// test that the server detects provider disconnecting and releases its capabilities
func TestConnectDisconnectProviders(t *testing.T) {
	ctx := context.Background()

	svc, stopFn := startResolverAndClient(testUseCapnp)
	assert.NotNil(t, svc)
	defer stopFn()

	// create the client and a registration
	conn2, err := net.Dial("unix", testResolverSocket)
	require.NoError(t, err)
	cl2, err := capnpclient.NewResolverServiceCapnpClient(ctx, conn2)
	assert.NoError(t, err)

	// register a capability provider
	ts := testsvc.NewTestService()
	err = ts.Start(testServiceSocket)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond)
	ts.Stop()

	// after releasing
	cl2.Release()
	time.Sleep(time.Millisecond)
}

// Test accessing the capability directly using GetCapability
func TestGetCapabilityDirect(t *testing.T) {
	//capability1 := "cap1"
	//serviceID1 := "service1"
	listenerSocket := "/tmp/test-resolver-direct.socket"
	ctx := context.Background()
	//svc, stopFn := startResolverAndClient(testUseCapnp)

	// start the test service
	ts := testsvc.NewTestService()
	err := ts.Start(listenerSocket)
	assert.NoError(t, err)
	defer ts.Stop()

	// obtain the test service capability directly from the service without using the resolver.
	// step 1: connect
	conn, err := net.Dial("unix", listenerSocket)
	require.NoError(t, err)
	// step 2: obtain the service bootstrap client
	transport := rpc.NewStreamTransport(conn)
	rpcConn := rpc.NewConn(transport, nil)
	bootClient := rpcConn.Bootstrap(ctx)

	// step 3: convert the bootstrap client to the service client
	capTestSvc := testsvc.CapTestService(bootClient)
	// step 4: obtain the capability for method1 from the service
	method, release := capTestSvc.CapMethod1(ctx, nil)
	defer release()
	capMethod1 := method.Capabilit()
	// step 5: invoke method1
	resp2, release2 := capMethod1.Method1(ctx, nil)
	defer release2()
	resp3, err := resp2.Struct()
	assert.NoError(t, err)
	// finally: the end
	forYouText, err := resp3.ForYou()
	assert.NotEmpty(t, forYouText)
	assert.NoError(t, err)
}

func TestGetCapabilityViaResolver(t *testing.T) {
	//capability1 := "cap1"
	serviceID1 := "service1"

	// Phase 1 - setup environment with a test service
	ctx := context.Background()
	svc, stopFn := startResolverAndClient(testUseCapnp)
	defer stopFn()

	// register a test service
	ts := testsvc.NewTestService()
	err := ts.Start(testServiceSocket)
	assert.NoError(t, err)

	// give the resolver time to discover the test service
	time.Sleep(time.Millisecond * 100)

	// Phase 2 - obtain the test service capability from the resolver
	caps, err := svc.ListCapabilities(ctx, hubapi.AuthTypeService)
	require.NoError(t, err)
	require.Equal(t, 1, len(caps))

	// Phase 3 - obtain the test service capability for method1 using the resolver connection
	resConn, _ := net.Dial("unix", testResolverSocket)
	transport := rpc.NewStreamTransport(resConn)
	rpcConn := rpc.NewConn(transport, nil)
	capability := testsvc.CapTestService(rpcConn.Bootstrap(ctx))

	method, release := capability.CapMethod1(ctx,
		func(params testsvc.CapTestService_capMethod1_Params) error {
			err2 := params.SetClientID(serviceID1)
			assert.NoError(t, err2)
			_ = params.SetAuthType(hubapi.AuthTypeService)
			return err2
		})
	capMethod1 := method.Capabilit()
	// invoke method 1
	method1, release := capMethod1.Method1(ctx, nil)
	// get the result
	resp, err3 := method1.Struct()
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
	time.Sleep(time.Millisecond * 10)
	caps, err = svc.ListCapabilities(ctx, hubapi.AuthTypeService)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(caps))

	// Phase 5 - capabilities should no longer resolve
	// when the service disconnects the capabilities should disappear
	method, release = capability.CapMethod1(ctx,
		func(params testsvc.CapTestService_capMethod1_Params) error {
			err2 := params.SetClientID(serviceID1)
			assert.NoError(t, err2)
			_ = params.SetAuthType(hubapi.AuthTypeService)
			return err2
		})
	capMethod1b := method.Capabilit()

	// invoke method 1
	method1b, release := capMethod1b.Method1(ctx, nil)
	// get the result
	_, err4 := method1b.Struct()

	assert.Error(t, err4)

	release()
}
