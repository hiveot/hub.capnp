package capnpclient

import (
	"context"

	"github.com/hiveot/hub/lib/caphelp"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/thing"
)

// SubscriptionHandlerCapnpServer is the capnp server callback for topic subscriptions.
// This is used by the client to receive callbacks and lives on the client side of the RPC connection.
// (client is a server and the server is its client. Get it? :))
// This implements the hubapi.CapSubscriptionHandler interface
type SubscriptionHandlerCapnpServer struct {
	handler func(value *thing.ThingValue)
}

// HandleValue is a Capnp Server method that invokes the client provided callback
// This unmarshals the ThingValue and passes it to the callback
func (capsrv *SubscriptionHandlerCapnpServer) HandleValue(
	ctx context.Context, call hubapi.CapSubscriptionHandler_handleValue) error {
	args := call.Args()
	tvCap, _ := args.Value()
	tv := caphelp.UnmarshalThingValue(tvCap)
	capsrv.handler(tv)
	return nil
}

// Shutdown - nothing to do
// Subscriptions are removed from the server when the capability closes, which is before the client is released
func (capsrv *SubscriptionHandlerCapnpServer) Shutdown() {
	// this is released if the capnp client is released
	//capsrv.handler.Release()
	//logrus.Infof("SubscriptionHandlerCapnpServer was released ... somehow?")
}

func NewSubscriptionHandlerCapnpServer(handler func(value *thing.ThingValue)) hubapi.CapSubscriptionHandler {
	capsrv := &SubscriptionHandlerCapnpServer{handler: handler}
	capability := hubapi.CapSubscriptionHandler_ServerToClient(capsrv)
	return capability
}
