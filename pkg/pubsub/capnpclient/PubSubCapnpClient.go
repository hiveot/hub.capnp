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
	connection *rpc.Conn               // connection to capnp server
	capability hubapi.CapPubSubService // capnp client of the directory
}

// CapDevicePubSub provides the capability to pub/sub thing information as an IoT device.
func (cl *PubSubCapnpClient) CapDevicePubSub(
	ctx context.Context, deviceID string) pubsub.IDevicePubSub {

	method, release := cl.capability.CapDevicePubSub(ctx, func(params hubapi.CapPubSubService_capDevicePubSub_Params) error {
		err := params.SetDeviceID(deviceID)
		return err
	})
	defer release()
	capability := method.Cap()
	deviceCl := NewDevicePubSubCapnpClient(capability.AddRef())
	return deviceCl
}

// CapServicePubSub provides the capability to pub/sub thing information as a hub service.
func (cl *PubSubCapnpClient) CapServicePubSub(
	ctx context.Context, serviceID string) pubsub.IServicePubSub {

	method, release := cl.capability.CapServicePubSub(ctx, func(params hubapi.CapPubSubService_capServicePubSub_Params) error {
		err := params.SetServiceID(serviceID)
		return err
	})
	defer release()
	capability := method.Cap()
	serviceCl := NewServicePubSubCapnpClient(capability.AddRef())
	return serviceCl
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

// StartPubSubCapnpClient create a new capnp RPC client using the given connection.
// The connection will be released by the client after it is Released.
func StartPubSubCapnpClient(ctx context.Context, connection net.Conn) (*PubSubCapnpClient, error) {
	var cl *PubSubCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapPubSubService(rpcConn.Bootstrap(ctx))

	cl = &PubSubCapnpClient{
		connection: rpcConn,
		capability: capability,
	}
	return cl, nil
}
