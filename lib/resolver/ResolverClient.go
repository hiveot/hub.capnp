package resolver

import (
	"capnproto.org/go/capnp/v3"
	"context"
	"errors"
	"fmt"
	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
	"github.com/sirupsen/logrus"
	"reflect"
)

// CapabilityMarshaller holds a locally registered (de)serializer for capabilities.
// A capabilityName is defined as an interface with one or more methods.
// The deserializer handles serialization of all interface methods of the capabilityName and
// is intended to be used in conjuncation with an RPC transport.
//
// During program startup, the application registers a factory function for each deserializer
// that is available. When requested, the factory function creates a new instance of a
// deserializer and is provided with the protocol RPC transport. It returns an instance
// with the registered golang interface that is connected to the remote service.
//
// The primary protocol used is capnp (Capn'proto), which is a bi-directional session based
// protocol.
//
// Use of a serializer is only needed for communication with remote services. Locally registered
// services do not need a serializer.
type CapabilityMarshaller struct {
	// Capability name as defined as a POGS
	capabilityName string

	// factory method for generating the serializer given the RPC connection
	// * capnp: func(client capnp.Client) ICapabilitySpecificInterface
	factory interface{}
	// the protocol this serializer supports
	// * capnp: capn'proto
	protocol string
	// url of the remote service providing the capabilityName when not using the resolver connection
	// if empty, the connected resolver client is used.
	url string
}

// ServiceRegistration for local services.
// Local services can be invoked directly without use of a marshaller.
// Also useful for registering testing stubs.
type ServiceRegistration struct {
	// Capability name provided by the service
	capabilityName string
	// Local services already have a singleton instance
	service interface{} // local service instance
}

// RemoteCapability defines a discovered remote capability and the connection to be used
// These capabilities are provided by the remote service and thus require a connection
// before the can be obtained.
type RemoteCapability struct {
	// Capability name as defined as a POGS
	capabilityName string
	// url of the remote service providing the capabilityName when not using the resolver connection
	// if empty, the connected resolver client is used.
	url string
	// protocol to use for the remote connection.
	protocol string
}

// ResolverClient for registration and generating of client proxies
type ResolverClient struct {
	// Capability marshallers by capability name for remote capabilities
	marshallers map[string]CapabilityMarshaller

	// Locally registered services by capability name
	services map[string]ServiceRegistration

	// Registration of remote capabilities, to be used in conjunction with the marshallers
	remoteCapabilities map[string]RemoteCapability

	// capnp resolver connection, if connected
	//resolverConn *rpc.Conn
	// go interface to resolver
	resolverService resolver.IResolverService
	// capnp client of the resolver used to obtain offered capabilities
	resolverCapnp capnp.Client

	// connection URL to remote resolver or gateway, eg tcp://host:port
	//remoteURL string
	// result of ListCapabilities on the resolver service.
	resolverCapabilities []resolver.CapabilityInfo
	// device or service client certificate based authentication, if available
	//clientCert *tls.Certificate
	// CA that signed the server certificate to verify authenticity of the resolver service
	//caCert *x509.Certificate
	// optional login with user credentials if this is a user connecting to the resolver
	loginID  string
	password string
}

// ConnectToResolverService connects this client to a resolver service using capnp and discover available
// capabilities.
//
// This is useful to obtain remote capabilities. A client certificate can be used to authenticate as a service or
// a device. If a certificate is not available then login must be used to gain access to remote capabilities.
//
// Note 1: Any service can be a resolver when it listens and uses the capnp protocol.
// Note 2: A resolver service is not needed if the requested server was previously registered with a URL and protocol, eg
//
//	when its address is known a direct connection can be made.
//
// The full URL is that of a resolver server listener, or gateway listener:
// - "unix://path/to/socket"
// - "tcp://address:port"
// - "wss://address:port/path"
//
//	fullURL to the remote resolver or gateway service, using the capnp protocol
//	clientCert optional client certificate to identify as
//	caCert CA's certificate to verify remote service authenticity
//func (cl *ResolverClient) ConnectToResolverService(fullURL string, clientCert *tls.Certificate, caCert *x509.Certificate) error {
//	var err error
//	authType := hubapi.AuthTypeService
//	if clientCert == nil {
//		authType = hubapi.AuthTypeUser
//	}
//	_ = authType
//	cl.remoteURL = fullURL
//	cl.clientCert = clientCert
//	cl.caCert = caCert
//	conn, err := hubclient.ConnectWithTCP(fullURL, clientCert, caCert)
//	if err != nil {
//		return err
//	}
//	capConn, capClient := hubclient.ConnectWithCapnp(conn)
//
//	cl.resolverConn = capConn
//	cl.resolverCapnp = capClient
//	cl.resolverService = capnpclient.NewResolverCapnpClient(capClient)
//	// getting capabilities verifies the resolver service is reachable
//	cl.resolverCapabilities, err = cl.resolverService.ListCapabilities(context.Background(), authType)
//	return err
//}

