package capnpclient

import (
	"capnproto.org/go/capnp/v3"
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

type ResolverCapnpClient struct {
	connection *rpc.Conn                 // connection to capnp server
	capability hubapi.CapResolverService // capnp client of the resolver service
}

// Capability of the capnp client used to talk to the resolver
func (cl *ResolverCapnpClient) Capability() hubapi.CapResolverService {
	return cl.capability
}

// ListCapabilities lists the available capabilities of the service
// Returns a list of capabilities that can be obtained through the service
func (cl *ResolverCapnpClient) ListCapabilities(
	ctx context.Context, authType string) (infoList []resolver.CapabilityInfo, err error) {

	infoList = make([]resolver.CapabilityInfo, 0)
	method, release := cl.capability.ListCapabilities(ctx,
		func(params hubapi.CapProvider_listCapabilities_Params) error {
			err2 := params.SetAuthType(authType)
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		infoListCapnp, err2 := resp.InfoList()
		if err = err2; err == nil {
			infoList = capserializer.UnmarshalCapabilyInfoList(infoListCapnp)
		}
	}
	return infoList, err
}

//
//// RegisterCapabilities registers a service's capabilities along with the CapProvider
//func (cl *ResolverCapnpClient) RegisterCapabilities(ctx context.Context,
//	serviceID string, capInfoList []resolver.CapabilityInfo,
//	capProvider hubapi.CapProvider) (err error) {
//
//	capInfoListCapnp := capserializer.MarshalCapabilityInfoList(capInfoList)
//	method, release := cl.capability.RegisterCapabilities(ctx,
//		func(params hubapi.CapResolverService_registerCapabilities_Params) error {
//			err = params.SetCapInfo(capInfoListCapnp)
//			_ = params.SetServiceID(serviceID)
//			_ = params.SetProvider(capProvider.AddRef()) // don't forget AddRef
//			return err
//		})
//	defer release()
//	_, err = method.Struct()
//	return err
//}

// Release the client
func (cl *ResolverCapnpClient) Release() {
	cl.capability.Release()
	if cl.connection != nil {
		err := cl.connection.Close()
		if err != nil {
			logrus.Error(err)
		}
	}
}

// NewResolverCapnpClientConnection create a new resolver client for obtaining capabilities.
// Intended for remote clients such as IoT devices, services or users to connect to the
// Hub's resolver. A connection must be established first.
//
//	conn is the network connection to use.
func NewResolverCapnpClientConnection(ctx context.Context, conn net.Conn) *ResolverCapnpClient {

	transport := rpc.NewStreamTransport(conn)
	rpcConn := rpc.NewConn(transport, nil)
	cl := NewResolverCapnpClient(rpcConn.Bootstrap(ctx))
	cl.connection = rpcConn
	return cl
}

// NewResolverCapnpClient create a new resolver client for obtaining capabilities.
func NewResolverCapnpClient(client capnp.Client) *ResolverCapnpClient {

	capability := hubapi.CapResolverService(client)

	cl := &ResolverCapnpClient{
		connection: nil,
		capability: capability,
	}
	return cl
}
