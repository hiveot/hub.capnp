package service

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hiveot/hub/lib/resolver"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/mqtt/mqttclient"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/mochi-co/mqtt/v2"
)

// MqttSession manages a MQTT client session with the HiveOT gateway
// It is created by the mochi hook on a new incoming connection.
// This session establishes a gateway session on startup and releases it on disconnect.
// This uses the client resolver to obtain capabilities, which also aids in testing using stubs.
type MqttSession struct {
	mqttClient *mqtt.Client
	//gwCapClient  capnp.Client
	gwClient     gateway.IGatewaySession
	refreshToken string
	// login ID of this client
	clientID string

	// user authentication
	userAuthn authn.IUserAuthn
	// pubsub capabilities is loaded on first use
	userPubSub   pubsub.IUserPubSub
	devicePubSub pubsub.IDevicePubSub

	// directory capabilities
	readDir directory.IReadDirectory

	// history capabilities
	readHist history.IReadHistory
}

// Return the user pubsub capability. Obtain it from the resolver on first use.
func (session *MqttSession) getUserPubSub() pubsub.IUserPubSub {
	if session.userPubSub == nil {
		session.userPubSub = resolver.GetCapability[pubsub.IUserPubSub]()
	}
	return session.userPubSub
}

// Return the device pubsub capability. Obtain it from the resolver on first use.
func (session *MqttSession) getDevicePubSub() pubsub.IDevicePubSub {
	if session.devicePubSub == nil {
		session.devicePubSub = resolver.GetCapability[pubsub.IDevicePubSub]()
	}
	return session.devicePubSub
}

// Return the read directory capability. Obtain it from the resolver on first use.
func (session *MqttSession) getReadDirectory() directory.IReadDirectory {
	if session.readDir == nil {
		session.readDir = resolver.GetCapability[directory.IReadDirectory]()
	}
	return session.readDir
}

// Return the device pubsub capability. Obtain it from the resolver on first use.
func (session *MqttSession) getReadHistory() history.IReadHistory {
	if session.readHist == nil {
		session.readHist = resolver.GetCapability[history.IReadHistory]()
	}
	return session.readHist
}

// OnDisconnect release the gateway session on a disconnect
func (session *MqttSession) OnDisconnect() {
}

// Login to the resolver session, most likely the gateway
// This requires that the resolver client is connected to the resolver service.
func (session *MqttSession) Login(loginID, password string) error {
	session.clientID = loginID
	err := resolver.Login(loginID, password)
	return err
}

// OnSubscribe is invoked when the MQTT client requests subscription on a topic.
// This is passed on to the pubsub service.
func (session *MqttSession) OnSubscribe(topic string, cb func(thing.ThingValue)) error {
	pubID, thingID, msgType, name, err := mqttclient.SplitTopic(topic)
	if err != nil {
		return fmt.Errorf("OnSubscribe: %w", err)
	}
	// TBD: authorization?
	if msgType == "event" {
		err = session.getUserPubSub().SubEvent(context.Background(), pubID, thingID, name, cb)
	} else if msgType == "action" {
		if pubID != session.clientID {
			return fmt.Errorf("subscribe to action by '%s' from different publiser '%s'", session.clientID, pubID)
		}
		err = session.getDevicePubSub().SubAction(context.Background(), thingID, name, cb)
	}
	return nil
}

// OnPublish handles a publish request.
//
//	This proxies the publication to the Hub's pubsub service.
//
// The publisher must be logged in and have permission to publishing
// The topic format is: things/{publisherID}/{thingID}/{msgType}/name
// * where msgType is one of 'event', 'action', 'td'
// * where name is the name of the event, action or the thing devicetype
func (session *MqttSession) OnPublish(topic string, payload []byte) (err error) {
	// first time obtain the publish capability
	pubID, thingID, msgType, name, err := mqttclient.SplitTopic(topic)
	if err != nil {
		return fmt.Errorf("OnPublish error: %w", err)
	}

	if msgType == "event" { // device api
		// events must come from the publisher
		if pubID != session.clientID {
			err = fmt.Errorf("event publisher '%s' doesn't match client ID '%s'", pubID, session.clientID)
		} else {
			err = session.getDevicePubSub().PubEvent(context.Background(), thingID, name, payload)
		}
	} else if msgType == "action" { // user api
		err = session.getUserPubSub().PubAction(context.Background(), pubID, thingID, name, payload)
	} else if msgType == "td" { // device api
		// TDs must come from the publisher
		if pubID != session.clientID {
			err = errors.New(fmt.Sprintf("TD publisher '%s' doesn't match client ID '%s'", pubID, session.clientID))
		} else {
			// TD's use the event name 'td'
			err = session.getDevicePubSub().PubEvent(context.Background(), thingID, msgType, payload)
		}
	}
	return err
}

// NewMqttSession starts a new session with the hub gateway
// This uses the client credentials, passed to mqtt, as gateway credentials.
//
//	resolverClient for resolving capabilities
//	caCert is optional to ensure a valid connection to the gateway
//	client is the mqtt instance of the client connection
//
// Returns a session instance or an error if the gateway connection fails
func NewMqttSession(caCert *x509.Certificate, client *mqtt.Client) (session *MqttSession, err error) {

	// TODO: use client credentials
	//gwClient := resolver.GetCapability[gateway.IGatewaySession]()
	//if gwClient == nil {
	//	err = errors.New("gateway is not accessible")
	//	return nil, err
	//}
	session = &MqttSession{
		mqttClient: client,
		gwClient:   nil, //gwClient,
	}
	return session, err
}