// ConnectToResolverService links this client to a resolver service using the given capnp client capability
// Use ConnectWithTCP or ConnectWithCapnpWebsockets to obtain the capability.
//
// Note 1: Any service can be a resolver when it listens and uses the capnp protocol.
// Note 2: A resolver service is not needed if the requested server was previously registered with a URL and protocol, eg
//
// capClient is the resolver capability
func (cl *ResolverClient) ConnectToResolverService(capClient capnp.Client) error {
	var err error
	//capConn, capClient := hubclient.ConnectWithCapnp(conn)

	//cl.resolverConn = capConn
	cl.resolverCapnp = capClient
	cl.resolverService = capnpclient.NewResolverCapnpClient(capClient)
	// getting capabilities verifies the resolver service is reachable
	cl.resolverCapabilities, err = cl.resolverService.ListCapabilities(context.Background(), hubapi.AuthTypeService)
	return err
}

// GetCapability returns a new instance of a capability of the given name.
//
// This will use the following approaches:
//  1. Locally registered services are returned immediately
//  2. A marshaller must have been registered to continue
//  3. If the marshashaller has a URL, connect to the URL
//  4. Last, use the marshaller with the remote resolver service, if connected.
//     Currently the resolver service only supports the capnp protocol.
//
// It is recommended to use the global resolver.GetCapabilities method instead of this
// instance method as it uses generics for type safety.
//
//	name must be the capabilityName of the interface that is returned
//	returns the interface to the capability or nil if not available
func (cl *ResolverClient) GetCapability(name string) interface{} {
	// 1. local capabilities are ready for use
	svcReg, found := cl.services[name]
	if found {
		return svcReg.service
	}
	// 2. remote capabilities must have a marshaller
	marshallerReg, found := cl.marshallers[name]
	if !found {
		errMsg := fmt.Sprintf("Marshaller for capability %s not found", name)
		logrus.Error(errMsg)
		return nil //, errors.New(errMsg)
	}
	// 3. try a direct connection if the marshaller came with a URL
	if marshallerReg.url != "" {
		// TODO support connections with different protocols
		// TODO support client certificate authentication for TCP
		capClient, err := hubclient.ConnectWithCapnpTCP(marshallerReg.url, nil, nil)
		//u, err := url.Parse(marshallerReg.url)
		//conn, err := net.DialTimeout(u.Scheme, u.Host, time.Second)
		//rpcConn, cap := hubclient.ConnectWithCapnp(conn)
		// if the registered endpoint URL isn't reachable, then fail
		if err != nil {
			logrus.Error(err)
			return nil
		}
		params := []reflect.Value{reflect.ValueOf(capClient)}
		fValue := reflect.ValueOf(marshallerReg.factory)
		out := fValue.Call(params)[0]
		proxy := out.Interface()
		return proxy
	}

	// 4. last, attempt to connect using the capnp resolver service
	if marshallerReg.protocol == "capnp" {
		if cl.resolverCapnp.IsValid() {
			params := []reflect.Value{reflect.ValueOf(cl.resolverCapnp)}
			fValue := reflect.ValueOf(marshallerReg.factory)
			out := fValue.Call(params)[0]
			proxy := out.Interface()
			return proxy
		} else {
			err := errors.New("no connection to the resolver service")
			logrus.Error(err)
			return nil
		}
	} else {
		err := errors.New("unsupported : " + marshallerReg.protocol)
		logrus.Error(err)
	}
	return nil
}

