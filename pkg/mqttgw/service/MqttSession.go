package service

import (
	"crypto/x509"
	"fmt"
	"github.com/hiveot/hub/lib/resolver"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/mqttgw/mqttclient"
	"github.com/mochi-co/mqtt/v2"
	"github.com/sirupsen/logrus"
	"strings"
)

// MqttSession manages a MQTT client session with the HiveOT gateway
// It is created by the mochi hook on a new incoming connection.
// This session establishes a gateway session on startup and releases it on disconnect.
// This uses the client resolver to obtain capabilities, which also aids in testing using stubs.
type MqttSession struct {
	mqttClient   *mqtt.Client
	gwClient     gateway.IGatewaySession
	refreshToken string

	userAuthn authn.IUserAuthn
	m2dir     *Mqtt2Directory
	m2hist    *Mqtt2History
	m2pubsub  *Mqtt2PubSub
}

// OnDisconnect release the gateway session on a disconnect
func (session *MqttSession) OnDisconnect() {
	session.m2dir.Release()
	session.m2hist.Release()
	session.m2pubsub.Release()
	if session.gwClient != nil {
		session.gwClient.Release()
	}
}

// LoginWithPassword to the resolver session, most likely the gateway
// This requires that the resolver client is connected to the resolver service.
func (session *MqttSession) LoginWithPassword(loginID, password string) error {
	//session.clientID = loginID
	err := resolver.Login(loginID, password)
	return err
}

// LoginWithCert login to the resolver session using a client certificate
func (session *MqttSession) LoginWithCert(loginID string, peerCert *x509.Certificate) error {
	//session.clientID = loginID
	err := resolver.LoginWithCert(loginID, peerCert)
	return err
}

// OnSubscribe is invoked by the MQTT Hook when the MQTT client requests subscription on a topic.
//
// Thing subscriptions on topic things/{publisherID}/{thingID}/{msgType}/{name} are
// passed on to the pubsub service if they pass the authorization check.
//
// Subscription to service responses are handled by the mqttgw broker and are ignored.
func (session *MqttSession) OnSubscribe(cl *mqtt.Client, mqttTopic string, payload []byte) (err error) {
	logrus.Infof("OnSubscribe to '%s' by client '%s'", mqttTopic, cl.ID)

	if mqttclient.IsThingsTopic(mqttTopic) {
		err = session.m2pubsub.HandleSubscribe(mqttTopic, payload)
	} else if mqttclient.IsDirectoryTopic(mqttTopic) {
		// nothing to do here
		//err = session.m2dir.HandleDirectorySubscribe(mqttTopic, payload)
	} else if mqttclient.IsHistoryTopic(mqttTopic) {
		// nothing to do here
		//err = session.m2hist.HandleHistorySubscribe(mqttTopic, payload)
	} else {
		// unsupported subscription
		err = fmt.Errorf("unsupported subscription to '%s' by client '%s'", mqttTopic, cl.ID)
	}
	return err
}

// OnPublish is invoked by the mqttgw Hook and handles a thing or service publish request.
//
//	This dispatches the request to the Hub's pubsub, directory or history service
//
// # The publisher must be logged in and have permission to publishing
func (session *MqttSession) OnPublish(cl *mqtt.Client, mqttTopic string, payload []byte) (err error) {
	// first time obtain the publish capability
	if strings.HasPrefix(mqttTopic, mqttclient.ThingsTopicPrefix) {
		err = session.m2pubsub.HandlePublish(mqttTopic, payload)
	} else if strings.HasPrefix(mqttTopic, mqttclient.DirectoryTopicPrefix) {
		err = session.m2dir.HandleDirectoryRequest(mqttTopic, payload)
	} else if strings.HasPrefix(mqttTopic, mqttclient.HistoryTopicPrefix) {
		err = session.m2hist.HandleHistoryRequest(mqttTopic, payload)
	} else {
		// not a regular mqttTopic
		err = fmt.Errorf("mqttTopic '%s' is not supported by the MQTT gateway", mqttTopic)
	}
	return err
}

// NewMqttSession starts a new session with the hub gateway
// This uses the client credentials, passed to mqttgw, as gateway credentials.
//
//	resolverClient for resolving capabilities
//	client is the mqttgw instance of the client connection
//
// Returns a session instance or an error if the gateway connection fails
func NewMqttSession(client *mqtt.Client) (session *MqttSession, err error) {

	// TODO: use client credentials
	//gwClient := resolver.GetCapability[gateway.IGatewaySession]()
	//if gwClient == nil {
	//	err = errors.New("gateway is not accessible")
	//	return nil, err
	//}
	clientID := string(client.Properties.Username)
	writer := NewMqttClientWriter(client)
	session = &MqttSession{
		mqttClient: client,
		gwClient:   nil, //gwClient,
		// FIXME: get the login ID
		m2dir:    NewMqtt2Directory(clientID, writer),
		m2hist:   NewMqtt2History(clientID, writer),
		m2pubsub: NewMqtt2PubSub(clientID, writer),
	}
	return session, err
}
