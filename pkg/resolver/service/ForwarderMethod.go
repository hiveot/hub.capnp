package service

import (
	"context"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"
)

// ForwarderMethod forwards the method to the destination client
type ForwarderMethod struct {
	capnp.Method
	destination capnp.Client
}

// Impl forward the requested method to the cap destination
func (m *ForwarderMethod) Impl(ctx context.Context, call *server.Call) error {
	// forward the request
	logrus.Infof("forward request to: %s:%s", m.InterfaceName, m.MethodName)

	// create a new message with a copy of the call args and the new destination
	s := capnp.Send{
		Method:   m.Method,
		ArgsSize: call.Args().Size(),
		PlaceArgs: func(s capnp.Struct) error {
			err := s.CopyFrom(call.Args())
			return err
		},
	}
	// Pass the message to the remote destination
	ans, release := (m.destination).SendCall(ctx, s)
	defer release()

	res, err := ans.Struct()
	if err != nil {
		return err
	}
	// TODO: if a capability is returned does this need an AddRef?
	// from testing this doesn't seem to be the case fortunately, although its puzzling...
	// Does the result add to a capability table and isn't that removed when the call is released?
	res2, err := call.AllocResults(res.Size())
	if err != nil {
		logrus.Error(err)
		return err
	}
	// copy the answer directly
	err = res2.CopyFrom(res)

	return err
}

// NewForwarderMethod creates a new server method that forwards the capnp method to its destination
func NewForwarderMethod(method capnp.Method, destination capnp.Client) *server.Method {
	forwarder := ForwarderMethod{
		Method:      method,
		destination: destination,
	}

	return &server.Method{
		Method: method,
		Impl:   forwarder.Impl,
	}

}
