package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/pubsub"
)

// PubSubCapnpClient is the capnp client for the pubsub service
type PubSubCapnpClient struct {
	capability hubapi.CapPubSubService
	connection *rpc.Conn // connection to capnp server
}

// CapDevicePubSub provides the capability to pub/sub thing information as an IoT device.
func (cl *PubSubCapnpClient) CapDevicePubSub(
	ctx context.Context, deviceID string) (deviceCl pubsub.IDevicePubSub) {

	method, release := cl.capability.CapDevicePubSub(ctx,
		func(params hubapi.CapPubSubService_capDevicePubSub_Params) error {
			err := params.SetDeviceID(deviceID)
			return err
		})
	defer release()
	capability := method.Cap()
	newCap := NewDevicePubSubCapnpClient(capability.AddRef())
	return newCap
}

// CapServicePubSub provides the capability to pub/sub thing information as a hub service.
func (cl *PubSubCapnpClient) CapServicePubSub(
	ctx context.Context, serviceID string) pubsub.IServicePubSub {

	method, release := cl.capability.CapServicePubSub(ctx,
		func(params hubapi.CapPubSubService_capServicePubSub_Params) error {
			err := params.SetServiceID(serviceID)
			return err
		})
	defer release()
	capability := method.Cap()
	newCap := NewServicePubSubCapnpClient(capability.AddRef())
	return newCap
}

// CapUserPubSub provides the capability for an end-user to publish or subscribe to messages.
func (cl *PubSubCapnpClient) CapUserPubSub(
	ctx context.Context, userID string) (pub pubsub.IUserPubSub) {

	method, release := cl.capability.CapUserPubSub(ctx, func(params hubapi.CapPubSubService_capUserPubSub_Params) error {
		err := params.SetUserID(userID)
		return err
	})
	defer release()
	capability := method.Cap()
	userCl := NewUserPubSubCapnpClient(capability.AddRef())
	return userCl
}

// Release stops the client connection and free its resources
func (cl *PubSubCapnpClient) Release() error {
	cl.capability.Release()
	err := cl.connection.Close()
	return err
}

// NewPubSubCapnpClient creates a new client for using the pubsub service.
//
// connection is optional and intended for direct connection to the pubsub service
// if nil, the resolver is used to find the pubsub capability.
// After use, the caller must invoke Release
func NewPubSubCapnpClient(ctx context.Context, connection net.Conn) *PubSubCapnpClient {
	var cl *PubSubCapnpClient

	// use a direct connection to the service
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capPubSub := hubapi.CapPubSubService(rpcConn.Bootstrap(ctx))

	cl = &PubSubCapnpClient{
		connection: rpcConn,
		capability: capPubSub,
	}
	return cl
}
