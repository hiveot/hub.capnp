package gateway_test

import (
	"context"
	"net"
	"os"
	"path"
	"syscall"
	"testing"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
)

const configDir = "/tmp/test-gateway"
const testUseCapnp = true
const testMethodName = "ping"
const testService = gateway.ServiceName // testing using this service
var testSocket = path.Join(configDir, gateway.ServiceName+".socket")

// capability of the gateway service for testing
const config1 = `
capabilities:
  Ping:
    clientType:
      - iotdevice
      - service
      - users
`

func startService(useCapnp bool) (gateway.IGatewayService, func() error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	_ = os.RemoveAll(configDir)
	err := os.MkdirAll(configDir, 0700)

	svc := service.NewGatewayService(configDir, configDir)
	err = svc.Start(ctx)
	if err != nil {
		logrus.Panicf("Failed to start with configdir %s", configDir)
	}
	// optionally test with capnp RPC
	if useCapnp {
		// start the capnpserver for the service
		_ = syscall.Unlink(testSocket)
		srvListener, err := net.Listen("unix", testSocket)
		if err != nil {
			logrus.Panic("Unable to create a listener, can't run test")
		}
		go capnpserver.StartGatewayServiceCapnpServer(ctx, srvListener, svc, configDir)

		// connect the client to the server above
		clConn, _ := net.Dial("unix", testSocket)
		capClient, err := capnpclient.NewGatewayServiceCapnpClient(ctx, clConn)
		return capClient, func() error {
			cancelFunc()
			//_ = capClient.Stop(ctx)
			return svc.Stop(ctx)
		}
	}
	return svc, func() error {
		cancelFunc()
		return svc.Stop(ctx)
	}
}
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	ctx, stopFn := context.WithCancel(context.Background())
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)
	svc := service.NewGatewayService(configDir, configDir)
	err = svc.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond)
	pong, err := svc.Ping(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "pong", pong)

	err = svc.Stop(ctx)
	assert.NoError(t, err)
	stopFn()
}

func TestExtractCapConfig(t *testing.T) {
	const service1 = "service1"
	ctx, stopFn := context.WithCancel(context.Background())
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	svc := service.NewGatewayService(configDir, configDir)
	err = svc.Start(ctx)
	assert.NoError(t, err)
	// extract capabilities
	caps := svc.ExtractCapabilitiesFromConfig(service1, []byte(config1))
	assert.Equal(t, 1, len(caps))
	assert.Equal(t, 3, len(caps[0].ClientType))
	err = svc.Stop(ctx)
	assert.NoError(t, err)
	stopFn()
}

func TestGetInfo(t *testing.T) {

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	// create a service configuration file with a capabilities section
	// the service config watcher should pick this up
	configFile := path.Join(configDir, testService+".yaml")
	err := os.WriteFile(configFile, []byte(config1), 0600)
	assert.NoError(t, err)
	// let the watcher renew
	time.Sleep(time.Millisecond)

	// expect 1 capability from testconfig
	info, err := svc.GetGatewayInfo(ctx)
	assert.NoError(t, err)
	require.Equal(t, 1, len(info.Capabilities))
	capInfo0 := info.Capabilities[0]
	assert.Equal(t, testMethodName, capInfo0.Name)

	err = stopFn()
	assert.NoError(t, err)
}

func TestGetCapability(t *testing.T) {
	ctx := context.Background()

	svc, stopFn := startService(testUseCapnp)
	// create a service configuration file with a capabilities section
	configFile := path.Join(configDir, testService+".yaml")
	err := os.WriteFile(configFile, []byte(config1), 0600)
	assert.NoError(t, err)

	// give the service some time to load the new config
	time.Sleep(time.Millisecond)

	// The 'Ping' capability should be accessible
	capability, err := svc.GetCapability(ctx,
		gateway.ClientTypeIotDevice, gateway.ServiceName)
	assert.NoError(t, err)
	assert.NotNil(t, capability)

	// this applies to all services. Just using the gateway service itself for testing
	gwSvc := capnpclient.NewGatewayServiceFromCapability(capability.(capnp.Client))
	resp, err := gwSvc.Ping(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "pong", resp)

	// get capability that doesn't exist
	capability, err = svc.GetCapability(ctx, gateway.ClientTypeIotDevice, "notaservice")
	assert.Error(t, err)
	assert.Nil(t, capability)

	err = stopFn()
	assert.NoError(t, err)
}