// Login to the resolver with user credentials to obtain additional capabilities.
func (cl *ResolverClient) Login(userID, password string) error {
	return errors.New("not implemented")
}

// Logout removes the current credentials and capabilities from the resolver
func (cl *ResolverClient) Logout() {
}

// RegisterService registers a local service with the given capability
//
//	name is the service interface name and identifies the capability
//	capability is the instance of the capability that implements its interface
func (cl *ResolverClient) RegisterService(name string, capability interface{}) {
	reg := ServiceRegistration{
		capabilityName: name,
		service:        capability,
	}
	cl.services[name] = reg
}

// RegisterCapnpMarshaller registers a capability (de)serializer for use with capnp
//
// The given factory MUST have this signature:
// >for capnp:  func(client capnp.Client) interface{}
//
//		capabilityName must be the capabilityName of the capabilityName interface of the factory function
//		factory is the method that generates the proxy client
//	 url is optional location of capnp server, or use "" to use the default resolver
func (cl *ResolverClient) RegisterCapnpMarshaller(name string,
	factory interface{}, url string) *CapabilityMarshaller {
	reg := CapabilityMarshaller{
		capabilityName: name,
		factory:        factory,
		url:            url,
		protocol:       "capnp",
	}
	cl.marshallers[name] = reg
	return &reg
}

// NewResolverClient provide an instance of a capabilityName resolver
func NewResolverClient() *ResolverClient {
	res := &ResolverClient{
		marshallers:        make(map[string]CapabilityMarshaller),
		services:           make(map[string]ServiceRegistration),
		remoteCapabilities: make(map[string]RemoteCapability),
	}
	return res
}

//// ConnectToResolverService connects this resolver client to the resolver service for discovering additional capabilities
////
////	fullURL to the remote resolver or gateway service
////	clientCert optional client certificate to identify as
////	caCert CA's certificate to verify remote service authenticity
//func ConnectToResolverService(fullURL string, clientCert *tls.Certificate, caCert *x509.Certificate) error {
//	return Resolver.ConnectToResolverService(fullURL, clientCert, caCert)
//}

// ConnectToResolverService connects this resolver client to the resolver service for discovering additional capabilities
//
//	conn the network connection to use
func ConnectToResolverService(capClient capnp.Client) error {
	return Resolver.ConnectToResolverService(capClient)
}

// GetCapability obtains a new instance of the capabilityName for the current user
//
// This operates on the Resolver singleton instance. Authentication is required
// for most capabilities. See resolver.Login()
func GetCapability[T interface{}]() T {
	var typeofT = reflect.TypeOf((*T)(nil))
	var capName = typeofT.Elem().Name()

	c := Resolver.GetCapability(capName)
	if c == nil {
		var zero T
		return zero
	}
	return c.(T)
}

// Login provides the resolver with credentials needed to obtain capabilities
func Login(userID, password string) error {
	return Resolver.Login(userID, password)
}

// Logout removes the current credentials and capabilities from the resolver
func Logout() {
	Resolver.Logout()
}

// RegisterCapnpMarshaller register a marshaller for a capnp capability with interface T
// Used to connect to marshal RPC messages to a remote capability provider.
//
//		factory is a factory method to create the marshaller instance
//	 url contains the location of the remote service, or "" to use the resolver service
func RegisterCapnpMarshaller[T interface{}](factory interface{}, url string) *CapabilityMarshaller {
	var typeofT = reflect.TypeOf((*T)(nil))
	var capName = typeofT.Elem().Name()
	reg := Resolver.RegisterCapnpMarshaller(capName, factory, url)
	return reg
}

// RegisterService registers a local service that provide capability with interface T
// When a service capability is requested, its singleton instance is returned.
//
//	T is the type of the interface returned by the factory function.
//	capability is the local service instance
//
// Returns the capabilityName of the registered capabilityName
func RegisterService[T interface{}](capability T) string {

	var typeofT = reflect.TypeOf((*T)(nil))
	var capName = typeofT.Elem().Name()
	Resolver.RegisterService(capName, capability)
	return capName
}

// Resolver is the global resolver instance
var Resolver = NewResolverClient()
