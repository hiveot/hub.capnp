package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
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
	thingAddr, _ := args.ThingAddr()
	name, _ := args.ActionName()
	value, _ := args.Value()
	err := capsrv.svc.PubAction(ctx, thingAddr, name, value)
	return err
}

func (capsrv *UserPubSubCapnpServer) SubEvent(
	ctx context.Context, call hubapi.CapUserPubSub_subEvent) error {
	args := call.Args()
	thingAddr, _ := args.ThingAddr()
	name, _ := args.EventName()
	handlerCap := args.Handler()
	handler := NewSubscriptionHandlerCapnpClient(handlerCap.AddRef())
	err := capsrv.svc.SubEvent(ctx, thingAddr, name, handler.HandleValue)
	return err
}

func (capsrv *UserPubSubCapnpServer) SubTDs(
	ctx context.Context, call hubapi.CapUserPubSub_subTDs) error {
	args := call.Args()
	handlerCap := args.Handler()
	handler := NewSubscriptionHandlerCapnpClient(handlerCap.AddRef())
	err := capsrv.svc.SubTDs(ctx, handler.HandleValue)
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
