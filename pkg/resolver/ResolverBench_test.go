package resolver_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/captest"
)

// BenchmarkRPC determines the time it takes for a direct and indirect call to the test service
func BenchmarkRPC(b *testing.B) {
	ctx := context.Background()
	const testServiceListenSocket = "/tmp/test-resolverbenchlisten.socket"

	// start the listening service
	svc, stopFn := startResolverAndClient(true)
	defer stopFn()

	// create a test server and register it with the resolver
	ts := captest.NewTestService()
	ts.Listen(testServiceListenSocket)

	err := ts.Start(testResolverSocket)
	assert.NoError(b, err)

	// obtain the test service capability for method1 via the resolver
	indirectMethod1Client, err2 := svc.GetCapability(
		ctx, "test", hubapi.ClientTypeService, "capMethod1", nil)
	require.NoError(b, err2)
	capMethod1 := captest.CapMethod1Service(indirectMethod1Client)

	// obtain the test service capability for method1 directly to the service
	clConn, err := net.DialTimeout("unix", testServiceListenSocket, time.Second)
	transport := rpc.NewStreamTransport(clConn)
	rpcConn := rpc.NewConn(transport, nil)
	capTestService := captest.CapTestService(rpcConn.Bootstrap(ctx))
	method, release := capTestService.CapMethod1(ctx, nil)
	defer release()
	resp, err := method.Struct()
	require.NoError(b, err)
	directCapability := resp.Capabilit()

	b.Run(fmt.Sprintf("Direct request"),
		func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				// invoke method 1
				method1, release := directCapability.Method1(ctx, nil)
				// get the result
				resp, err3 := method1.Struct()
				assert.NoError(b, err3)
				msg1, _ := resp.ForYou()
				assert.NotEmpty(b, msg1)
				release()
			}
		})

	b.Run(fmt.Sprintf("GetCapability via resolver"),
		func(b *testing.B) {

			for n := 0; n < b.N; n++ {
				// invoke method 1
				method1, release := capMethod1.Method1(ctx, nil)
				// get the result
				resp, err3 := method1.Struct()
				assert.NoError(b, err3)
				msg1, _ := resp.ForYou()
				assert.NotEmpty(b, msg1)
				release()
			}
		})

	capMethod1.Release()
	directCapability.Release()
}
