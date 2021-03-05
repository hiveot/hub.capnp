package messaging_test

import (
	"flag"
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/pkg/config"
	"github.com/wostzone/gateway/pkg/messaging"
	"github.com/wostzone/gateway/pkg/smbserver"
)

// Test create the Simple Message Bus protocol connections
func TestNewSmbClientConnection(t *testing.T) {
	clientID := "test"
	// serverAddr := "localhost:9999"
	// Remove testing package created commandline and flags so we can test ours
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = os.Args[0:1]

	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	gwConfig, err := config.SetupConfig(homeFolder, "", nil)

	// this needs a msbserver to connect to
	assert.NoError(t, err)
	srv, err := smbserver.StartSmbServer(gwConfig)

	logrus.Info("Testing create channels")
	gwc := messaging.NewGatewayConnection(
		messaging.ConnectionProtocolSmbus, gwConfig.Messenger.CertFolder, gwConfig.Messenger.HostPort)
	gwc.Connect(clientID, 1)
	gwc.Disconnect()

	srv.Stop()
	// _ = gwc
}

// func TestInvalidProtocol(t *testing.T) {
// 	serverAddr := "localhost"
// 	gwc := messaging.NewGatewayConnection("invalid", "", serverAddr)
// 	require.Nil(t, gwc)
// }

// Test a MQTT connection. A mqtt server must be running
func TestNewMqttConnection(t *testing.T) {
	clientID := "test"
	serverAddr := "localhost:8883"
	certFolder := "/etc/mosquitto/certs"
	gwc := messaging.NewGatewayConnection(messaging.ConnectionProtocolMQTT, certFolder, serverAddr)
	require.NotNil(t, gwc)
	err := gwc.Connect(clientID, 10)
	assert.NoError(t, err)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	gwc.Disconnect()
}

func TestStartGatewayMessenger(t *testing.T) {
	clientID := "test"
	cwd, _ := os.Getwd()
	// Remove testing package created commandline and flags so we can test ours
	os.Args = os.Args[0:1]
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	homeFolder := path.Join(cwd, "../../test")
	gwConfig, err := config.SetupConfig(homeFolder, clientID, nil)
	assert.NoError(t, err)

	// this needs a msbserver to connect to
	assert.NoError(t, err)
	srv, err := smbserver.StartSmbServer(gwConfig)

	// note that no connection is not a failure as the server can be down at the moment
	m, err := messaging.StartGatewayMessenger(clientID, gwConfig)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	srv.Stop()
}
