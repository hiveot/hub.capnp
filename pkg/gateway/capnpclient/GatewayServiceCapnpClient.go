package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/gateway"
)

type GatewayServiceCapnpClient struct {
	connection *rpc.Conn                // connection to capnp server
	capability hubapi.CapGatewayService // capnp client of the gateway service
}

// GetCapability obtains the capability with the given name, if available
// This returns the remote capability that still has to be wrapped in the POGS client for
// that capability.
//
// for example:
//
//	cap, err := GetCapability(ctx, gateway.ClientTypeService, pubsub.ServiceName, "CapServicePubSub", "urn:myservicename")
//	servicePubSub := NewServicePubSubCapnpClient(cap)  // IServicePubSub
func (cl *GatewayServiceCapnpClient) GetCapability(ctx context.Context,
	clientType string, service string) (interface{}, error) {

	method, release := cl.capability.GetCapability(ctx,
		func(params hubapi.CapGatewayService_getCapability_Params) error {
			_ = params.SetService(service)
			err := params.SetClientType(clientType)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err != nil {
		return nil, err
	}
	capability := resp.Cap()
	capclient := capnp.Client(capability.AddRef())
	return capclient, err
}

// GetGatewayInfo describes the capabilities and capacity of the gateway
func (cl *GatewayServiceCapnpClient) GetGatewayInfo(
	ctx context.Context) (gwInfo gateway.GatewayInfo, err error) {

	method, release := cl.capability.GetGatewayInfo(ctx, nil)
	defer release()

	resp, err := method.Struct()
	if err == nil {
		infoCapnp, err2 := resp.Info()
		err = err2
		if err2 == nil {
			// Unmarshal the gateway info
			gwInfo.URL, _ = infoCapnp.Url()
			gwInfo.Latency = int(infoCapnp.Latency())
			gwInfo.Capabilities = make([]gateway.CapabilityInfo, 0)
			//
			infoList, err2 := infoCapnp.Capabilities()
			err = err2
			if err2 == nil {
				for i := 0; i < infoList.Len(); i++ {
					capInfoCapnp := infoList.At(i)
					service, _ := capInfoCapnp.Service()
					name, _ := capInfoCapnp.Name()
					clientType, _ := capInfoCapnp.ClientType()
					capInfoPogs := gateway.CapabilityInfo{
						Service:    service,
						Name:       name,
						ClientType: caphelp.UnmarshalStringList(clientType),
					}
					gwInfo.Capabilities = append(gwInfo.Capabilities, capInfoPogs)
				}
			}
		}
	}
	return gwInfo, err
}

// Ping performs a ping test
func (cl *GatewayServiceCapnpClient) Ping(ctx context.Context) (response string, err error) {
	method, release := cl.capability.Ping(ctx, nil)
	defer release()

	resp, err := method.Struct()
	if err == nil {
		response, err = resp.Response()
	}
	return response, err
}

// Stop the service and release its resources
func (cl *GatewayServiceCapnpClient) Stop(_ context.Context) (err error) {
	cl.capability.Release()
	if cl.connection != nil {
		err = cl.connection.Close()
	}
	return err
}

func NewGatewayServiceCapnpClient(ctx context.Context,
	connection net.Conn) (cl *GatewayServiceCapnpClient, err error) {

	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapGatewayService(rpcConn.Bootstrap(ctx))

	cl = &GatewayServiceCapnpClient{
		connection: rpcConn,
		capability: capability,
	}
	return cl, nil
}

func NewGatewayServiceFromCapability(capability capnp.Client) (cl *GatewayServiceCapnpClient) {

	cl = &GatewayServiceCapnpClient{
		capability: hubapi.CapGatewayService(capability),
	}
	return cl
}
