package capprovider

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// CapServer implements the capnp capability server for the hubapi.CapProvider interface
// This injects a ListCapability method that allows the resolver to retrieve the list of available capabilities.
type CapServer struct {
	// This provider server capability
	capProviderCapability hubapi.CapProvider

	exportedCapabilities map[string]resolver.CapabilityInfo

	// Known methods of this service initialized on start
	knownMethods map[string]server.Method

	// Name of the service offering the capabilities
	serviceName string

	lis net.Listener
}

// ExportCapability exports the name of the method that provides the capability.
// The method is implemented by the service bootstrap. By convention this name is the
// same as the name of the capability/interface that is provided.
// This should be called before invoking Connect.
//
//	methodName is the name of the method in the service bootstrap interface that provides the capability
//	clientTypes defines the type of clients for whom this capability is intended.
func (capsrv *CapServer) ExportCapability(methodName string, clientTypes []string) {
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
		ClientTypes:   clientTypes,
		ServiceID:     capsrv.serviceName,
	}
	capsrv.exportedCapabilities[methodName] = newCap
}

// GetCapability invokes the requested method to return the capability it provides
//func (capsrv *CapServer) GetCapability(
//	ctx context.Context, call hubapi.CapProvider_getCapability) (err error) {
//
//	args := call.Args()
//	clientID, _ := args.ClientID()
//	capabilityName, _ := args.CapName()
//	clientType, _ := args.ClientType()
//	methodArgsCapnp, _ := args.Args()
//	methodArgs := caphelp.UnmarshalStringList(methodArgsCapnp)
//
//	capInfo, found := capsrv.exportedCapabilities[capabilityName]
//	if !found {
//		err = fmt.Errorf("capability '%s' not found", capabilityName)
//		return err
//	}
//	logrus.Infof("clientID='%s', clientType='%s', capabilityName='%s', args='%v'",
//		clientID, clientType, capabilityName, methodArgs)
//
//	// validate the client type is allowed on this method
//	allowedTypes := strings.Join(capInfo.ClientTypes, ",")
//	isAllowed := clientType != "" && strings.Contains(allowedTypes, clientType)
//	if !isAllowed {
//		err = fmt.Errorf("denied: capability '%s' is not available to client '%s' of type '%s'",
//			capabilityName, clientID, clientType)
//		logrus.Warning(err)
//		return err
//	}
//	// invoke the method to get the capability
//	// the clientID and clientType arguments from this call are passed on to the capability request
//	// and available in the method.
//	// okay, not quite sure how this works but the results of the method are applied to 'call'
//	// and returned by this method. The 'Capability' result doesn't need a matching name apparently as
//	// the first result from the capability table in the message is used. Quite convenient.
//	// TBD: Can this behavior be relied on in future versions of go-capnp?
//	method, _ := capsrv.knownMethods[capabilityName]
//	mc := call.Call
//	err = method.Impl(ctx, mc)
//
//	return err
//}

//func (capsrv *CapServer) HandleUnknownMethod(ctx context.Context, r capnp.Recv) *server.Method {
//	r.Reject(capnp.Unimplemented("unimplemented"))
//	return nil
//}
//

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
	return err
}

// Start listening for incoming connections.
// This transfers ownership of the listener to this server and waits until the listener is closed
// Use Stop() to stop listening and close the listener.
func (capsrv *CapServer) Start(lis net.Listener) error {
	logrus.Infof("CapServer listening on %s", lis.Addr())
	//err := rpc.Serve(lis, capnp.Client(capsrv.capProviderCapability))
	capsrv.lis = lis
	err := caphelp.Serve(lis, capnp.Client(capsrv.capProviderCapability), nil)
	return err
}

// Stop listening
//func (capsrv *CapServer) Stop() {
//	_ = capsrv.lis.Close()
//}

// NewCapServer injects the ListCapabilities method in the given list of methods in addition to
// serving the given methods from the service.
//
// Next, use ExportCapabilities to make capabilities available to clients, followed by
// Start() to start listening for incoming connections.
// Note that ExportCapabilities only affects the capabilities that are available through
// the resolver. Anyone with access to the service socket will have access to all its capabilities.
//
// This returns a capnp server that can handle all given methods.
func NewCapServer(serviceName string, methods []server.Method) *CapServer {

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

	// create a new capnp server instance that serves all capnp methods including ListCapabilities
	allMethods := hubapi.CapProvider_Methods(methods, srv)
	clientHook := server.New(allMethods, srv, c)

	// turn it into a capability client for use by the server
	providerClient := capnp.NewClient(clientHook)
	// cast the client to the provider for things like AddRef and Release
	srv.capProviderCapability = hubapi.CapProvider(providerClient)

	return srv
}
