package dummy

import (
	"context"
	"errors"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/testenv"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
	"github.com/hiveot/hub/pkg/resolver"
	capnpserver2 "github.com/hiveot/hub/pkg/resolver/capnpserver"
	service2 "github.com/hiveot/hub/pkg/resolver/service"
	"net"
	"os"
	"path"
	"time"
)

var testServiceSocketPath = path.Join(testSocketDir, "testService.socket")

const testSocketDir = "/tmp/test-gateway"
const testClientID = "client1"
const testUseCapnp = true
const testUseWS = false //true

// DummyGateway for testing gateway clients
// This starts a real gateway service instance with a resolver service.
// Use 'AddCapability' to add a capability that the gateway can provide.
type DummyGateway struct {
	url          string
	lis          net.Listener
	stopResolver func()
	resClient    resolver.IResolverService
}

// AddCapability adds a capability for testing of clients
// If a capability is requested of the given name, a client to this interface
// will be returned.
func (dummy *DummyGateway) AddCapability(capName string, capApi interface{}) {

}

/* resolver handling of capabilities
addCapability(name,capIF,acl) where cap is either direct interface or a rpc interface using capnp or other protocol
capIF = getCapability(name) return the registered interface or rpc interface
capIF.methodName(params)

1. separate serializer from capability rpc
2. RPC invocation:
     A: invoke(name, msg)
	 B: invoke(name, interface{}, ...) - introspect to determine serializer
     C: use serializer:  funcName(a,b,c) -> {msg=serialize abc; invoke(funcName, msg)}
	 D: session = getCapability;  session.invoke(funcName,msg)
     E: session.Invoke( name, serialize(arg1,arg2)) (session, err)
3. combine 1+2
     cap ICapabilityXyz = session.getCapability(name)

Issues to resolve:
1. major - params serialization is rpc dependent
2. major - how does resolver proxy a cap? method and params
3. minor - access control to cap
4. minor - capIF must have a language specific definition
5. minor - resilience of capIF session lifecycle vs connection lifecycle

local test setup:
1. write service
2. register service capability with local resolver
3. get capability from local resolver

resolver service setup
1. connect local resolver to resolver service
2. register available capabilities (ID, protocol)
3. get capability from local resolver
4. if local, return capability, else ...
5. local resolver requests capability from resolver service for protocol
6. resolver service returns session for capability


*/

// Start a test gateway service with a dummy resolver
// This returns the gateway listening address:port
func (dummy *DummyGateway) Start(certs testenv.TestCerts) (addr string, err error) {
	_ = os.RemoveAll(testSocketDir)
	_ = os.MkdirAll(testSocketDir, 0700)

	var authnDummy authn.IAuthnService = NewDummyAuthnService()

	// the gateway uses the resolver to connect to services
	resolverSocketPath := path.Join(testSocketDir, resolver.ServiceName+".socket")
	dummy.stopResolver = dummy.startResolver(resolverSocketPath)

	// start the gateway service using the resolver
	svc := service.NewGatewayService(resolverSocketPath, authnDummy)
	err = svc.Start()
	if err != nil {
		panic(err)
	}

	// start the gateway capnp server to handle requests
	srvListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		err = errors.New("Unable to create a listener, can't run test:" + err.Error())
		return "", err
	}
	dummy.lis = listener.CreateTLSListener(srvListener, certs.ServerCert, certs.CaCert)
	go capnpserver.StartGatewayCapnpServer(svc, dummy.lis, "")

	time.Sleep(time.Millisecond)
	return srvListener.Addr().String(), nil
}

// start resolver
func (dummy *DummyGateway) startResolver(resolverSocketPath string) (stopfn func()) {
	_ = os.RemoveAll(resolverSocketPath)
	svc := service2.NewResolverService(testSocketDir)
	err := svc.Start(context.Background())
	if err != nil {
		panic("can't start resolver: " + err.Error())
	}
	lis, err := net.Listen("unix", resolverSocketPath)
	if err != nil {
		panic("need resolver")
	}
	go capnpserver2.StartResolverServiceCapnpServer(svc, lis, svc.HandleUnknownMethod)

	return func() {
		_ = lis.Close()
		_ = svc.Stop()
	}
}

// Stop listening
func (dummy *DummyGateway) Stop() {
	if dummy.stopResolver != nil {
		dummy.stopResolver()
	}
	if dummy.lis != nil {
		dummy.lis.Close()
	}
}

func NewDummyGateway(resClient resolver.IResolverService) *DummyGateway {
	dummy := &DummyGateway{resClient: resClient}
	return dummy
}
