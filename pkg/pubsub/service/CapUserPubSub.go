package service

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/core"
)

// CapUserPubSub provides the capability to pub/sub for end-users
type CapUserPubSub struct {
	userID          string
	core            *core.PubSubCore
	subscriptionIDs []string
	subMutex        sync.RWMutex
}

// PubAction publishes an action by the user to a thing
func (cap *CapUserPubSub) PubAction(ctx context.Context, action *thing.ThingValue) (err error) {
	topic := MakeThingTopic(action.GatewayID, action.ThingID, pubsub.MessageTypeAction, action.Name)
	message, _ := json.Marshal(action)
	cap.core.Publish(topic, message)
	return
}

func (cap *CapUserPubSub) subMessage(gatewayID, thingID, msgType, name string,
	handler func(msgValue *thing.ThingValue)) error {

	subTopic := MakeThingTopic(gatewayID, thingID, msgType, name)

	subID, err := cap.core.Subscribe(subTopic,
		func(topic string, message []byte) {

			msgValue := &thing.ThingValue{}
			err := json.Unmarshal(message, msgValue)
			if err != nil {
				logrus.Error(err)
			}
			handler(msgValue)
		})

	// track the subscriptions to be able to unsubscribe
	if err == nil {
		cap.subMutex.Lock()
		cap.subscriptionIDs = append(cap.subscriptionIDs, subID)
		cap.subMutex.Unlock()
	}
	return err
}

// SubEvent creates a topic for receiving events
//
//	gatewayID publisher of the event. Use "" to subscribe to all publishers
//	thingID of the publisher event. Use "" to subscribe to events from all Things
//	eventName of the event. Use "" to subscribe to all events of publisher things
func (cap *CapUserPubSub) SubEvent(ctx context.Context,
	gatewayID, thingID, eventName string,
	handler func(event *thing.ThingValue)) (err error) {

	err = cap.subMessage(gatewayID, thingID, pubsub.MessageTypeEvent, eventName, handler)
	return err
}

// SubTDs subscribes to all (eligible) TDs from a gateway.
//
//	 gatewayID is optional. "" to subscribe to TD messages from all gateways.
//		handler is a callback invoked when a TD is received from a thing's publisher
func (cap *CapUserPubSub) SubTDs(_ context.Context, gatewayID string,
	handler func(event *thing.ThingValue)) (err error) {

	err = cap.subMessage(gatewayID, "", pubsub.MessageTypeTD, "+", handler)
	return
}

// Release the capability and end subscriptions
func (cap *CapUserPubSub) Release() {
	cap.subMutex.Lock()
	_ = cap.core.Unsubscribe(cap.subscriptionIDs)
	cap.subscriptionIDs = nil
	cap.subMutex.Unlock()
}

func NewCapUserPubSub(userID string, core *core.PubSubCore) *CapUserPubSub {
	cap := &CapUserPubSub{
		userID:          userID,
		core:            core,
		subscriptionIDs: make([]string, 0),
	}
	return cap
}
