package gateway_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/api/go/vocab"
	"github.com/hiveot/hub/lib/dummy"
	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/lib/resolver"
	"github.com/hiveot/hub/lib/testenv"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/bucketstore/kvbtree"
	"github.com/hiveot/hub/pkg/directory"
	service3 "github.com/hiveot/hub/pkg/directory/service"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/config"
	service4 "github.com/hiveot/hub/pkg/history/service"
	"github.com/hiveot/hub/pkg/mqttgw/mqttclient"
	"github.com/hiveot/hub/pkg/mqttgw/service"
	"github.com/hiveot/hub/pkg/pubsub"
	service2 "github.com/hiveot/hub/pkg/pubsub/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
	"time"
)

const testSocketDir = "/tmp/test-mqttgw"
const testUserID = "urn:user1"
const testDeviceID = "urn:device1"
const testPublisherID = "urn:pub1"
const testThingID = "urn:thing1"
const testServiceID = "urn:service1"
const testPassword = "test1"
const testMqttTcpPort = 9331
const testMqttWSPort = 9332

// var testGatewayURL = ""
var testHistConfig = config.HistoryConfig{
	Backend:   "",
	Directory: "",
	ServiceID: history.ServiceName,
	Retention: nil,
}

// CA, service, device and user test certificate
var testCerts testenv.TestCerts = testenv.CreateCertBundle()

// start the mqttgw service for testing
func startService() (stopFn func()) {
	_ = os.RemoveAll(testSocketDir)
	_ = os.MkdirAll(testSocketDir, 0700)
	//
	svc := service.NewMqttGatewayService()
	go svc.Start(testMqttTcpPort, testMqttWSPort, testCerts.ServerCert, testCerts.CaCert)

	time.Sleep(time.Millisecond * 10)
	return func() {
		_ = svc.Stop()
	}
}

// setup a test environment with capabilities for pubsub, directory and history
// This returns a stop function
func startTestEnv() func() {
	// setup the global resolver with dummies for required capabilities
	dummyStore := kvbtree.NewKVStore("mqtttest", "")
	_ = dummyStore.Open()

	// auth
	dummyAuthn := dummy.NewDummyAuthnService()
	userAuthn, _ := dummyAuthn.CapUserAuthn(nil, testUserID)

	// pubsub is in-memory
	dummyPubSub := service2.NewPubSubService()
	_ = dummyPubSub.Start()
	devicePubSub, _ := dummyPubSub.CapDevicePubSub(nil, testDeviceID)
	userPubSub, _ := dummyPubSub.CapUserPubSub(nil, testUserID)
	servicePubSub, _ := dummyPubSub.CapServicePubSub(nil, testServiceID)

	// use a directory and history using a dummy store
	dummyDirSvc := service3.NewDirectoryService(directory.ServiceName, dummyStore, servicePubSub)
	_ = dummyDirSvc.Start()
	readDir, _ := dummyDirSvc.CapReadDirectory(nil, testUserID)
	dummyHistSvc := service4.NewHistoryService(nil, dummyStore, servicePubSub)
	_ = dummyHistSvc.Start()
	readHist, _ := dummyHistSvc.CapReadHistory(nil, testUserID)

	//resolver.RegisterService[gateway.IGatewaySession](dummyGwSession)
	resolver.RegisterService[authn.IAuthnService](dummyAuthn)
	resolver.RegisterService[authn.IUserAuthn](userAuthn)
	resolver.RegisterService[pubsub.IPubSubService](dummyPubSub)
	resolver.RegisterService[pubsub.IDevicePubSub](devicePubSub)
	resolver.RegisterService[pubsub.IUserPubSub](userPubSub)
	resolver.RegisterService[directory.IDirectory](dummyDirSvc)
	resolver.RegisterService[directory.IReadDirectory](readDir)
	resolver.RegisterService[history.IReadHistory](readHist)

	return func() {
		_ = dummyStore.Close()
		dummyDirSvc.Release()
		_ = dummyHistSvc.Stop()
		_ = dummyPubSub.Stop()
	}
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	stopFn := startTestEnv()

	res := m.Run()

	stopFn()
	os.Exit(res)
}

// connect and login using TCP
func TestStartStopTcp(t *testing.T) {
	stopFn := startService()
	url := fmt.Sprintf("tls://127.0.0.1:%d", testMqttTcpPort)
	cl := mqttclient.NewHubMqttClient()
	err := cl.Connect(url, testUserID, testPassword, nil, testCerts.CaCert)
	assert.NoError(t, err)
	cl.Disconnect()
	stopFn()
}

