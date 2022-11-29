package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

type GatewayServiceCapnpClient struct {
	// implement the IHiveOTService interface
	//caphelp.HiveOTServiceCapnpClient
	connection *rpc.Conn                // connection to capnp server
	capability hubapi.CapGatewayService // capnp client of the gateway service
}

// GetCapability obtains the capability with the given name.
// The caller must release the capability when done.
func (cl *GatewayServiceCapnpClient) GetCapability(ctx context.Context,
	clientID string, clientType string, capabilityName string, args []string) (
	capabilityRef capnp.Client, err error) {

	method, release := cl.capability.GetCapability(ctx,
		func(params hubapi.CapGatewayService_getCapability_Params) error {
			_ = params.SetClientID(clientID)
			_ = params.SetClientType(clientType)
			_ = params.SetCapabilityName(capabilityName)
			if args != nil {
				err = params.SetArgs(caphelp.MarshalStringList(args))
			}
			return err
		})
	defer release()
	// return a future. Caller must release
	//capability = method.Cap().AddRef()

	// Just return the actual capability instead of a future, so the error is obtained if it isn't available.
	// Would be nice to return the future but this is an infrequent call anyways.
	resp, err := method.Struct()
	if err == nil {
		capability := resp.Capability().AddRef()
		capabilityRef = capability
	}
	return capabilityRef, err
}

// ListCapabilities lists the available capabilities of the service
// Returns a list of capabilities that can be obtained through the service
func (cl *GatewayServiceCapnpClient) ListCapabilities(
	ctx context.Context, clientType string) (infoList []caphelp.CapabilityInfo, err error) {

	infoList = make([]caphelp.CapabilityInfo, 0)
	method, release := cl.capability.ListCapabilities(ctx,
		func(params hubapi.CapGatewayService_listCapabilities_Params) error {
			err = params.SetClientType(clientType)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		infoListCapnp, err2 := resp.InfoList()
		if err = err2; err == nil {
			infoList = caphelp.UnmarshalCapabilities(infoListCapnp)
		}
	}
	return infoList, err
}

// Login to the gateway
func (cl *GatewayServiceCapnpClient) Login(ctx context.Context,
	clientID string, password string) (success bool, err error) {

	success = false
	method, release := cl.capability.Login(ctx,
		func(params hubapi.CapGatewayService_login_Params) error {
			err = params.SetClientID(clientID)
			params.SetPassword(password)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		success = resp.Success()
	}
	return success, err
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
	capGatewayService := hubapi.CapGatewayService(rpcConn.Bootstrap(ctx))
	//capHiveOTService := hubapi.CapHiveOTService(capGatewayService)

	cl = &GatewayServiceCapnpClient{
		//HiveOTServiceCapnpClient: *caphelp.NewHiveOTServiceCapnpClient(capHiveOTService),
		connection: rpcConn,
		capability: capGatewayService,
	}
	return cl, nil
}

// NewGatewayServiceFromCapability creates a capnp client of the gateway service
// capability is the gateway CapGatewayService obtained through getCapability.
func NewGatewayServiceFromCapability(capability hubapi.CapGatewayService) (cl *GatewayServiceCapnpClient) {
	//capHiveOTService := hubapi.CapHiveOTService(capability)

	cl = &GatewayServiceCapnpClient{
		capability: hubapi.CapGatewayService(capability),
		//HiveOTServiceCapnpClient: *caphelp.NewHiveOTServiceCapnpClient(capHiveOTService),
	}
	return cl
}
