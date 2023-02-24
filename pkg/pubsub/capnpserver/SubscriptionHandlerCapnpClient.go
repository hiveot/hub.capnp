package capnpserver

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/lib/caphelp"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/thing"
)

// SubscriptionHandlerCapnpClient provides the client side of the subscription callback.
// This is call by the server to pass a message to the subscription handler.
// This implements the capnp SubscriptionHandler API
type SubscriptionHandlerCapnpClient struct {
	// the capnp generated handler of the subscription callback api
	handlerCapnp hubapi.CapSubscriptionHandler
}

// HandleValue invokes the remote callback handler with the given value
func (cl *SubscriptionHandlerCapnpClient) HandleValue(value *thing.ThingValue) {
	ctx := context.Background()
	method, release := cl.handlerCapnp.HandleValue(ctx,
		func(params hubapi.CapSubscriptionHandler_handleValue_Params) error {
			tvCap := caphelp.MarshalThingValue(value)
			err := params.SetValue(tvCap)
			return err
		})
	_, err := method.Struct()
	if err != nil {
		logrus.Errorf("failed invoking callback: %s", err)
	}
	release()
}

func (cl *SubscriptionHandlerCapnpClient) Release() {
	cl.handlerCapnp.Release()
}

// NewSubscriptionHandlerCapnpClient returns a POGS client instance of the capnp subscription handler
// that invokes the remote handler over capnp RPC.
//
//	handlerCapnp is the capability provided by?
func NewSubscriptionHandlerCapnpClient(handlerCapnp hubapi.CapSubscriptionHandler) *SubscriptionHandlerCapnpClient {
	cl := &SubscriptionHandlerCapnpClient{
		handlerCapnp: handlerCapnp,
	}
	return cl
}
