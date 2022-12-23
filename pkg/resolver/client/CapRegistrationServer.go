package client

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// CapRegistrationServer implements the boilerplate for registration of capabilities
// by services.  This connects to the resolver service and implements the
// CapProvider callback, eg: the ListCapabilities and GetCapability methods.
//
// Connect() creates an outgoing connection to the resolver service to which it serves
// available capabilities. Therefore, this is a TCP client for the connection and a server for capabilities.
//
// It is intended to be used with a service's capnp server to expose its capabilities to the resolver.
/* for example, a service named 'MyService' integrates with the resolver as follows:
 ```
> myserviceCapnpServer.go:
// Capnp server for MyService embeds the CapRegistrationServer code
type MyServiceCapnpServer {
     svc: *MyPOGSService
     capReg *CapRegistrationServer
 }

// On startup, initialize the CapRegistrationServer and register exported capabilities
// Thats it, your service is reachable through the resolver and its clients.
func StartMyServiceCapnpServer( context.Context, svc *MyPOGSService) {
   capReg := NewCapRegistrationServer("serviceName", CapMyService_Methods(nil,capsrv))
   capsrv := &MyServiceCapnpServer{
     capReg: capReg,
     svc: svc,
   }
   // the export name must match that in the Xxx_Methods provided list
   capReg.SetKnownMethods(CapMyService_Methods(nil,capsrv))
   capReg.ExportCapability("name", []string{hubapi.ClientTypeService})
   // main is this service's capability
   main := hubapi.CapMyService_ServerToClient(capsrv)
   capReg.Start(socketPath)
}
*/
type CapRegistrationServer struct {
	exportedCapabilities map[string]resolver.CapabilityInfo

	// Known methods of this service initialized on start
	knownMethods map[string]server.Method

	// Name of the service offering the capabilities
	serviceName string

	// This provider server capability for use by the resolver
	capProviderCapability hubapi.CapProvider

	// The resolver client with the registration and provider capability
	// Closing this closes the connection.
	// Only set when 'Connect' is used.
	resolverClient *capnpclient.ResolverSessionCapnpClient
}

// ExportCapability adds the method to the list of exported capabilities and allows it to be
// returned in GetCapability. This should be called before invoking Start as start uses it to
// register the capabilities.
// This only stores the capabilities for retrieval later by ListCapabilities.
func (rcl *CapRegistrationServer) ExportCapability(methodName string, clientTypes []string) {
	_, found := rcl.knownMethods[methodName]
	if !found {
		err := fmt.Errorf("method '%s' is not a known method. Unable to enable it", methodName)
		logrus.Error(err)
		panic(err)
	}
	newCap := resolver.CapabilityInfo{
		CapabilityName: methodName,
		ClientTypes:    clientTypes,
		ServiceID:      rcl.serviceName,
	}
	rcl.exportedCapabilities[methodName] = newCap
}

// GetCapability provides the requested capability of this service by invoking the associated method.
// This returns an error if the capability is not found or not available to the client type
func (rcl *CapRegistrationServer) GetCapability(
	ctx context.Context, call hubapi.CapProvider_getCapability) (err error) {

	args := call.Args()
	clientID, _ := args.ClientID()
	capabilityName, _ := args.CapabilityName()
	clientType, _ := args.ClientType()
	methodArgsCapnp, _ := args.Args()
	methodArgs := caphelp.UnmarshalStringList(methodArgsCapnp)

	capInfo, found := rcl.exportedCapabilities[capabilityName]
	if !found {
		err = fmt.Errorf("capability '%s' not found", capabilityName)
		return err
	}
	logrus.Infof("clientID='%s', clientType='%s', capabilityName='%s', args='%v'",
		clientID, clientType, capabilityName, methodArgs)

	// validate the client type is allowed on this method
	allowedTypes := strings.Join(capInfo.ClientTypes, ",")
	isAllowed := clientType != "" && strings.Contains(allowedTypes, clientType)
	if !isAllowed {
		err = fmt.Errorf("denied: capability '%s' is not available to client '%s' of type '%s'",
			capabilityName, clientID, clientType)
		logrus.Warning(err)
		return err
	}
	// invoke the method to get the capability
	// the clientID and clientType arguments from this call are passed on to the capability request
	// and available in the method.
	// okay, not quite sure how this works but the results of the method are applied to 'call'
	// and returned by this method. The 'Capability' result doesn't need a matching name apparently as
	// the first result from the capability table in the message is used. Quite convenient.
	// TBD: Can this behavior be relied on in future versions of go-capnp?
	method, _ := rcl.knownMethods[capabilityName]
	mc := call.Call
	err = method.Impl(ctx, mc)

	return err
}

