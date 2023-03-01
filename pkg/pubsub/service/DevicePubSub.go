package service

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/vocab"
	"github.com/hiveot/hub/lib/caphelp"

	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/core"
)

// DevicePubSub provides pub/sub capability to IoT devices.
// The IoT device is a gateway for the Things it manages, hence it has a gateway ID that is also
// its ThingID.
type DevicePubSub struct {
	// the publisherID is the thingID of the IoT device or service
	publisherID string
	// core is the pubsub engine
	core *core.PubSubCore
	// subscriptionIDs from the core
	subscriptionIDs []string
}

// PubEvent publishes the given thing event. The payload is an event value as per TD.
func (dps *DevicePubSub) PubEvent(
	_ context.Context, thingID, eventID string, value []byte) (err error) {

	logrus.Infof("publisherID=%s, thingID=%s, name=%s", dps.publisherID, thingID, eventID)

	tv := thing.NewThingValue(dps.publisherID, thingID, eventID, caphelp.Clone(value))
	// note that marshal will copy the value so its buffer can be reused by capnp
	tvSerialized, _ := json.Marshal(tv)
	topic := MakeThingTopic(dps.publisherID, thingID, pubsub.MessageTypeEvent, eventID)
	go dps.core.Publish(topic, tvSerialized)
	return
}

// PubProperties publishes an event with the given properties.
// The props is a map of property name-value pairs.
func (dps *DevicePubSub) PubProperties(
	_ context.Context, thingID string, props map[string][]byte) (err error) {

	logrus.Infof("publisherID=%s, thingID=%s, nrProps=%d", dps.publisherID, thingID, len(props))

	propsValue, _ := json.Marshal(props)
	tv := thing.NewThingValue(dps.publisherID, thingID, vocab.WoTProperties, propsValue)

	// note that marshal will copy the props map so its buffer can be reused by capnp
	tvSerialized, _ := json.Marshal(tv)
	topic := MakeThingTopic(dps.publisherID, thingID, pubsub.MessageTypeEvent, vocab.WoTProperties)
	dps.core.Publish(topic, tvSerialized)
	return
}

// PubTD publishes the given thing TD as an event. The payload is a TD document.
func (dps *DevicePubSub) PubTD(_ context.Context, thingID string, td []byte) (err error) {

	logrus.Infof("publisherID=%s, thingID=%s", dps.publisherID, thingID)

	tv := thing.NewThingValue(
		dps.publisherID, thingID, pubsub.MessageTypeTD, td)

	// note that marshal will copy the TD so its buffer can be reused by capnp
	topic := MakeThingTopic(dps.publisherID, thingID, pubsub.MessageTypeTD, "")

	tvSerialized, _ := json.Marshal(tv)
	dps.core.Publish(topic, tvSerialized)
	return
}

// SubAction subscribes to messages for the given thingID and action name
//
//	thingID and actionID are optional. Use "" to receive actions for all things or names.
func (dps *DevicePubSub) SubAction(
	_ context.Context, thingID string, actionID string,
	handler func(actionValue *thing.ThingValue)) (err error) {

	logrus.Infof("publisherID=%s, thingID=%s, actionName=%s", dps.publisherID, thingID, actionID)

	topic := MakeThingTopic(dps.publisherID, thingID, pubsub.MessageTypeAction, actionID)
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
func (dps *DevicePubSub) Release() {
	err := dps.core.Unsubscribe(dps.subscriptionIDs)

	if err != nil {
		logrus.Errorf("IoT device %s unsubscribe failed: %s", dps.publisherID, err)
	}
	dps.subscriptionIDs = nil
}

// NewDevicePubSub provides the capability for a device to publish actions and subscribe to events
//
//	publisherID is the thingID of the IoT device doing the publishing
//	core is the core pubsub that is used for publishing and subscribing
func NewDevicePubSub(gatewayID string, core *core.PubSubCore) *DevicePubSub {
	deviceCap := &DevicePubSub{
		publisherID:     gatewayID,
		core:            core,
		subscriptionIDs: make([]string, 0),
	}
	return deviceCap
}
