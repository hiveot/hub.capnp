package capnpclient

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/thing"
)

// ServicePubSubCapnpClient is the capnp RPC client for service pubsub capabilities
// This implements the IServicePubSub interface
type ServicePubSubCapnpClient struct {
	UserPubSubCapnpClient
	DevicePubSubCapnpClient
	capability hubapi.CapServicePubSub
}

func (cl *ServicePubSubCapnpClient) SubActions(
	ctx context.Context, publisherID, thingID, actionID string,
	handler func(thing.ThingValue)) (err error) {

	method, release := cl.capability.SubActions(ctx,
		func(params hubapi.CapServicePubSub_subActions_Params) error {
			_ = params.SetPublisherID(publisherID)
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

func (cl *ServicePubSubCapnpClient) SubEvents(
	ctx context.Context, publisherID, thingID, eventID string,
	handler func(thing.ThingValue)) (err error) {

	method, release := cl.capability.SubEvents(ctx,
		func(params hubapi.CapServicePubSub_subEvents_Params) error {
			_ = params.SetPublisherID(publisherID)
			_ = params.SetThingID(thingID)
			_ = params.SetEventID(eventID)
			handlerCapnp := NewSubscriptionHandlerCapnpServer(handler)
			err = params.SetHandler(handlerCapnp)
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

func (cl *ServicePubSubCapnpClient) Release() {
	cl.UserPubSubCapnpClient.Release()
	cl.DevicePubSubCapnpClient.Release()
}

// NewServicePubSubCapnpClient returns a capnp RPC client for the service pubsub capability
func NewServicePubSubCapnpClient(capability hubapi.CapServicePubSub) *ServicePubSubCapnpClient {
	serviceCl := &ServicePubSubCapnpClient{
		DevicePubSubCapnpClient: *NewDevicePubSubCapnpClient(hubapi.CapDevicePubSub(capability)),
		UserPubSubCapnpClient:   *NewUserPubSubCapnpClient(hubapi.CapUserPubSub(capability)),
		capability:              capability,
	}
	return serviceCl
}
