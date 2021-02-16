package messenger_test

import (
	"testing"

	messenger "github.com/wostzone/gateway/src/messenger/go"
)

const mqttServerHostPort = "localhost:8883"
const mqttCertFolder = "/etc/mosquitto/certs"

// These tests require an MQTT TLS server on localhost with TLS support
func TestMqttConnect(t *testing.T) {
	client := messenger.NewMqttMessenger(mqttCertFolder)
	TMessengerConnect(t, client, mqttServerHostPort)
}

func TestMqttNoConnect(t *testing.T) {
	client := messenger.NewMqttMessenger(mqttCertFolder)
	TMessengerNoConnect(t, client)
}

func TestMQTTPubSub(t *testing.T) {
	client := messenger.NewMqttMessenger(mqttCertFolder)
	TMessengerPubSubNoTLS(t, client, mqttServerHostPort)
}

func TestMQTTMultipleSubscriptions(t *testing.T) {
	client := messenger.NewMqttMessenger(mqttCertFolder)
	TMessengerMultipleSubscriptions(t, client, mqttServerHostPort)
}

func TestMQTTBadUnsubscribe(t *testing.T) {
	client := messenger.NewMqttMessenger(mqttCertFolder)
	TMessengerBadUnsubscribe(t, client, mqttServerHostPort)
}

func TestMQTTPubNoConnect(t *testing.T) {
	client := messenger.NewMqttMessenger(mqttCertFolder)
	TMessengerPubNoConnect(t, client)
}

func TestMQTTSubBeforeConnect(t *testing.T) {
	client := messenger.NewMqttMessenger(mqttCertFolder)
	TMessengerSubBeforeConnect(t, client, mqttServerHostPort)
}
