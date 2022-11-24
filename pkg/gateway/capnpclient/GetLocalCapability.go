package capnpclient

import (
	"context"
	"net"
	"path"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// GetLocalCapability connect to the local service and return the capability instance
// This returns the connection and capability of the service RPC server
func GetLocalCapability(ctx context.Context, socketFolder string, serviceName string) (
	rpcConn *rpc.Conn, capability hubapi.HiveService, err error) {

	socketPath := path.Join(socketFolder, serviceName+".socket")
	connection, err := net.Dial("unix", socketPath)
	if err == nil {
		transport := rpc.NewStreamTransport(connection)
		rpcConn = rpc.NewConn(transport, nil)
		capability = hubapi.HiveService(rpcConn.Bootstrap(ctx))
	}
	// Experiment with making a call, for example to get the capability return by a method.
	// While this works, it needs the interfaceID and methodID which are not easily obtained generically.
	// for now lets simply return the service capability and let the remote side figure out how to invoke any methods.
	//
	//// invoke the method on the service
	//if methodName != "" {
	//	//c2 := hubapi.CapGatewayService(capability)
	//	//method, release := c2.Ping(ctx, nil)
	//	s := capnp.Send{
	//		Method: capnp.Method{
	//			InterfaceID: hubapi.CapGatewayService_TypeID,
	//			MethodID:    2,
	//			//InterfaceName: "hubapi/Gateway.capnp:CapGatewayService",
	//			//MethodName:    "ping", //methodName,
	//		},
	//	}
	//	method, release := capability.SendCall(ctx, s)
	//	defer release()
	//	resp, err2 := method.Struct()
	//	resp2 := hubapi.CapGatewayService_ping_Results(resp)
	//	text, _ := resp2.Response()
	//	logrus.Infof("text: %s", text)
	//	err = err2
	//	_ = resp
	//}
	return rpcConn, capability, err
}
