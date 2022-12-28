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
	destination *capnp.Client
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
	ans, release := (*m.destination).SendCall(ctx, s)
	defer release()

	res, err := ans.Struct()
	if err != nil {
		logrus.Error(err)
		return err
	}
	res2, err := call.AllocResults(res.Size())
	if err != nil {
		logrus.Error(err)
		return err
	}
	// copy the answer directly
	err = res2.CopyFrom(res)

	return err
}
