package capnpclient

import (
	"context"

	"github.com/hiveot/hub/lib/caphelp"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/thing"
)

// DevicePubSubCapnpClient is the capnp RPC client for device pubsub capabilities
// This implements the IDevicePubSub interface
type DevicePubSubCapnpClient struct {
	capability hubapi.CapDevicePubSub
}

func (cl *DevicePubSubCapnpClient) PubEvent(
	ctx context.Context, thingID, name string, value []byte) (err error) {

	method, release := cl.capability.PubEvent(ctx,
		func(params hubapi.CapDevicePubSub_pubEvent_Params) error {
			_ = params.SetThingID(thingID)
			_ = params.SetName(name)
			err = params.SetValue(value)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// PubProperties publishes properties of a thing.
func (cl *DevicePubSubCapnpClient) PubProperties(ctx context.Context, thingID string, props map[string][]byte) (err error) {

	method, release := cl.capability.PubProperties(ctx,
		func(params hubapi.CapDevicePubSub_pubProperties_Params) error {
			_ = params.SetThingID(thingID)
			propsCapnp := caphelp.MarshalKeyValueMap(props)
			err = params.SetProps(propsCapnp)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// PubTD publishes the given thing TD. The payload is a serialized TD document.
func (cl *DevicePubSubCapnpClient) PubTD(
	ctx context.Context, thingID string, deviceType string, tdDoc []byte) (err error) {

	method, release := cl.capability.PubTD(ctx,
		func(params hubapi.CapDevicePubSub_pubTD_Params) error {
			_ = params.SetThingID(thingID)
			_ = params.SetDeviceType(deviceType)
			err = params.SetTdDoc(tdDoc)
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

// SubAction creates a topic and registers a listener for actions to things with this gateway.
// This supports receiving queued messages for this gateway since it last disconnected.
//
//	thingID is the thing to subscribe for, or "" to subscribe to all things of this gateway
//	name is the action name, or "" to subscribe to all actions
//	handler will be invoked when an action is received for this device
func (cl *DevicePubSubCapnpClient) SubAction(
	ctx context.Context, thingID string, name string,
	handler func(action *thing.ThingValue)) (err error) {

	method, release := cl.capability.SubAction(ctx,
		func(params hubapi.CapDevicePubSub_subAction_Params) error {
			_ = params.SetThingID(thingID)
			_ = params.SetName(name)
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
