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

	// start the listening service
	svc, stopFn := startResolverAndClient(true)
	defer stopFn()
	_ = svc

	// create a test server and register it with the resolver
	ts := captest.NewTestService()
	ts.Start(testServiceSocket)

	// obtain the test service capability for method1 via the resolver
	resConn, _ := net.Dial("unix", testResolverSocket)
	transport := rpc.NewStreamTransport(resConn)
	rpcConn := rpc.NewConn(transport, nil)
	capability := captest.CapTestService(rpcConn.Bootstrap(ctx))

	method, release := capability.CapMethod1(ctx,
		func(params captest.CapTestService_capMethod1_Params) error {
			err2 := params.SetClientID("benchrpc")
			assert.NoError(b, err2)
			_ = params.SetClientType(hubapi.ClientTypeService)
			return err2
		})
	defer release()
	indirectMethod1Client := method.Capabilit()
	capMethod1 := captest.CapMethod1Service(indirectMethod1Client)

	// obtain the test service capability for method1 directly to the service
	clConn2, err := net.DialTimeout("unix", testServiceSocket, time.Second)
	transport2 := rpc.NewStreamTransport(clConn2)
	rpcConn2 := rpc.NewConn(transport2, nil)
	capTestService2 := captest.CapTestService(rpcConn2.Bootstrap(ctx))
	method2, release2 := capTestService2.CapMethod1(ctx, nil)
	defer release2()
	resp2, err := method2.Struct()
	require.NoError(b, err)
	directCapability := resp2.Capabilit()

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
