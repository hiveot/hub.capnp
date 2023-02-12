package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
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
	ctx context.Context, deviceID string) (capability pubsub.IDevicePubSub, err error) {

	method, release := cl.capability.CapDevicePubSub(ctx,
		func(params hubapi.CapPubSubService_capDevicePubSub_Params) error {
			err := params.SetDeviceID(deviceID)
			return err
		})
	defer release()
	capFuture := method.Cap()
	newCap := NewDevicePubSubCapnpClient(capFuture.AddRef())
	return newCap, err
}

// CapServicePubSub provides the capability to pub/sub thing information as a hub service.
func (cl *PubSubCapnpClient) CapServicePubSub(
	ctx context.Context, serviceID string) (capability pubsub.IServicePubSub, err error) {

	method, release := cl.capability.CapServicePubSub(ctx,
		func(params hubapi.CapPubSubService_capServicePubSub_Params) error {
			err2 := params.SetServiceID(serviceID)
			return err2
		})
	defer release()
	capFuture := method.Cap()
	newCap := NewServicePubSubCapnpClient(capFuture.AddRef())
	return newCap, err
}

// CapUserPubSub provides the capability for an end-user to publish or subscribe to messages.
func (cl *PubSubCapnpClient) CapUserPubSub(
	ctx context.Context, userID string) (capability pubsub.IUserPubSub, err error) {

	method, release := cl.capability.CapUserPubSub(ctx, func(params hubapi.CapPubSubService_capUserPubSub_Params) error {
		err := params.SetUserID(userID)
		return err
	})
	defer release()
	capFuture := method.Cap()
	userCl := NewUserPubSubCapnpClient(capFuture.AddRef())
	return userCl, err
}

// Release stops the client and frees its resources
// If the rpc connection was made on instantiation, it will be closed.
func (cl *PubSubCapnpClient) Release() {
	cl.capability.Release()
	if cl.connection != nil {
		_ = cl.connection.Close()
	}
}

// NewPubSubCapnpClient creates a new client for using the pubsub service with the given connection.
// After use, the caller must invoke Release
func NewPubSubCapnpClient(ctx context.Context, c net.Conn) *PubSubCapnpClient {
	var cl *PubSubCapnpClient

	// use a direct connection to the service
	transport := rpc.NewStreamTransport(c)
	rpcConn := rpc.NewConn(transport, nil)
	capPubSub := hubapi.CapPubSubService(rpcConn.Bootstrap(ctx))

	cl = &PubSubCapnpClient{
		connection: rpcConn,
		capability: capPubSub,
	}
	return cl
}

// NewPubSubClient creates a new client for using the pubsub service with the given capnp client.
// The capnp client can be that of the service, the resolver or the gateway
func NewPubSubClient(capClient capnp.Client) *PubSubCapnpClient {
	var cl *PubSubCapnpClient

	// use a direct connection to the service
	capPubSub := hubapi.CapPubSubService(capClient)
	cl = &PubSubCapnpClient{
		connection: nil,
		capability: capPubSub,
	}
	return cl
}
