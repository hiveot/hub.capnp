package captest

import (
	"context"
	"net"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver/capprovider"
)

// TestService is the concrete implementation of test service using CapServer.
// It represents how services make their capabilities available to clients.
// This implements the capnp generated TestService_Server interface.
type TestService struct {
	// The capnp server used to register capabilities
	capServer    *capprovider.CapServer
	serviceName  string
	lis          net.Listener // listening socket for direct serving
	listenSocket string
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

func (ts *TestService) Stop() {
	//ts.capServer.Stop()
	ts.lis.Close()
	_ = os.Remove(ts.listenSocket)
}

// Start listening in the background. This returns an error if no listening socket
// can be opened.
func (ts *TestService) Start(listenSocket string) (err error) {
	logrus.Infof("listening on %s", listenSocket)
	_ = os.Remove(listenSocket)
	ts.listenSocket = listenSocket
	lis, err := net.Listen("unix", listenSocket)
	ts.lis = lis
	if err != nil {
		return err
	}
	go ts.capServer.Start(lis)
	return err
}

// NewTestService creates a new instance of the test service capnp server
func NewTestService() *TestService {
	ts := &TestService{
		serviceName: "testservice",
	}

	// The capability server provides the exported capabilities of this service
	ts.capServer = capprovider.NewCapServer(
		ts.serviceName, CapTestService_Methods(nil, ts))

	// export the methods that are available as capabilities
	ts.capServer.ExportCapability("capMethod1",
		[]string{hubapi.ClientTypeService, hubapi.ClientTypeIotDevice})

	return ts
}
