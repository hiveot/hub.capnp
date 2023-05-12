package service

import (
	"context"
	"encoding/json"
	"github.com/hiveot/hub/api/go/hubapi"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/pubsub/core"
)

// ServicePubSub provides the capability to pub/sub for services
// This embeds the device and user pubsub capabilities
type ServicePubSub struct {
	DevicePubSub
	UserPubSub
	serviceID       string
	core            *core.PubSubCore
	subscriptionIDs []string
}

// SubActions subscribe to all actions aimed at things
// Services can subscribe to other actions for logging, automation and other use-cases.
// For subscribing to service directed actions, use SubAction.
//
//	publisherID of the action target. Use "" to subscribe to all publishers
//	thingID of the action target. Use "" to subscribe to all Things
//	actionID or "" to subscribe to all actions
//	handler is a callback invoked when actions are received
func (svc *ServicePubSub) SubActions(
	_ context.Context, publisherID, thingID, actionID string,
	handler func(thing.ThingValue)) (err error) {

	logrus.Infof("publisherID=%s, thingID=%s, actionID=%s", publisherID, thingID, actionID)
	subTopic := MakeThingTopic(publisherID, thingID, hubapi.MessageTypeAction, actionID)
	subID, err := svc.core.Subscribe(subTopic,
		func(topic string, message []byte) {
			// FIXME: capnp serialization of messageValue?
			msgValue := thing.ThingValue{}
			err := json.Unmarshal(message, &msgValue)
			if err != nil {
				logrus.Error(err)
			}
			handler(msgValue)
		})
	if err == nil {
		svc.subscriptionIDs = append(svc.subscriptionIDs, subID)
	}
	return err
}

// SubEvents subscribes to events aimed at things from any publisher
func (svc *ServicePubSub) SubEvents(
	_ context.Context, publisherID, thingID, eventID string,
	handler func(thing.ThingValue)) (err error) {

	logrus.Infof("publisherID=%s, thingID=%s, eventID=%s", publisherID, thingID, eventID)
	subTopic := MakeThingTopic(publisherID, thingID, hubapi.MessageTypeEvent, eventID)
	subID, err := svc.core.Subscribe(subTopic,
		func(topic string, message []byte) {
			msgValue := thing.ThingValue{}
			err := json.Unmarshal(message, &msgValue)
			if err != nil {
				logrus.Error(err)
			}
			handler(msgValue)
		})
	if err == nil {
		svc.subscriptionIDs = append(svc.subscriptionIDs, subID)
	}
	return err
}

// Release the capability and end subscriptions
func (svc *ServicePubSub) Release() {
	_ = svc.core.Unsubscribe(svc.subscriptionIDs)
	svc.subscriptionIDs = nil
	svc.DevicePubSub.Release()
	svc.UserPubSub.Release()
}

func NewServicePubSub(serviceID string, core *core.PubSubCore) *ServicePubSub {
	servicePubSub := &ServicePubSub{
		UserPubSub: UserPubSub{
			userID:          serviceID,
			core:            core,
			subscriptionIDs: make([]string, 0),
		},
		DevicePubSub: DevicePubSub{
			publisherID:     serviceID,
			core:            core,
			subscriptionIDs: make([]string, 0),
		},
		serviceID:       serviceID,
		core:            core,
		subscriptionIDs: make([]string, 0),
	}
	return servicePubSub
}
