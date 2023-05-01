package service

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"crypto/x509"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/mochi-co/mqtt/v2"
)

// MqttSession manages a MQTT client session with the HiveOT gateway
type MqttSession struct {
	mqttClient *mqtt.Client
	rpcConn    *rpc.Conn
	gwClient   capnp.Client
}

func (session *MqttSession) OnDisconnect() {

}

// NewMqttSession starts a new session with the hub gateway
// This assumes the gateway lives on localhost on the default port.
//
//	 gwUrl address of the gateway to connect to
//		caCert is optional to ensure a valid connection to the gateway
//		client is the mqtt instance of the client connection
func NewMqttSession(gwUrl string, caCert *x509.Certificate, client *mqtt.Client) (*MqttSession, error) {
	rpcConn, gwClient, err := hubclient.ConnectWithCapnp(gwUrl, 0, nil, caCert)

	session := &MqttSession{
		mqttClient: client,
		rpcConn:    rpcConn,
		gwClient:   gwClient,
	}

	return session, err
}
