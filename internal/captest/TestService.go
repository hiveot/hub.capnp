package captest

import (
	"context"
	"net"
	"os"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/resolver/client"
)

// TestService is the concrete implementation of test service using the resolver.
// This implements the capnp generated TestService_Server interface.
type TestService struct {
	// The capnp server used to register capabilities
	capRegSrv   *client.CapRegistrationServer
	serviceName string
	lis         net.Listener // listening socket for direct serving
}

// CapMethod1 returns the capability to call method1
func (ts *TestService) CapMethod1(_ context.Context, call CapTestService_capMethod1) error {
	args := call.Args()
	_ = args
	clientID, _ := args.ClientID()
	clientType, _ := args.ClientType()

	res, err := call.AllocResults()
	if err == nil {
		testMethod1 := NewMethod1ServiceCapnpServer(clientID, clientType)
		capMethod1 := CapMethod1Service_ServerToClient(testMethod1)
		// the name doesn't matter as the first capability returned is used
		_ = res.SetCapabilit(capMethod1.AddRef())
	}
	logrus.Info("TestService.method1 called :)")
	return err
}

// Listen for direct incoming connections
func (ts *TestService) Listen(listenSocket string) {
	err := os.Remove(listenSocket)
	ts.lis, err = net.Listen("unix", listenSocket)
	if err != nil {
		panic("failed to start listening ")
	}
	go func() {
		main := CapTestService_ServerToClient(ts)
		err = caphelp.Serve(ts.lis, capnp.Client(main), nil)
	}()
}

// Start connects the test service to the resolver and register its capabilities
func (ts *TestService) Start(resolverSocket string) (err error) {

	// export the capability for running method1
	// connect to the resolver service and register capabilities
	err = ts.capRegSrv.Start(resolverSocket)
	if err != nil {
		return err
	}
	return err
}

// Stop the test service and close its connection to the resolver
func (ts *TestService) Stop() {
	ts.capRegSrv.Stop()
	if ts.lis != nil {
		_ = ts.lis.Close()
	}
}

// NewTestService creates a new instance of the test service capnp server
func NewTestService() *TestService {
	ts := &TestService{
		serviceName: "testservice",
	}

	// register the exported methods of this service as available capabilities
	ts.capRegSrv = client.NewCapRegistrationServer(
		ts.serviceName, CapTestService_Methods(nil, ts))

	// export the methods that are available as capabilities
	ts.capRegSrv.ExportCapability("capMethod1",
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeIotDevice})

	return ts
}
