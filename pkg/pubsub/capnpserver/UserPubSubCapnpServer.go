package capnpserver

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/pubsub"
)

// UserPubSubCapnpServer provides the capnp RPC server for user pubsub services.
// This implements the capnproto generated interface CapUserPubSub_Server
type UserPubSubCapnpServer struct {
	svc pubsub.IUserPubSub
}

func (capsrv *UserPubSubCapnpServer) PubAction(
	ctx context.Context, call hubapi.CapUserPubSub_pubAction) error {

	args := call.Args()
	thingID, _ := args.ThingID()
	publisherID, _ := args.PublisherID()
	actionID, _ := args.ActionID()
	value, _ := args.Value()
	err := capsrv.svc.PubAction(ctx, publisherID, thingID, actionID, value)
	return err
}

func (capsrv *UserPubSubCapnpServer) SubEvent(
	ctx context.Context, call hubapi.CapUserPubSub_subEvent) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	publisherID, _ := args.PublisherID()
	eventID, _ := args.EventID()
	handlerCap := args.Handler()
	handler := NewSubscriptionHandlerCapnpClient(handlerCap.AddRef())

	//logrus.Infof("subscribing to event %s/%s/%s", publisherID, thingID, eventID)

	err := capsrv.svc.SubEvent(ctx, publisherID, thingID, eventID, handler.HandleValue)
	return err
}

func (capsrv *UserPubSubCapnpServer) Shutdown() {
	// Client is released, release the subscriptions
	capsrv.svc.Release()
}

// NewUserPubSubCapnpServer returns a capnp server for receiving callbacks
func NewUserPubSubCapnpServer(svc pubsub.IUserPubSub) *UserPubSubCapnpServer {
	capsvc := &UserPubSubCapnpServer{
		svc: svc,
	}
	return capsvc
}
