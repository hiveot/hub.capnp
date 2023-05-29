package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/resolver"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/mqttgw/mqttclient"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/sirupsen/logrus"
)

// Mqtt2PubSub handles Hub pubsub requests over MQTT
type Mqtt2PubSub struct {
	// the Hub user or device ID that
	clientID string

	// pubsub capabilities is loaded on first use
	userPubSub   pubsub.IUserPubSub
	devicePubSub pubsub.IDevicePubSub
	writer       *MqttClientWriter
}

// Return the user pubsub capability. Obtain it from the resolver on first use.
func (m2pubsub *Mqtt2PubSub) getUserPubSub() pubsub.IUserPubSub {
	if m2pubsub.userPubSub == nil {
		m2pubsub.userPubSub = resolver.GetCapability[pubsub.IUserPubSub]()
	}
	return m2pubsub.userPubSub
}

// Return the device pubsub capability. Obtain it from the resolver on first use.
func (m2pubsub *Mqtt2PubSub) getDevicePubSub() pubsub.IDevicePubSub {
	if m2pubsub.devicePubSub == nil {
		m2pubsub.devicePubSub = resolver.GetCapability[pubsub.IDevicePubSub]()
	}
	return m2pubsub.devicePubSub
}

// Release the pubsub session
func (m2pubsub *Mqtt2PubSub) Release() {
	if m2pubsub.devicePubSub != nil {
		m2pubsub.devicePubSub.Release()
	}
	if m2pubsub.userPubSub != nil {
		m2pubsub.userPubSub.Release()
	}
}

// HandlePublish handles the request to publish a message to the Hub pubsub
//
// The following mqttgw topics are mapped to Hub pubsub:
//
//	things/{publisherID}/{thingID}/event/{name}  -> DevicePubSub.PubEvent
//	things/{publisherID}/{thingID}/td            -> DevicePubSub.PubTD
//	things/{publisherID}/{thingID}/action/{name} -> UserPubSub.PubAction
//
// * where msgType is one of 'event', 'action', 'td'
// * where name is the name of the event, action or the thing devicetype
//
// This returns an error if the client is not authorized
func (m2pubsub *Mqtt2PubSub) HandlePublish(mqttTopic string, payload []byte) (err error) {
	// only handle things topics
	// first time obtain the publish capability
	if !mqttclient.IsThingsTopic(mqttTopic) {
		return nil
	}
	pubID, thingID, msgType, name, err := mqttclient.SplitThingsTopic(mqttTopic)
	if err != nil {
		return fmt.Errorf("invalid mqttgw topic: %w", err)
	}
	if msgType == mqttclient.MessageTypeEvent { // device api
		// events must come from the publisher
		if pubID != m2pubsub.clientID {
			err = fmt.Errorf("event publisher '%s' doesn't match client ID '%s'", pubID, m2pubsub.clientID)
		} else {
			err = m2pubsub.getDevicePubSub().PubEvent(context.Background(), thingID, name, payload)
		}
	} else if msgType == mqttclient.MessageTypeAction { // user api
		err = m2pubsub.getUserPubSub().PubAction(context.Background(), pubID, thingID, name, payload)
	}
	return err
}

// HandleSubscribe is invoked when the MQTT client requests subscription on a topic.
//
// Thing subscriptions on topic things/{publisherID}/{thingID}/{msgType}/{name} are
// passed on to the pubsub service if they pass the authorization check.
//
// Other subscriptions are ignored and will be handled by the mqttgw broker as normal.
// This returns an error if the client is unauthorized.
func (m2pubsub *Mqtt2PubSub) HandleSubscribe(mqttTopic string, payload []byte) error {

	logrus.Infof("OnSubscribe to '%s' by client %s", mqttTopic, m2pubsub.clientID)

	// ignore if the topic isn't for 'things'
	if !mqttclient.IsThingsTopic(mqttTopic) {
		return nil
	}
	pubID, thingID, msgType, name, err := mqttclient.SplitThingsTopic(mqttTopic)
	if err != nil {
		return fmt.Errorf("invalid Things mqttTopic '%s' by client '%s': %w",
			mqttTopic, m2pubsub.clientID, err)
	}

	// TODO: authorization

	// pass the subscription to the pubsub service and the resulting subscription messages to the mqttgw client.
	if msgType == hubapi.MessageTypeEvent {
		err = m2pubsub.getUserPubSub().SubEvent(context.Background(), pubID, thingID, name,
			func(event thing.ThingValue) {
				mqttTopic = mqttclient.MakeEventTopic(pubID, thingID, name)
				evJson, _ := json.Marshal(event)
				err = m2pubsub.writer.Write(mqttTopic, evJson)
				if err != nil {
					logrus.Errorf("Failed to publish received event to mqttgw bus on topic '%s': %s", mqttTopic, err)
				}
			})
		return err
	} else if msgType == "action" {
		if pubID != m2pubsub.clientID {
			return fmt.Errorf("subscribe to action by '%s' from different publisher '%s'", m2pubsub.clientID, pubID)
		}
		err = m2pubsub.getDevicePubSub().SubAction(context.Background(), thingID, name,
			func(thingAction thing.ThingValue) {
				mqttTopic = mqttclient.MakeActionTopic(pubID, thingID, name)
				actionJson, _ := json.Marshal(thingAction)
				err = m2pubsub.writer.Write(mqttTopic, actionJson)
				if err != nil {
					logrus.Errorf("Failed to publish received action to mqttgw bus on topic '%s': %s", mqttTopic, err)
				}
			})
		return err
	}

	return err
}

// NewMqtt2PubSub starts a new session with the hub gateway
// This uses the client credentials, passed to mqttgw, as gateway credentials.
//
//	resolverClient for resolving capabilities
//	caCert is optional to ensure a valid connection to the gateway
//	client is the mqttgw instance of the client connection
//
// Returns a session instance or an error if the gateway connection fails
func NewMqtt2PubSub(clientID string, writer *MqttClientWriter) *Mqtt2PubSub {
	pubsub := &Mqtt2PubSub{
		clientID: clientID,
		writer:   writer,
	}
	return pubsub
}
