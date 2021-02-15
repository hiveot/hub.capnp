package messenger_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	messenger "github.com/wostzone/gateway/src/messenger/go"
)

const gwCertFolder = "../../test"

// Test create the Internal Service Bus protocol connections
func TestGWISBConnection(t *testing.T) {
	certFolder := "../../test"
	clientID := "test"
	serverAddr := "localhost"
	logrus.Info("Testing create channels")
	gwc := messenger.NewGatewayConnection(messenger.ConnectionProtocolISB, certFolder)
	gwc.Connect(serverAddr, clientID, 1)
	gwc.Disconnect()
	// _ = gwc
}

func TestInvalidProtocol(t *testing.T) {
	certFolder := "../../test"

	gwc := messenger.NewGatewayConnection("invalid", certFolder)
	require.Nil(t, gwc)
}

func TestGWMqttConnection(t *testing.T) {
	clientID := "test"
	serverAddr := "localhost:8883"
	certFolder := "/etc/mosquitto/certs"
	gwc := messenger.NewGatewayConnection(messenger.ConnectionProtocolMQTT, certFolder)
	require.NotNil(t, gwc)
	err := gwc.Connect(serverAddr, clientID, 1)
	assert.NoError(t, err)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	gwc.Disconnect()
}
