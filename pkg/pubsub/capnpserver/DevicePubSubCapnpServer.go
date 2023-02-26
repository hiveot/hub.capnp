package capnpserver

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
	"github.com/hiveot/hub/pkg/pubsub"
)

// DevicePubSubCapnpServer provides the capnp RPC server for device pubsub services.
// This implements the capnproto generated interface CapDevicePubSub_Server
type DevicePubSubCapnpServer struct {
	svc pubsub.IDevicePubSub
}

func (capsrv *DevicePubSubCapnpServer) PubEvent(
	ctx context.Context, call hubapi.CapDevicePubSub_pubEvent) error {

	args := call.Args()
	thingID, _ := args.ThingID()
	eventID, _ := args.EventID()
	value, _ := args.Value()
	err := capsrv.svc.PubEvent(ctx, thingID, eventID, value)
	return err
}

func (capsrv *DevicePubSubCapnpServer) PubProperties(
	ctx context.Context, call hubapi.CapDevicePubSub_pubProperties) error {

	args := call.Args()
	thingID, _ := args.ThingID()
	propsCap, _ := args.Props()
	props := caphelp.UnmarshalKeyValueMap(propsCap)
	err := capsrv.svc.PubProperties(ctx, thingID, props)
	return err
}

func (capsrv *DevicePubSubCapnpServer) PubTD(
	ctx context.Context, call hubapi.CapDevicePubSub_pubTD) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	tdDoc, _ := args.TdDoc()
	err := capsrv.svc.PubTD(ctx, thingID, tdDoc)
	return err
}

func (capsrv *DevicePubSubCapnpServer) SubAction(
	ctx context.Context, call hubapi.CapDevicePubSub_subAction) error {
	args := call.Args()
	thingID, _ := args.ThingID()
	actionID, _ := args.ActionID()
	handlerCap := args.Handler()
	handlerClient := NewSubscriptionHandlerCapnpClient(handlerCap.AddRef())
	// The server registers the handler and invokes it when an action request is received
	err := capsrv.svc.SubAction(ctx, thingID, actionID, handlerClient.HandleValue)
	return err
}

func (capsrv *DevicePubSubCapnpServer) Shutdown() {
	// Client is released, release the subscriptions
	capsrv.svc.Release()
}

func NewDevicePubSubCapnpServer(svc pubsub.IDevicePubSub) *DevicePubSubCapnpServer {
	capsvc := &DevicePubSubCapnpServer{
		svc: svc,
	}
	return capsvc
}
