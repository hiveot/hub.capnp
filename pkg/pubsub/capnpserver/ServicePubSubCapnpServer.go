package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/pubsub"
)

// ServicePubSubCapnpServer provides the capnp RPC server for service's pubsub.
// This implements the capnproto generated interface CapServicePubSub_Server
type ServicePubSubCapnpServer struct {
	DevicePubSubCapnpServer
	UserPubSubCapnpServer
	svc pubsub.IServicePubSub
}

func (capsrv *ServicePubSubCapnpServer) SubActions(
	ctx context.Context, call hubapi.CapServicePubSub_subActions) error {
	args := call.Args()
	thingAddr, _ := args.ThingAddr()
	name, _ := args.ActionName()
	handlerCap := args.Handler()
	handlerClient := NewSubscriptionHandlerCapnpClient(handlerCap.AddRef())
	// the server registers the callback handler and invokes it when actions for the Thing are received
	err := capsrv.svc.SubActions(ctx, thingAddr, name, handlerClient.HandleValue)
	return err
}

func (capsrv *ServicePubSubCapnpServer) Shutdown() {
	// Client is released, release the subscriptions
	capsrv.svc.Release()
}

func NewServicePubSubCapnpServer(svc pubsub.IServicePubSub) *ServicePubSubCapnpServer {
	capsrv := &ServicePubSubCapnpServer{
		svc:                     svc,
		DevicePubSubCapnpServer: DevicePubSubCapnpServer{svc},
		UserPubSubCapnpServer:   UserPubSubCapnpServer{svc},
	}
	return capsrv
}
