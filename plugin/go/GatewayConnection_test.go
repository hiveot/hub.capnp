package plugin

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gwCertFolder = "../../test"

// Test create the Internal Service Bus protocol connections
func TestGWISBConnection(t *testing.T) {
	certFolder := "../../test"
	clientID := "test"
	serverAddr := "localhost"
	logrus.Info("Testing create channels")
	gwc := NewGatewayConnection(clientID, ConnectionProtocolISB, certFolder)
	gwc.Connect(serverAddr, 1)
	gwc.Disconnect()
	// _ = gwc
}

func TestInvalidProtocol(t *testing.T) {
	clientID := "test"
	certFolder := "../../test"

	gwc := NewGatewayConnection(clientID, "invalid", certFolder)
	require.Nil(t, gwc)
}

func TestGWMqttConnection(t *testing.T) {
	clientID := "test"
	serverAddr := "localhost:8883"
	certFolder := "/etc/mosquitto/certs"
	gwc := NewGatewayConnection(clientID, ConnectionProtocolMQTT, certFolder)
	require.NotNil(t, gwc)
	err := gwc.Connect(serverAddr, 1)
	assert.NoError(t, err)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	gwc.Disconnect()
}