// connect and login using websockets
func TestStartStopWs(t *testing.T) {
	stopFn := startService()
	url := fmt.Sprintf("wss://127.0.0.1:%d", testMqttWSPort)
	cl := mqttclient.NewHubMqttClient()
	err := cl.Connect(url, testUserID, "", nil, testCerts.CaCert)
	assert.NoError(t, err)
	cl.Disconnect()
	stopFn()
}

//
//// login with a previous refresh token
//func TestTokenLogin(t *testing.T) {
//	url := fmt.Sprintf("tls://127.0.0.1:%d", testMqttTcpPort)
//	stopFn := startService()
//	cl := mqttclient.NewHubMqttClient()
//	err := cl.Connect(url, testUserID, "", nil, testCerts.CaCert)
//	require.NoError(t, err)
//
//	cl.Disconnect()
//	stopFn()
//}

//func TestRefresh(t *testing.T) {
//	assert.Fail(t, "notimplemented")
//}

func TestPubSubAction(t *testing.T) {
	var mqttUrl = fmt.Sprintf("tls://127.0.0.1:%d", testMqttTcpPort)
	var action1Name = "action1"
	var action1Payload = []byte("this is action 1")
	var mux sync.RWMutex
	var action1Count = 0
	var receivedMsg thing.ThingValue

	stopFn := startService()
	defer stopFn()

	// device subscribes to actions from a user device
	cl := mqttclient.NewHubMqttClient()
	err := cl.Connect(mqttUrl, testDeviceID, "", testCerts.DeviceCert, testCerts.CaCert)
	require.NoError(t, err)

	// test
	err = cl.SubAction(testThingID, "",
		func(val thing.ThingValue) {
			logrus.Infof("Received action: %s", val)
			mux.Lock()
			defer mux.Unlock()
			action1Count++
			receivedMsg = val
		})
	assert.NoError(t, err)
	// wait for subscription to complete
	//time.Sleep(time.Millisecond * 1)

	// test user publishes an action
	cl2 := mqttclient.NewHubMqttClient()
	err2 := cl2.Connect(mqttUrl, testUserID, testPassword, nil, testCerts.CaCert)
	require.NoError(t, err2)
	err2 = cl2.PubAction(testDeviceID, testThingID, action1Name, action1Payload)
	assert.NoError(t, err2)

	// wait for the background processes to start
	time.Sleep(time.Millisecond * 10)

	// expect the device to receive the action message
	mux.Lock()
	require.Equal(t, 1, action1Count)
	assert.Equal(t, testDeviceID, receivedMsg.PublisherID)
	assert.Equal(t, testThingID, receivedMsg.ThingID)
	assert.Equal(t, action1Name, receivedMsg.ID)
	assert.Equal(t, action1Payload, receivedMsg.Data)
	mux.Unlock()
	cl.Disconnect()
	cl2.Disconnect()
}

func TestPubSubEvent(t *testing.T) {
	var mqttUrl = fmt.Sprintf("tls://127.0.0.1:%d", testMqttTcpPort)
	var event1Name = "event1"
	var event1Message = []byte("message1")
	var mux sync.RWMutex
	var event1Count = 0
	var receivedMsg thing.ThingValue

	stopFn := startService()
	defer stopFn()

	// test user subscribes to events from a test device
	cl := mqttclient.NewHubMqttClient()
	err := cl.Connect(mqttUrl, testUserID, testPassword, nil, testCerts.CaCert)
	require.NoError(t, err)

	// test
	err = cl.SubEvent("", "", "",
		func(val thing.ThingValue) {
			logrus.Infof("Received event: %s", val)
			mux.Lock()
			defer mux.Unlock()
			event1Count++
			receivedMsg = val
		})
	assert.NoError(t, err)
	// wait for subscription to complete
	//time.Sleep(time.Millisecond * 1)

	// test device publishes an event
	cl2 := mqttclient.NewHubMqttClient()
	err2 := cl2.Connect(mqttUrl, testDeviceID, "", testCerts.DeviceCert, testCerts.CaCert)
	require.NoError(t, err2)
	err2 = cl2.PubEvent(testThingID, event1Name, event1Message)
	assert.NoError(t, err2)

	// wait for the background processes to start
	time.Sleep(time.Millisecond * 10)

	// expect the user to receive the device message
	mux.Lock()
	require.Equal(t, 1, event1Count)
	assert.Equal(t, testDeviceID, receivedMsg.PublisherID)
	assert.Equal(t, testThingID, receivedMsg.ThingID)
	assert.Equal(t, event1Name, receivedMsg.ID)
	assert.Equal(t, []byte(event1Message), receivedMsg.Data)
	mux.Unlock()
	cl.Disconnect()
	cl2.Disconnect()
}

