package capnpclient

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/thing"
)

// UserPubSubCapnpClient is the capnp RPC client for user pubsub capabilities
// This implements the IUserPubSub interface
type UserPubSubCapnpClient struct {
	capability hubapi.CapUserPubSub
}

func (cl *UserPubSubCapnpClient) PubAction(
	ctx context.Context, publisherID, thingID, actionID string, value []byte) (err error) {

	method, release := cl.capability.PubAction(ctx,
		func(params hubapi.CapUserPubSub_pubAction_Params) error {
			_ = params.SetPublisherID(publisherID)
			_ = params.SetThingID(thingID)
			_ = params.SetActionID(actionID)
			err = params.SetValue(value)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// Release the capability and end subscriptions
func (cl *UserPubSubCapnpClient) Release() {
	cl.capability.Release()
}

func (cl *UserPubSubCapnpClient) SubEvent(
	ctx context.Context, publisherID, thingID string, eventID string,
	handler func(action *thing.ThingValue)) (err error) {

	//logrus.Infof("subscribing to event %s/%s/%s", publisherID, thingID, name)
	method, release := cl.capability.SubEvent(ctx,
		func(params hubapi.CapUserPubSub_subEvent_Params) error {
			_ = params.SetPublisherID(publisherID)
			_ = params.SetThingID(thingID)
			_ = params.SetEventID(eventID)
			handlerCapnp := NewSubscriptionHandlerCapnpServer(handler)
			err = params.SetHandler(handlerCapnp)
			return err
		})
	defer release()
	_, err = method.Struct()
	// logrus.Infof("subscribed to event %s/%s/%s. err=%v", publisherID, thingID, name, err)
	return err
}

// NewUserPubSubCapnpClient returns a capnp RPC client for the user pubsub capability
func NewUserPubSubCapnpClient(capability hubapi.CapUserPubSub) *UserPubSubCapnpClient {
	userCl := &UserPubSubCapnpClient{
		capability: capability,
	}
	return userCl
}
