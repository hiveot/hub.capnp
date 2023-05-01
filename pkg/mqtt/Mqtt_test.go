package gateway_test

import (
	"fmt"
	"github.com/hiveot/hub/lib/dummy"
	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/lib/testenv"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/mqtt/mqttclient"
	"github.com/hiveot/hub/pkg/mqtt/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

const testSocketDir = "/tmp/test-mqtt"
const testClientID = "client1"
const testMqttWSPort = 9331
const testMqttTcpPort = 9332

var testGatewayURL = ""

// CA, service, device and user test certificate
var testCerts testenv.TestCerts = testenv.CreateCertBundle()

// start a dummy gateway
// returns the gateway URL to connect to
func startDummyGateway() (dummyGw *dummy.DummyGateway, url string, err error) {
	dummyGw = dummy.NewDummyGateway()
	url, err = dummyGw.Start(testCerts)
	return dummyGw, url, err
}

// start the mqtt service for testing
func startService() (stopFn func()) {
	_ = os.RemoveAll(testSocketDir)
	_ = os.MkdirAll(testSocketDir, 0700)

	gw, gwURL, err := startDummyGateway()
	if err != nil {
		panic(err)
	}
	//
	svc := service.NewMqttGatewayService()
	go svc.Start(testMqttTcpPort, testMqttWSPort, testCerts.ServerCert, testCerts.CaCert, gwURL)

	time.Sleep(time.Millisecond)
	return func() {
		_ = svc.Stop()
		gw.Stop()
	}
}
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	stopFn := startService()
	stopFn()
}

func TestLogin(t *testing.T) {
	stopFn := startService()
	stopFn()
}

func TestRefresh(t *testing.T) {
	t.Error(t, "notimplemented")
}

func TestPubSubEvent(t *testing.T) {
	var mqttUrl = fmt.Sprintf("tls://127.0.0.1:%d", testMqttTcpPort)
	var loginID = "test"
	var password = "test"
	const publisher1ID = "urn:device1"
	const thing1ID = "urn:thing1"
	const user1ID = "urn:user"
	const event1Name = "event1"
	var event1Message = "message1"
	var event1Count = int32(0)
	var receivedMsg []byte

	stopFn := startService()
	defer stopFn()

	cl := mqttclient.NewHubMqttClient()
	err := cl.Connect(mqttUrl, loginID, password)
	require.NoError(t, err)

	//
	err = cl.SubEvent(publisher1ID, thing1ID, event1Name,
		func(val thing.ThingValue) {
			atomic.AddInt32(&event1Count, 1)
			receivedMsg = val.Data
		})
	assert.NoError(t, err)

	err = cl.PubEvent(thing1ID, event1Name, event1Message)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond)
	assert.Equal(t, 1, event1Count)
	assert.Equal(t, []byte(event1Message), receivedMsg)

	cl.Disconnect()
}