func TestUpdateReadDirectory(t *testing.T) {
	var mqttUrl = fmt.Sprintf("tls://127.0.0.1:%d", testMqttTcpPort)
	var completed []thing.ThingValue
	respChan := make(chan []thing.ThingValue)

	stopFn := startService()
	defer stopFn()

	// test device publishes a TD via MQTT
	td1 := thing.NewTD(testThingID, "test thing", vocab.DeviceTypeThermometer)
	td1Json, _ := json.Marshal(&td1)
	deviceClient := mqttclient.NewHubMqttClient()
	err := deviceClient.Connect(mqttUrl, testDeviceID, "", testCerts.DeviceCert, testCerts.CaCert)
	require.NoError(t, err)
	defer deviceClient.Disconnect()
	err = deviceClient.PubEvent(testThingID, hubapi.EventNameTD, td1Json)
	assert.NoError(t, err)

	// read the directory
	userClient := mqttclient.NewHubMqttClient()
	err = userClient.Connect(mqttUrl, testUserID, "", testCerts.UserCert, testCerts.CaCert)
	require.NoError(t, err)
	defer userClient.Disconnect()

	err = userClient.SubReadDirectory(func(response *mqttclient.ReadDirectoryResponse) {
		t.Logf("received directory. %d items", len(response.TDs))
		respChan <- response.TDs
		close(respChan)
	})
	assert.NoError(t, err)
	err = userClient.PubReadDirectory("")
	assert.NoError(t, err)

	// check results
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	select {
	case completed = <-respChan:
	case <-ctx.Done():
	}
	assert.Greater(t, len(completed), 0)
}

func TestUpdateReadHistory(t *testing.T) {
	const evTemperature = "temperature"
	const evValue = "12.5"

	var mqttUrl = fmt.Sprintf("tls://127.0.0.1:%d", testMqttTcpPort)
	var historyResp mqttclient.ReadHistoryResponse
	var latestResp mqttclient.ReadLatestResponse
	historyChan := make(chan mqttclient.ReadHistoryResponse)
	latestChan := make(chan mqttclient.ReadLatestResponse)

	stopFn := startService()
	defer stopFn()

	// setup the mqttgw client as a device
	deviceClient := mqttclient.NewHubMqttClient()
	err := deviceClient.Connect(mqttUrl, testDeviceID, "", testCerts.DeviceCert, testCerts.CaCert)
	require.NoError(t, err)
	defer deviceClient.Disconnect()

	// publish the temperature for the history service
	err = deviceClient.PubEvent(testThingID, evTemperature, []byte(evValue))
	assert.NoError(t, err)

	// read the history
	userClient := mqttclient.NewHubMqttClient()
	err = userClient.Connect(mqttUrl, testUserID, "", testCerts.UserCert, testCerts.CaCert)
	require.NoError(t, err)
	defer userClient.Disconnect()

	err = userClient.SubReadHistory(func(response *mqttclient.ReadHistoryResponse) {
		t.Logf("received histry. %d items", len(response.Values))
		historyChan <- *response
		close(historyChan)
	})
	err = userClient.SubReadLatest(func(response *mqttclient.ReadLatestResponse) {
		latestChan <- *response
		close(latestChan)
	})

	err = userClient.PubReadHistory(testDeviceID, testThingID, evTemperature, "", 0, 100)
	assert.NoError(t, err)
	err = userClient.PubReadLatest(testDeviceID, testThingID, nil)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond)

	// check results
	historyResp = <-historyChan
	latestResp = <-latestChan

	assert.Equal(t, testDeviceID, historyResp.PublisherID)
	assert.Equal(t, testThingID, historyResp.ThingID)
	assert.Equal(t, evTemperature, historyResp.Name)
	assert.Equal(t, 1, len(historyResp.Values))

	// this should also be the latest value
	assert.Equal(t, testDeviceID, latestResp.PublisherID)
	assert.Equal(t, testThingID, latestResp.ThingID)
	require.Greater(t, len(latestResp.Values), 0)
	//assert.Equal(t, []byte(evValue), latestResp.Values[0].Data)
}
