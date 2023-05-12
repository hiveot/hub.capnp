package capprovider

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc/transport"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// CapServer implements the capnp server for serving canpnp Methods/Capabilities.
// This is similar to simply calling the MyService_ServerToClient(server),
// except that it also:
//  1. inject the ListCapabilities method that returns a list of exported capabilities.
//  2. adds a hook to handle requests for methods that aren't known on startup
//
// The handleUnknownMethod hook can be used to make a proxy service that forwards
// requests to the actual service. See the resolver and gateway for examples on how this is done.
//
// The list of methods are generated by the go-capnp compiler. For an interface called
// MyService, the list of methods can be obtained with MyService_Methods(nil,server)
// where 'server' is the instance that implements the methods.
//
// Usage:
//  1. call NewCapServer and provide the methods to serve.
//  2. call ExportCapability for each of the methods to include in ListCapabilities
//  3. call Start with the listening connection to start serving the methods
type CapServer struct {
	// This provider server capability
	capProviderCapability hubapi.CapProvider

	exportedCapabilities map[string]resolver.CapabilityInfo

	// Known methods of this service initialized on start
	knownMethods map[string]server.Method

	// ID of the service offering the capabilities
	serviceName string

	// The hook into the capnp protocol server
	clientHook *server.Server

	lis net.Listener
}

// ExportCapability adds the Method of the given name to the result of ListCapabilities.
// A method of this name must have been included when creating this NewCapServer.
//
//	methodName is the name of the method to include in ListCapabilities.
//	authTypes defines the required authentication type needed to use this capability
func (capsrv *CapServer) ExportCapability(methodName string, authTypes []string) {
	methodInfo, found := capsrv.knownMethods[methodName]
	if !found {
		err := fmt.Errorf("method '%s' is not a known method. Unable to enable it", methodName)
		logrus.Error(err)
		panic(err)
	}
	newCap := resolver.CapabilityInfo{
		InterfaceID:   methodInfo.InterfaceID,
		MethodID:      methodInfo.MethodID,
		InterfaceName: methodInfo.InterfaceName,
		MethodName:    methodName,
		AuthTypes:     authTypes,
		ServiceID:     capsrv.serviceName,
	}
	capsrv.exportedCapabilities[methodName] = newCap
}

// ListCapabilities returns the list of exported capabilities
func (capsrv *CapServer) ListCapabilities(
	_ context.Context, call hubapi.CapProvider_listCapabilities) (err error) {

	infoList := make([]resolver.CapabilityInfo, 0, len(capsrv.exportedCapabilities))
	for _, capInfo := range capsrv.exportedCapabilities {
		infoList = append(infoList, capInfo)
	}

	if err == nil {
		resp, err2 := call.AllocResults()
		err = err2
		if err == nil {
			infoListCapnp := capserializer.MarshalCapabilityInfoList(infoList)
			err = resp.SetInfoList(infoListCapnp)
		}
	}
	logrus.Infof("returned %d capabilities for '%s'", len(infoList), capsrv.serviceName)
	return err
}

// Set a handler to dynamically resolve method requests.
// Intended for proxying requests to other servers
func (capsrv *CapServer) SetUnknownMethodHandler(handler func(m capnp.Method) *server.Method) {
	capsrv.clientHook.HandleUnknownMethod = handler
}

// Start listening for incoming connections.
// This transfers ownership of the listener to this server and waits until the listener is closed
//
//	lis is a TCP or TLS listening socket.
//	useWS is to upgrade the listening sockets to http websockets.
//
// Use Stop() to stop listening and close the listener.
func (capsrv *CapServer) Start(lis net.Listener) error {
	//logrus.Infof("CapServer listening on %s", lis.Addr())
	//err := rpc.Serve(lis, capnp.Client(capsrv.capProviderCapability))
	capsrv.lis = lis
	err := listener.Serve(
		capsrv.serviceName,
		lis,
		capnp.Client(capsrv.capProviderCapability),
		nil, nil)
	return err
}

// StartWithWS starts listening using a capnp rpc transport over websockets.
//
//	lis will be used to start a http server whose requests are upgraded to websockets.
//	onConnection is na optional callback to track the session.
//
// This transfers ownership of the transport to this server and waits until the it is closed
// Use Stop() to stop listening and close the transport.
func (capsrv *CapServer) StartWithWS(
	lis net.Listener, wsPath string,
	onConnection func(conn net.Conn, capTransport transport.Transport)) error {
	logrus.Infof("CapServer listening on WS %s%s", lis.Addr(), wsPath)

	capsrv.lis = lis
	err := listener.ServeWS(
		capsrv.serviceName,
		lis,
		wsPath,
		capnp.Client(capsrv.capProviderCapability),
		nil, nil)
	return err
}

// Stop listening
//func (capsrv *CapServer) Stop() {
//	_ = capsrv.lis.Close()
//}

// NewCapServer prepares a capnproto protocol server for listening to incoming requests
// for capabilities/methods.
//
// ServiceName is the name of this service instance under which is can be reached through
// the resolver or gateway. It must be unique on the local network.
// The binding between capnp protocol and server is provided by the methods.
//
// Each capnp compiled interface makes this available through a '_Methods' API on
// the  '{ServiceName}_Methods'.
// For example for a service defined as MyService in capnp:
//
//	 svc := MyService{}
//	 svcMethods := TestService_Methods(nil, svc)
//	 capServer = capprovider.NewCapServer("testService", svcMethods)
//	 capServer.ExportCapability("method1", []string{hubapi.AuthTypeService})
//	 lis,_ := net.dial("unix","/path/to/socket")
//	 capServer.Start(lis)
//
//		serviceName identifies this service instance.
//		methods is the list of capnp methods to serve.
//
// This returns a capnp server that can handle all given methods.
func NewCapServer(serviceName string, methods []server.Method) *CapServer {
	// logrus.Infof("serviceName=%s", serviceName)

	srv := &CapServer{
		exportedCapabilities: make(map[string]resolver.CapabilityInfo),
		knownMethods:         make(map[string]server.Method),
		serviceName:          serviceName,
	}
	for _, method := range methods {
		srv.knownMethods[method.MethodName] = method
	}
	// the following code replaces the usual Xyz_ServerToClient() call. Instead, the capnp server passes
	// its Methods list to be served by this server.
	srv.capProviderCapability = hubapi.CapProvider_ServerToClient(srv)

	// get the shutdown method of the server if it has one
	c, _ := hubapi.CapProvider_Server(srv).(server.Shutdowner)

	// Inject the ListCapabilities method provided by this server
	allMethods := hubapi.CapProvider_Methods(methods, srv)
	srv.clientHook = server.New(allMethods, srv, c)

	// turn it into a capability client for use by the server
	providerClient := capnp.NewClient(srv.clientHook)
	// cast the client to the provider for things like AddRef and Release
	srv.capProviderCapability = hubapi.CapProvider(providerClient)

	return srv
}
