package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub.go/pkg/vocab"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/core"
)

// CapDevicePubSub provides pub/sub capability to IoT devices.
// The IoT device is a gateway for the Things it manages, hence it has a gateway ID that is also
// its ThingID.
type CapDevicePubSub struct {
	// the gatewayID is the thingID of the IoT device itself
	gatewayID string
	// core is the pubsub engine
	core *core.PubSubCore
	// subscriptionIDs from the core
	subscriptionIDs []string
}

// PubEvent publishes the given thing event. The payload is an event value as per TD.
func (dps *CapDevicePubSub) PubEvent(_ context.Context, thingEvent *thing.ThingValue) (err error) {

	value, _ := json.Marshal(thingEvent)
	topic := MakeThingTopic(dps.gatewayID, thingEvent.ThingID, pubsub.MessageTypeEvent, thingEvent.Name)
	dps.core.Publish(topic, value)
	return
}

// PubProperties publishes an event with the given properties.
// The props is a map of property name-value pairs.
func (dps *CapDevicePubSub) PubProperties(
	_ context.Context, thingID string, props map[string]string) (err error) {

	propsValue, _ := json.Marshal(props)
	topic := MakeThingTopic(dps.gatewayID, thingID, pubsub.MessageTypeEvent, vocab.WoTProperties)
	thingEvent := thing.ThingValue{
		GatewayID: dps.gatewayID,
		ThingID:   thingID,
		Name:      vocab.WoTProperties,
		ValueJSON: propsValue,
		Created:   time.Now().Format(vocab.ISO8601Format),
	}
	eventValue, _ := json.Marshal(thingEvent)
	dps.core.Publish(topic, eventValue)
	return
}

// PubTD publishes the given thing TD as an event. The payload is a TD document.
// The event MUST be from the same device.
func (dps *CapDevicePubSub) PubTD(_ context.Context,
	thingID string, deviceType string, td []byte) (err error) {

	topic := MakeThingTopic(dps.gatewayID, thingID, pubsub.MessageTypeTD, deviceType)
	thingEvent := thing.ThingValue{
		GatewayID: dps.gatewayID,
		ThingID:   thingID,
		Name:      pubsub.MessageTypeTD,
		ValueJSON: td,
		Created:   time.Now().Format(vocab.ISO8601Format),
	}
	eventValue, _ := json.Marshal(thingEvent)
	dps.core.Publish(topic, eventValue)
	return
}

// SubAction subscribes to messages for the given thingID and action name
//
//	thingID and actionName are optional. Use "" to receive actions for all things or names.
func (dps *CapDevicePubSub) SubAction(
	_ context.Context, thingID string, actionName string,
	handler func(actionValue *thing.ThingValue)) (err error) {

	topic := MakeThingTopic(dps.gatewayID, thingID, pubsub.MessageTypeAction, actionName)
	subscriptionID, err := dps.core.Subscribe(topic,
		func(topic string, message []byte) {

			msgValue := &thing.ThingValue{}
			err = json.Unmarshal(message, msgValue)
			if err != nil {
				logrus.Error(err)
			}
			handler(msgValue)
		})
	if err == nil {
		dps.subscriptionIDs = append(dps.subscriptionIDs, subscriptionID)
	}
	return err
}

// Release the capability and end subscriptions
func (dps *CapDevicePubSub) Release() {
	err := dps.core.Unsubscribe(dps.subscriptionIDs)

	if err != nil {
		logrus.Errorf("IoT device %s unsubscribe failed: %s", dps.gatewayID, err)
	}
	dps.subscriptionIDs = nil
}

// NewCapDevicePubSub provides the capability for a device to publish actions and subscribe to events
//
//	gatewayID is the thingID of the IoT device doing the publishing
//	core is the core pubsub that is used for publishing and subscribing
func NewCapDevicePubSub(gatewayID string, core *core.PubSubCore) *CapDevicePubSub {
	cap := &CapDevicePubSub{
		gatewayID:       gatewayID,
		core:            core,
		subscriptionIDs: make([]string, 0),
	}
	return cap
}
