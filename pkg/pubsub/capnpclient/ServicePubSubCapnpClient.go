package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
)

// ServicePubSubCapnpClient is the capnp RPC client for service pubsub capabilities
// This implements the IServicePubSub interface
type ServicePubSubCapnpClient struct {
	UserPubSubCapnpClient
	DevicePubSubCapnpClient
	capability hubapi.CapServicePubSub
}

func (cl *ServicePubSubCapnpClient) SubActions(ctx context.Context, thingAddr string, name string,
	handler func(action *thing.ThingValue)) (err error) {

	method, release := cl.capability.SubActions(ctx,
		func(params hubapi.CapServicePubSub_subActions_Params) error {
			_ = params.SetThingAddr(thingAddr)
			_ = params.SetActionName(name)
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