// ListCapabilities returns the list of registered capabilities
func (rcl *CapRegistrationServer) ListCapabilities(
	_ context.Context, call hubapi.CapProvider_listCapabilities) error {
	var err error

	infoList := make([]resolver.CapabilityInfo, 0, len(rcl.exportedCapabilities))
	for _, capInfo := range rcl.exportedCapabilities {
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

// Provider returns the capnp capability of the capability provider
func (rcl *CapRegistrationServer) Provider() hubapi.CapProvider {
	return rcl.capProviderCapability
}

// Release is invoked by capnp when a client is released
// FIXME: determine when this is called by capnp and what needs to be done
// for now release the resolver service and close the rpc connection
func (rcl *CapRegistrationServer) Release() {
	logrus.Infof("client of registration service is released")
}

// Start establishes an RPC connection with the Resolver Service, obtain the registration
// capability and register this service capabilities.
// Users must call ExportCapabilities before connecting.
//
// If no socket path is given, then use the default path.
func (rcl *CapRegistrationServer) Start(resolverSocket string) (err error) {
	ctx := context.Background()
	if resolverSocket == "" {
		resolverSocket = resolver.DefaultResolverPath
	}

	conn, err := net.DialTimeout("unix", resolverSocket, time.Second)
	if err != nil {
		return err
	}
	// keep the resolver client alive as capabilities use its RPC connection
	rcl.resolverClient, err = capnpclient.NewResolverSessionCapnpClient(ctx, conn)
	if err != nil {
		return err
	}

	// list capabilities
	capList := make([]resolver.CapabilityInfo, 0, len(rcl.exportedCapabilities))
	for _, capInfo := range rcl.exportedCapabilities {
		capList = append(capList, capInfo)
	}

	err = rcl.resolverClient.RegisterCapabilities(ctx, rcl.serviceName, capList, rcl.capProviderCapability)

	return err
}

// Stop the registration server and close the RPC connection
func (rcl *CapRegistrationServer) Stop() {
	logrus.Warning("service registration server is shutting down")
	rcl.capProviderCapability.Release()
	// stop the connection and the client, if set
	if rcl.resolverClient != nil {
		rcl.resolverClient.Release()
	}
}

// NewCapRegistrationServer creates a capnp server for integration with the resolver service.
// Use ExportCapability to allow invoking the given methods through GetCapability.
//
//	serviceName is the name of the service whose methods are made available
//	methods is the list of known methods generated by capnp.
func NewCapRegistrationServer(serviceName string, methods []server.Method) *CapRegistrationServer {
	srv := &CapRegistrationServer{
		exportedCapabilities: make(map[string]resolver.CapabilityInfo),
		knownMethods:         make(map[string]server.Method),
		serviceName:          serviceName,
	}
	// the capnp capability for providing capabilities
	srv.capProviderCapability = hubapi.CapProvider_ServerToClient(srv)
	// does state carry to the remote side?
	capnp.Client(srv.capProviderCapability).State().Metadata.Put("serviceName", serviceName)

	for _, method := range methods {
		srv.knownMethods[method.MethodName] = method
	}
	return srv
}
