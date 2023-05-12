package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hiveot/hub/api/go/hubapi"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/lib/thing"
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
	_ context.Context, publisherID, thingID, actionID string, value []byte) (err error) {

	logrus.Infof("userID=%s, thingID=%s, actionName=%s", cap.userID, thingID, actionID)

	topic := MakeThingTopic(publisherID, thingID, hubapi.MessageTypeAction, actionID)
	tv := thing.NewThingValue(publisherID, thingID, actionID, value)
	// note that marshal will copy the values so changes to the buffer containing value will not affect it
	message, _ := json.Marshal(tv)
	cap.core.Publish(topic, message)
	return
}

// SubEvent creates a topic for receiving events.
// Either a thingID or eventID must be provided.
//
//	publisherID publisher of the event. Use "" to subscribe to all publishers
//	thingID of the publisher event. Use "" to subscribe to events from all Things
//	eventID of the event. Use "" to subscribe to all events of a thing
func (cap *UserPubSub) SubEvent(_ context.Context, publisherID, thingID, eventID string,
	handler func(thing.ThingValue)) error {

	// it is not allowed to subscribe to all events of all things. Pick one or the other.
	if thingID == "" && eventID == "" {
		return fmt.Errorf("a thingID or eventID must be provided")
	}

	//logrus.Infof("userID=%s, thingID=%s, eventID=%s", cap.userID, thingID, eventID)
	subTopic := MakeThingTopic(publisherID, thingID, hubapi.MessageTypeEvent, eventID)

	subID, err := cap.core.Subscribe(subTopic,
		func(topic string, message []byte) {
			msgValue := thing.ThingValue{}
			err := json.Unmarshal(message, &msgValue)
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
