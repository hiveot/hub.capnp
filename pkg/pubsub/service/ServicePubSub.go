package service

import (
	"context"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/core"
)

// ServicePubSub provides the capability to pub/sub for services
// This embeds the device and user pubsub capabilities
type ServicePubSub struct {
	DevicePubSub
	UserPubSub
	serviceID string
	core      *core.PubSubCore

	//subscriptionIDs []string
}

// SubActions extends the UserPubSub to subscribe to all actions aimed at things.
// Services can subscribe to other actions for logging, automation and other use-cases.
// For subscribing to service directed actions, use SubAction.
//
//	publisherID of the action target. Use "" to subscribe to all publishers
//	thingID of the action target. Use "" to subscribe to all Things
//	actionName or "" to subscribe to all actions
//	handler is a callback invoked when actions are received
func (cap *ServicePubSub) SubActions(
	ctx context.Context, thingAddr string, actionName string,
	handler func(action *thing.ThingValue)) (err error) {

	err = cap.UserPubSub.subMessage(thingAddr, pubsub.MessageTypeAction, actionName, handler)
	return
}

// Release the capability and end subscriptions
func (cap *ServicePubSub) Release() {
	cap.DevicePubSub.Release()
	cap.UserPubSub.Release()
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
		serviceID: serviceID,
		core:      core,
	}
	return servicePubSub
}
