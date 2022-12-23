package resolver_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/captest"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
	"github.com/hiveot/hub/pkg/resolver/capnpserver"
	"github.com/hiveot/hub/pkg/resolver/client"
	"github.com/hiveot/hub/pkg/resolver/service"
)

const testResolverSocket = "/tmp/test-resolver.socket"
const testUseCapnp = true

func startResolverAndClient(useCapnp bool) (resolver.IResolverSession, func() error) {
	var session resolver.IResolverSession

	ctx, cancelFunc := context.WithCancel(context.Background())
	_ = os.RemoveAll(testResolverSocket)

	svc := service.NewResolverService()
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
		go capnpserver.StartResolverCapnpServer(srvListener, svc)
		time.Sleep(time.Millisecond)

		// connect the client to the server above
		capClient, err := capnpclient.ConnectToResolver(testResolverSocket)
		return capClient, func() error {
			capClient.Release()
			time.Sleep(time.Millisecond)
			err = svc.Stop()
			cancelFunc()
			return err
		}
	} else {
		session = svc.OnIncomingConnection(nil)
	}

	return session, func() error {
		cancelFunc()
		err = svc.Stop()
		//certSvc.Stop()
		return err
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

	err := stopFn()
	assert.NoError(t, err)
}

// test that the server detects clients disconnecting
func TestConnectDisconnectClients(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()
	svc, stopFn := startResolverAndClient(testUseCapnp)

	require.NotNil(t, svc)
	capInfo1, err := svc.ListCapabilities(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, capInfo1)

	// second connection
	cl2, err := capnpclient.ConnectToResolver(testResolverSocket)
	assert.NoError(t, err)
	capInfo2, err := cl2.ListCapabilities(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, capInfo2)

	// after releasing the client this should fail
	cl2.Release()
	_, err = cl2.ListCapabilities(ctx)
	assert.Error(t, err)
	time.Sleep(time.Millisecond)

	err = stopFn()
	assert.NoError(t, err)
	//time.Sleep(time.Second)
}

// test that the server detects provider disconnecting and releases its capabilities
func TestConnectDisconnectProviders(t *testing.T) {
	ctx := context.Background()
	svc, stopFn := startResolverAndClient(testUseCapnp)
	assert.NotNil(t, svc)

	// create the client and a registration
	cl, err := capnpclient.ConnectToResolver(testResolverSocket)
	assert.NoError(t, err)

	// register a capability provider
	regSrv := client.NewCapRegistrationServer("testservice1", nil)
	err = cl.RegisterCapabilities(ctx, "test", nil, regSrv.Provider())
	assert.NoError(t, err)

	// after releasing
	cl.Release()
	time.Sleep(time.Millisecond)

	err = stopFn()
	assert.NoError(t, err)
	time.Sleep(time.Second)
}

func TestGetCapability(t *testing.T) {
	//capability1 := "cap1"
	//serviceID1 := "service1"

	ctx := context.Background()
	svc, stopFn := startResolverAndClient(testUseCapnp)

	// register a test service
	ts := captest.NewTestService()
	err := ts.Start(testResolverSocket)
	assert.NoError(t, err)

	// give the test service time to register with the resolver
	time.Sleep(time.Millisecond * 10)

	// obtain the test service capability from the resolver
	caps, err := svc.ListCapabilities(ctx)
	require.NoError(t, err)
	if assert.Equal(t, 1, len(caps)) {

		// obtain the test service capability for method1
		capMethod1Client, err2 := svc.GetCapability(ctx, "test", hubapi.ClientTypeService,
			caps[0].CapabilityName, nil)
		assert.NoError(t, err2)
		// cast the actual type
		capMethod1 := captest.CapMethod1Service(capMethod1Client)
		// invoke method 1
		method1, release := capMethod1.Method1(ctx, nil)
		// get the result
		resp, err3 := method1.Struct()
		assert.NoError(t, err3)

		msg1, _ := resp.ForYou()
		assert.NotEmpty(t, msg1)
		// release the method
		release()
		// release the capability
		capMethod1.Release()
	}

	// when the service disconnects the capabilities should disappear
	ts.Stop()
	// remote side needs time to discover disconnect
	time.Sleep(time.Millisecond * 1)
	//
	c1, err := svc.GetCapability(
		ctx, "test", hubapi.ClientTypeService, caps[0].CapabilityName, nil)
	err = c1.Resolve(ctx)
	c2 := captest.CapMethod1Service(c1)
	m2, r2 := c2.Method1(ctx, nil)
	_, err = m2.Struct()
	r2()
	assert.Error(t, err)

	//
	caps, err = svc.ListCapabilities(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(caps))

	err = stopFn()
	assert.NoError(t, err)
}
