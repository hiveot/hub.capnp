package messaging_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/pkg/messaging"
)

// Test create the Simple Message Bus protocol connections
func TestNewSmbusConnection(t *testing.T) {
	clientID := "test"
	serverAddr := "localhost"
	smbusCertFolder := "../../test/certs"
	logrus.Info("Testing create channels")
	gwc := messaging.NewGatewayConnection(messaging.ConnectionProtocolSmbus, smbusCertFolder, serverAddr)
	gwc.Connect(clientID, 1)
	gwc.Disconnect()
	// _ = gwc
}

func TestInvalidProtocol(t *testing.T) {
	serverAddr := "localhost"
	gwc := messaging.NewGatewayConnection("invalid", "", serverAddr)
	require.Nil(t, gwc)
}

func TestNewMqttConnection(t *testing.T) {
	clientID := "test"
	serverAddr := "localhost:8883"
	certFolder := "/etc/mosquitto/certs"
	gwc := messaging.NewGatewayConnection(messaging.ConnectionProtocolMQTT, certFolder, serverAddr)
	require.NotNil(t, gwc)
	err := gwc.Connect(clientID, 1)
	assert.NoError(t, err)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	gwc.Disconnect()
}
