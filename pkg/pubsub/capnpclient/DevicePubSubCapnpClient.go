package capnpclient

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/thing"
)

// DevicePubSubCapnpClient is the capnp RPC client for device pubsub capabilities
// This implements the IDevicePubSub interface
type DevicePubSubCapnpClient struct {
	capability hubapi.CapDevicePubSub
}

func (cl *DevicePubSubCapnpClient) PubEvent(
	ctx context.Context, thingID, eventID string, value []byte) (err error) {

	method, release := cl.capability.PubEvent(ctx,
		func(params hubapi.CapDevicePubSub_pubEvent_Params) error {
			_ = params.SetThingID(thingID)
			_ = params.SetEventID(eventID)
			err = params.SetValue(value)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// Release the capability and end subscriptions
func (cl *DevicePubSubCapnpClient) Release() {
	cl.capability.Release()
}

// SubAction creates registers a callback for action requests made to things managed by this device.
//
//	thingID is the thing to subscribe for, or "" to subscribe to all things of this device
//	actionID is the action ID, or "" to subscribe to all actions
//	handler will be invoked when an action is received for this device
func (cl *DevicePubSubCapnpClient) SubAction(
	ctx context.Context, thingID string, actionID string,
	handler func(action *thing.ThingValue)) (err error) {

	method, release := cl.capability.SubAction(ctx,
		func(params hubapi.CapDevicePubSub_subAction_Params) error {
			_ = params.SetThingID(thingID)
			_ = params.SetActionID(actionID)
			handlerCapnp := NewSubscriptionHandlerCapnpServer(handler)
			err = params.SetHandler(handlerCapnp)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// NewDevicePubSubCapnpClient returns a capnp RPC client for the device pubsub capability
func NewDevicePubSubCapnpClient(capability hubapi.CapDevicePubSub) *DevicePubSubCapnpClient {
	deviceCl := &DevicePubSubCapnpClient{
		capability: capability,
	}
	return deviceCl
}
