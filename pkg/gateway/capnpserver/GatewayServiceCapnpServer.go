package capnpserver

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/gateway"
)

// GatewayServiceCapnpServer implements the capnp server of the gateway service
// This server does something special for the GetCapability method.
// Instead of calling the POGS service, it connects to the requested service directly, obtains the
// capnp capability and returns it to the client. This is a pure capnp 'capability' (pun intended)
// This implements the capnp hubapi.CapGatewayService_server interface.
type GatewayServiceCapnpServer struct {
	caphelp.HiveOTServiceCapnpServer // getCapability and listCapabilities
	svc                              gateway.IGatewayService
	socketFolder                     string
}

// GetCapability invokes the requested method to return the capability it provides
// This returns an error if the capability is not found or not available to the client type
func (capsrv *GatewayServiceCapnpServer) GetCapability(
	ctx context.Context, call hubapi.CapGatewayService_getCapability) (err error) {

	args := call.Args()
	capabilityName, _ := args.CapabilityName()
	clientID, _ := args.ClientID()
	clientType, _ := args.ClientType()
	methodArgsCapnp, _ := args.Args()
	methodArgs := caphelp.UnmarshalStringList(methodArgsCapnp)
	//_ = methodArgs
	cap, err := capsrv.svc.GetCapability(ctx, clientID, clientType, capabilityName, methodArgs)
	if err != nil {
		return err
	}
	resp, err := call.AllocResults()
	if err != nil {
		return err
	}
	resp.SetCapability(cap)

	//
	//// validate the client type is allowed on this method
	//allowedTypes := strings.Join(capInfo.ClientTypes, ",")
	//isAllowed := clientType != "" && strings.Contains(allowedTypes, clientType)
	//if !isAllowed {
	//	err = fmt.Errorf("capability '%s' is not available to clients of type '%s'", capabilityName, clientType)
	//	return err
	//}
	//// invoke the method to get the capability
	//// the clientID and clientType arguments from this call are passed on to the capability request
	//// TODO: is there a need for further arugments?
	//method, _ := capsrv.knownMethods[capabilityName]
	//mc := call.Call
	////for i, argText := range methodArgs {
	////	mc.Args().SetText()
	////}
	//err = method.Impl(ctx, mc)
	return err
}

// ListCapabilities returns the list of registered capabilities
// Only capabilities with a client type are returned.
func (capsrv *GatewayServiceCapnpServer) ListCapabilities(
	ctx context.Context, call hubapi.CapGatewayService_listCapabilities) (err error) {

	args := call.Args()
	clientType, _ := args.ClientType()
	infoList, err := capsrv.svc.ListCapabilities(ctx, clientType)
	if err == nil {
		resp, err2 := call.AllocResults()
		if err = err2; err == nil {
			infoListCapnp := caphelp.MarshalCapabilities(infoList)
			err = resp.SetInfoList(infoListCapnp)
		}
	}
	return err
}

func (capsrv *GatewayServiceCapnpServer) Ping(
	ctx context.Context, call hubapi.CapGatewayService_ping) error {

	response, err := capsrv.svc.Ping(ctx)
	if err != nil {
		err = fmt.Errorf("ping somehow managed to fail")
		logrus.Error(err)
		return err
	}
	res, err := call.AllocResults()

	if err == nil {
		err = res.SetResponse(response)
	}
	return err
}

func (capsrv *GatewayServiceCapnpServer) Login(
	ctx context.Context, call hubapi.CapGatewayService_login) error {
	args := call.Args()
	loginID, _ := args.ClientID()
	password, _ := args.Password()
	success, err := capsrv.svc.Login(ctx, loginID, password)
	if err == nil {
		resp, err := call.AllocResults()
		if err == nil {
			resp.SetSuccess(success)
		}
	}
	return err
}

func StartGatewayServiceCapnpServer(
	ctx context.Context, lis net.Listener, svc gateway.IGatewayService, socketFolder string) error {

	srv := &GatewayServiceCapnpServer{
		//HiveOTServiceCapnpServer: caphelp.NewHiveOTServiceCapnpServer(gateway.ServiceName),
		svc:          svc,
		socketFolder: socketFolder,
	}
	// register the methods available through getCapability
	//methods := hubapi.CapGatewayService_Methods(nil, srv)
	//srv.RegisterKnownMethods(methods)
	//srv.ExportCapability("getPingCap", []string{hubapi.ClientTypeUser, hubapi.ClientTypeService, hubapi.ClientTypeIotDevice, hubapi.ClientTypeUnauthenticated})

	main := hubapi.CapGatewayService_ServerToClient(srv)
	err := caphelp.Serve(lis, capnp.Client(main))
	return err
}
