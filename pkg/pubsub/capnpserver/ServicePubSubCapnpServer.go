package capnpserver

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
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
	thingID, _ := args.ThingID()
	publisherID, _ := args.PublisherID()
	actionID, _ := args.ActionID()
	handlerCap := args.Handler()
	handlerClient := NewSubscriptionHandlerCapnpClient(handlerCap.AddRef())
	// the server registers the callback handler and invokes it when actions for the Thing are received
	err := capsrv.svc.SubActions(ctx, publisherID, thingID, actionID, handlerClient.HandleValue)
	return err
}

func (capsrv *ServicePubSubCapnpServer) SubEvents(
	ctx context.Context, call hubapi.CapServicePubSub_subEvents) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	publisherID, _ := args.PublisherID()
	eventID, _ := args.EventID()
	handlerCap := args.Handler()
	handlerClient := NewSubscriptionHandlerCapnpClient(handlerCap.AddRef())
	// the server registers the callback handler and invokes it when actions for the Thing are received
	err := capsrv.svc.SubEvents(ctx, publisherID, thingID, eventID, handlerClient.HandleValue)
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
