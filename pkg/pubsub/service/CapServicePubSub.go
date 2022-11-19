package service

import (
	"context"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/core"
)

// CapServicePubSub provides the capability to pub/sub for services
// This embeds the device and user pubsub capabilities
type CapServicePubSub struct {
	CapDevicePubSub
	CapUserPubSub
	serviceID string
	core      *core.PubSubCore

	//subscriptionIDs []string
}

// SubActions extends the UserPubSub to subscribe to all actions aimed at things.
// Services can subscribe to other actions for logging, automation and other use-cases.
// For subscribing to service directed actions, use SubAction.
//
//	gatewayID of the action target. Use "" to subscribe to all publishers
//	thingID of the action target. Use "" to subscribe to all Things
//	actionName or "" to subscribe to all actions
//	handler is a callback invoked when actions are received
func (cap *CapServicePubSub) SubActions(ctx context.Context,
	gatewayID string, thingID string, actionName string,
	handler func(action *thing.ThingValue)) (err error) {

	err = cap.CapUserPubSub.subMessage(gatewayID, thingID, pubsub.MessageTypeAction, actionName, handler)
	return
}

// Release the capability and end subscriptions
func (cap *CapServicePubSub) Release() {
	cap.CapDevicePubSub.Release()
	cap.CapUserPubSub.Release()
}

func NewCapServicePubSub(serviceID string, core *core.PubSubCore) *CapServicePubSub {
	cap := &CapServicePubSub{
		CapUserPubSub: CapUserPubSub{
			userID:          serviceID,
			core:            core,
			subscriptionIDs: make([]string, 0),
		},
		CapDevicePubSub: CapDevicePubSub{
			gatewayID:       serviceID,
			core:            core,
			subscriptionIDs: make([]string, 0),
		},
		serviceID: serviceID,
		core:      core,
	}
	return cap
}
