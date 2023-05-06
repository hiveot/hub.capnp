package service

import (
	"capnproto.org/go/capnp/v3"
	"context"
	"crypto/x509"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/mochi-co/mqtt/v2"
)

// MqttSession manages a MQTT client session with the HiveOT gateway
// It is created by the mochi hook on a new incoming connection.
// This session establishes a gateway session on startup and releases it on disconnect.
type MqttSession struct {
	mqttClient   *mqtt.Client
	gwCapClient  capnp.Client
	gwClient     gateway.IGatewaySession
	refreshToken string
}

// OnDisconnect release the gateway session on a disconnect
func (session *MqttSession) OnDisconnect() {
	session.gwCapClient.Release()
}

// Login to the gateway
func (session *MqttSession) Login(loginID, password string) error {
	authToken, refreshToken, err := session.gwClient.Login(context.Background(), loginID, password)
	_ = authToken
	_ = refreshToken
	return err
}

// NewMqttSession starts a new session with the hub gateway
// This uses the client credentials, passed to mqtt, as gateway credentials.
//
//	gwUrl address of the gateway to connect to
//	caCert is optional to ensure a valid connection to the gateway
//	client is the mqtt instance of the client connection
//
// Returns a session instance or an error if the gateway connection fails
func NewMqttSession(
	gwUrl string, caCert *x509.Certificate, client *mqtt.Client) (*MqttSession, error) {

	// TODO: use client credentials
	gwCapClient, err := hubclient.ConnectWithCapnpTCP(gwUrl, nil, caCert)

	session := &MqttSession{
		mqttClient:  client,
		gwCapClient: gwCapClient,
		gwClient:    capnpclient.NewGatewaySessionCapnpClient(gwCapClient),
	}

	return session, err
}
