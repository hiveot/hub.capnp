package mqtt_test

import (
	"testing"

	"github.com/wostzone/hub/pkg/messaging/mqtt"
	testhelper "github.com/wostzone/hub/pkg/messaging/testhelper"
)

const mqttServerHostPort = "localhost:8883"
const mqttCertFolder = "/etc/mosquitto/certs"

// These tests require an MQTT TLS server on localhost with TLS support

func TestMqttConnect(t *testing.T) {
	client := mqtt.NewMqttMessenger(mqttCertFolder, mqttServerHostPort)
	testhelper.TMessengerConnect(t, client)
}

func TestMqttNoConnect(t *testing.T) {
	invalidHost := "localhost:1111"
	client := mqtt.NewMqttMessenger(mqttCertFolder, invalidHost)
	testhelper.TMessengerNoConnect(t, client)
}

func TestMQTTPubSub(t *testing.T) {
	client := mqtt.NewMqttMessenger(mqttCertFolder, mqttServerHostPort)
	testhelper.TMessengerPubSub(t, client)
}

func TestMQTTMultipleSubscriptions(t *testing.T) {
	client := mqtt.NewMqttMessenger(mqttCertFolder, mqttServerHostPort)
	testhelper.TMessengerMultipleSubscriptions(t, client)
}

func TestMQTTBadUnsubscribe(t *testing.T) {
	client := mqtt.NewMqttMessenger(mqttCertFolder, mqttServerHostPort)
	testhelper.TMessengerBadUnsubscribe(t, client)
}

func TestMQTTPubNoConnect(t *testing.T) {
	invalidHost := "localhost:1111"
	client := mqtt.NewMqttMessenger(mqttCertFolder, invalidHost)
	testhelper.TMessengerPubNoConnect(t, client)
}

func TestMQTTSubBeforeConnect(t *testing.T) {
	client := mqtt.NewMqttMessenger(mqttCertFolder, mqttServerHostPort)
	testhelper.TMessengerSubBeforeConnect(t, client)
}
