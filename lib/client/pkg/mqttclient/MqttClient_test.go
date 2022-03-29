package mqttclient_test

import (
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
)

// Use test/mosquitto-test.conf and a client cert port
var mqttCertAddress = fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortCert)

// TODO: test with username/password login
var mqttUnpwAddress = fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortUnpw)

// var mqttWSAddress = fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortWS)

// CA, server and plugin test certificate
var certs testenv.TestCerts

var homeFolder string
var configFolder string

// var homeFolder string

const TEST_TOPIC = "test"

// For running mosquitto in test
const testPluginID = "test-plugin"

// easy cleanup for existing  certificate
// func removeCerts(folder string) {
// 	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(folder, "*.pem")).Output()
// }

// TestMain - launch mosquitto
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	configFolder = path.Join(homeFolder, "config")
	certFolder := path.Join(homeFolder, "certs")
	os.Chdir(homeFolder)

	testenv.SetLogging("info", "")
	certs = testenv.CreateCertBundle()
	mosquittoCmd, err := testenv.StartMosquitto(configFolder, certFolder, &certs)
	if err != nil {
		logrus.Fatalf("Unable to start mosquitto: %s", err)
	}

	result := m.Run()
	testenv.StopMosquitto(mosquittoCmd)
	os.Exit(result)
}

func TestMqttConnectWithCert(t *testing.T) {
	logrus.Infof("--- TestMqttConnectWithCert ---")

	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)
	// reconnect
	err = client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)
	client.Close()
}

func TestMqttConnectNoCert(t *testing.T) {
	logrus.Infof("--- TestMqttConnectNoCert ---")

	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, nil)
	assert.Error(t, err)
	client.Close()
}

func TestMqttConnectWithUnpw(t *testing.T) {
	logrus.Infof("--- TestMqttConnectWithUnpw ---")
	username := "user1"
	password := "user1"

	// FIXME: this used to work using the MQTT protocol port. For some reason that stopped
	// client := mqttclient.NewMqttClient(mqttUnpwAddress, certsclient.CaCert, 0)
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithAccessToken(mqttUnpwAddress, username, password)
	assert.NoError(t, err)
	client.Close()
}

func TestMqttConnectWrongAddress(t *testing.T) {
	logrus.Infof("--- TestMqttConnectWrongAddress ---")

	invalidHost := "nohost:1111"
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	require.NotNil(t, client)
	err := client.ConnectWithClientCert(invalidHost, certs.PluginCert)
	assert.Error(t, err)
	client.Close()
}

func TestMQTTPubSub(t *testing.T) {
	logrus.Infof("--- TestMQTTPubSub ---")

	var rx string
	rxMutex := sync.Mutex{}
	var msg1 = "Hello world"

	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)

	require.NoError(t, err)

	client.Subscribe(TEST_TOPIC, func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx = string(msg)
		logrus.Infof("Received message: %s", msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", TEST_TOPIC)

	err = client.Publish(TEST_TOPIC, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")

	// allow time to receive
	time.Sleep(1000 * time.Millisecond)
	rxMutex.Lock()
	defer rxMutex.Unlock()
	require.Equalf(t, msg1, rx, "Did not receive the message")

	client.Close()
}

func TestMQTTMultipleSubscriptions(t *testing.T) {
	logrus.Infof("--- TestMQTTMultipleSubscriptions ---")

	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	var rx1 string
	var rx2 string
	rxMutex := sync.Mutex{}
	var msg1 = "Hello world 1"
	var msg2 = "Hello world 2"
	// clientID := "test"

	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	require.NoError(t, err)
	handler1 := func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx1 = string(msg)
		logrus.Infof("Received message on handler 1: %s", msg)
	}
	handler2 := func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx2 = string(msg)
		logrus.Infof("Received message on handler 2: %s", msg)
	}
	_ = handler2
	client.Subscribe(TEST_TOPIC, handler1)
	client.Subscribe(TEST_TOPIC, handler2)
	err = client.Publish(TEST_TOPIC, []byte(msg1))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	// tbd
	assert.Equalf(t, "", rx1, "Did not expect a message on handler 1")
	assert.Equalf(t, msg1, rx2, "Did not receive the message on handler 2")
	// after unsubscribe no message should be received by handler 1
	rx1 = ""
	rx2 = ""
	rxMutex.Unlock()
	client.Unsubscribe(TEST_TOPIC)
	err = client.Publish(TEST_TOPIC, []byte(msg2))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Received a message on handler 2 after unsubscribe")
	rx1 = ""
	rx2 = ""
	rxMutex.Unlock()

	client.Subscribe(TEST_TOPIC, handler1)
	err = client.Publish(TEST_TOPIC, []byte(msg2))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, msg2, rx1, "Did not receive a message on handler 1 after subscribe")
	assert.Equalf(t, "", rx2, "Receive the message on handler 2")
	rxMutex.Unlock()

	// when unsubscribing without handler, all handlers should be unsubscribed
	rx1 = ""
	rx2 = ""
	client.Subscribe(TEST_TOPIC, handler1)
	client.Subscribe(TEST_TOPIC, handler2)
	client.Unsubscribe(TEST_TOPIC)
	err = client.Publish(TEST_TOPIC, []byte(msg2))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")
	rxMutex.Unlock()

	client.Close()
}

func TestMQTTBadUnsubscribe(t *testing.T) {
	logrus.Infof("--- TestMQTTBadUnsubscribe ---")

	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	require.NoError(t, err)

	client.Unsubscribe(TEST_TOPIC)
	client.Close()
}

func TestMQTTPubNoConnect(t *testing.T) {
	logrus.Infof("--- TestMQTTPubNoConnect ---")

	// invalidHost := "localhost:1111"
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	var msg1 = "Hello world 1"

	err := client.Publish(TEST_TOPIC, []byte(msg1))
	require.Error(t, err)

	client.Close()
}

func TestMQTTSubBeforeConnect(t *testing.T) {
	logrus.Infof("--- TestMQTTSubBeforeConnect ---")

	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	const msg = "hello 1"
	var rx string
	rxMutex := sync.Mutex{}

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("Received message on handler 1: %s", msg)
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx = string(msg)
	}
	client.Subscribe(TEST_TOPIC, handler1)

	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	err = client.Publish(TEST_TOPIC, []byte(msg))
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equal(t, msg, rx)
	rxMutex.Unlock()

	client.Close()
}

func TestSubscribeWildcard(t *testing.T) {
	logrus.Infof("--- TestSubscribeWildcard ---")
	const testTopic1 = "test/1/5"
	const wildcardTopic = "test/+/#"

	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	const msg = "hello 1"
	var rx string
	rxMutex := sync.Mutex{}

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("Received message on handler 1: %s", msg)
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx = string(msg)
	}
	client.Subscribe(wildcardTopic, handler1)

	// connect after subscribe uses resubscribe
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	err = client.Publish(testTopic1, []byte(msg))
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equal(t, msg, rx)
	rxMutex.Unlock()

	client.Close()
}
