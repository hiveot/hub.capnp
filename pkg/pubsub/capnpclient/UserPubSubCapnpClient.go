package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
)

// UserPubSubCapnpClient is the capnp RPC client for user pubsub capabilities
// This implements the IUserPubSub interface
type UserPubSubCapnpClient struct {
	capability hubapi.CapUserPubSub
}

func (cl *UserPubSubCapnpClient) PubAction(
	ctx context.Context, thingAddr, name string, value []byte) (err error) {

	method, release := cl.capability.PubAction(ctx,
		func(params hubapi.CapUserPubSub_pubAction_Params) error {
			_ = params.SetThingAddr(thingAddr)
			_ = params.SetActionName(name)
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

func (cl *UserPubSubCapnpClient) SubEvent(ctx context.Context, thingAddr string, name string,
	handler func(action *thing.ThingValue)) (err error) {

	method, release := cl.capability.SubEvent(ctx,
		func(params hubapi.CapUserPubSub_subEvent_Params) error {
			_ = params.SetThingAddr(thingAddr)
			_ = params.SetEventName(name)
			handlerCapnp := NewSubscriptionHandlerCapnpServer(handler)
			err = params.SetHandler(handlerCapnp)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

func (cl *UserPubSubCapnpClient) SubTDs(ctx context.Context,
	handler func(action *thing.ThingValue)) (err error) {

	method, release := cl.capability.SubTDs(ctx,
		func(params hubapi.CapUserPubSub_subTDs_Params) error {
			handlerCapnp := NewSubscriptionHandlerCapnpServer(handler)
			err = params.SetHandler(handlerCapnp)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// NewUserPubSubCapnpClient returns a capnp RPC client for the user pubsub capability
func NewUserPubSubCapnpClient(capability hubapi.CapUserPubSub) *UserPubSubCapnpClient {
	userCl := &UserPubSubCapnpClient{
		capability: capability,
	}
	return userCl
}
