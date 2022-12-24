package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
)

// PubSubCapnpClient is the capnp client for the pubsub service
type PubSubCapnpClient struct {
	// use either the capability provider via the resolver,...
	resolverClient resolver.IResolverSession
	// or the direct connection to the server
	capability hubapi.CapPubSubService

	connection *rpc.Conn // connection to capnp server
}

// CapDevicePubSub provides the capability to pub/sub thing information as an IoT device.
func (cl *PubSubCapnpClient) CapDevicePubSub(
	ctx context.Context, deviceID string) (deviceCl pubsub.IDevicePubSub) {
	// using capclient is experimental
	if cl.resolverClient != nil {
		capability, err := cl.resolverClient.GetCapability(ctx, deviceID, hubapi.ClientTypeIotDevice, "capDevicePubSub", nil)
		if err == nil {
			deviceCl = NewDevicePubSubCapnpClient(hubapi.CapDevicePubSub(capability))
		}
	} else {

		method, release := cl.capability.CapDevicePubSub(ctx, func(params hubapi.CapPubSubService_capDevicePubSub_Params) error {
			err := params.SetDeviceID(deviceID)
			return err
		})
		defer release()
		capability := method.Cap()
		deviceCl = NewDevicePubSubCapnpClient(capability.AddRef())
	}
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

// StartPubSubCapnpClient creates a new client for using the pubsub service.
// This can be used in 2 modes:
// 1. direct mode: connect directly to the pubsub service using the given connection
// 2. resolver mode: use the given connection to the resolver service or create one if nil
//
// connection is optional and intended for direct connection to the pubsub service
// if nil, the resolver is used to find the pubsub capability.
// After use, the caller must invoke Release
func StartPubSubCapnpClient(ctx context.Context, connection net.Conn) (*PubSubCapnpClient, error) {
	var cl *PubSubCapnpClient
	var capPubSub hubapi.CapPubSubService
	var rpcConn *rpc.Conn
	var resolverClient resolver.IResolverSession
	var err error

	// use a direct connection to the service
	if connection != nil {
		transport := rpc.NewStreamTransport(connection)
		rpcConn = rpc.NewConn(transport, nil)
		capPubSub = hubapi.CapPubSubService(rpcConn.Bootstrap(ctx))
	} else {
		// use the resolver service
		resolverClient, err = capnpclient.ConnectToResolver("")
	}
	cl = &PubSubCapnpClient{
		resolverClient: resolverClient,
		connection:     rpcConn,
		capability:     capPubSub,
	}
	return cl, err
}
