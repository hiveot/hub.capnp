package service

import (
	"context"
	"encoding/json"
	"github.com/hiveot/hub/api/go/hubapi"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/vocab"
	"github.com/hiveot/hub/lib/caphelp"

	"github.com/hiveot/hub/lib/thing"
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
	// subscriptionIDs from the core to be released with the capability
	subscriptionIDs []string
}

// PubEvent publishes the given thing event. The payload is an event value as per TD.
func (svc *DevicePubSub) PubEvent(
	_ context.Context, thingID, eventID string, value []byte) (err error) {

	logrus.Infof("publisherID=%s, thingID=%s, name=%s", svc.publisherID, thingID, eventID)

	tv := thing.NewThingValue(svc.publisherID, thingID, eventID, caphelp.Clone(value))
	// note that marshal will copy the value so its buffer can be reused by capnp
	tvSerialized, _ := json.Marshal(tv)
	topic := MakeThingTopic(svc.publisherID, thingID, hubapi.MessageTypeEvent, eventID)
	go svc.core.Publish(topic, tvSerialized)
	return
}

// SubAction subscribes to messages for the given thingID and action name
//
//	thingID and actionID are optional. Use "" to receive actions for all things or names.
func (svc *DevicePubSub) SubAction(
	_ context.Context, thingID string, actionID string,
	handler func(actionValue *thing.ThingValue)) (err error) {

	logrus.Infof("publisherID=%s, thingID=%s, actionName=%s",
		svc.publisherID, thingID, actionID)

	topic := MakeThingTopic(svc.publisherID, thingID, hubapi.MessageTypeAction, actionID)
	subscriptionID, err := svc.core.Subscribe(topic,
		func(topic string, message []byte) {
			msgValue := &thing.ThingValue{}
			err = json.Unmarshal(message, msgValue)
			if err != nil {
				logrus.Error(err)
			}
			// Do not pass properties configuration action messages
			if msgValue.ID != vocab.WoTProperties {
				handler(msgValue)
			}
		})
	if err == nil {
		svc.subscriptionIDs = append(svc.subscriptionIDs, subscriptionID)
	}
	return err
}

// Release the capability and end subscriptions
func (svc *DevicePubSub) Release() {
	err := svc.core.Unsubscribe(svc.subscriptionIDs)

	if err != nil {
		logrus.Errorf("IoT device %s unsubscribe failed: %s", svc.publisherID, err)
	}
	svc.subscriptionIDs = nil
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
