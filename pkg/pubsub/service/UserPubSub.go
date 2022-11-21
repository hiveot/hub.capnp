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

// UserPubSub provides the capability to pub/sub for end-users
type UserPubSub struct {
	userID          string
	core            *core.PubSubCore
	subscriptionIDs []string
	subMutex        sync.RWMutex
}

// PubAction publishes an action by the user to a thing
func (cap *UserPubSub) PubAction(
	_ context.Context, thingAddr, actionName string, value []byte) (err error) {

	topic := MakeThingTopic(thingAddr, pubsub.MessageTypeAction, actionName)
	tv := thing.NewThingValue(thingAddr, actionName, value)
	message, _ := json.Marshal(tv)
	cap.core.Publish(topic, message)
	return
}

func (cap *UserPubSub) subMessage(thingAddr, msgType, name string,
	handler func(msgValue *thing.ThingValue)) error {

	subTopic := MakeThingTopic(thingAddr, msgType, name)

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
//	publisherID publisher of the event. Use "" to subscribe to all publishers
//	thingID of the publisher event. Use "" to subscribe to events from all Things
//	eventName of the event. Use "" to subscribe to all events of publisher things
func (cap *UserPubSub) SubEvent(ctx context.Context, thingAddr, eventName string,
	handler func(event *thing.ThingValue)) (err error) {

	err = cap.subMessage(thingAddr, pubsub.MessageTypeEvent, eventName, handler)
	return err
}

// SubTDs subscribes to all (eligible) TDs from a gateway.
//
//	 publisherID is optional. "" to subscribe to TD messages from all gateways.
//		handler is a callback invoked when a TD is received from a thing's publisher
func (cap *UserPubSub) SubTDs(_ context.Context, handler func(event *thing.ThingValue)) (err error) {

	err = cap.subMessage("", pubsub.MessageTypeTD, "+", handler)
	return
}

// Release the capability and end subscriptions
func (cap *UserPubSub) Release() {
	cap.subMutex.Lock()
	_ = cap.core.Unsubscribe(cap.subscriptionIDs)
	cap.subscriptionIDs = nil
	cap.subMutex.Unlock()
}

func NewUserPubSub(userID string, core *core.PubSubCore) *UserPubSub {
	userPubSub := &UserPubSub{
		userID:          userID,
		core:            core,
		subscriptionIDs: make([]string, 0),
	}
	return userPubSub
}
