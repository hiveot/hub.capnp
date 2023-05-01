package pubsub_test

import (
	"context"
	"github.com/hiveot/hub/lib/hubclient"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/capnpclient"
	"github.com/hiveot/hub/pkg/pubsub/capnpserver"
	"github.com/hiveot/hub/pkg/pubsub/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testAddress = "/tmp/pubsub_test.socket"
const testUseCapnp = true

func startService(useCapnp bool) (pubsub.IPubSubService, func()) {
	svc := service.NewPubSubService()
	err := svc.Start()
	if err != nil {
		logrus.Panicf("not happy: %s", err)
	}
	if useCapnp {
		_ = os.Remove(testAddress)
		srvListener, err := net.Listen("unix", testAddress)
		if err != nil {
			logrus.Panic("Unable to create a listener, can't run test")
		}
		go capnpserver.StartPubSubCapnpServer(svc, srvListener)

		// connect the client to the server above
		capClient, _ := hubclient.ConnectWithCapnpUDS("", testAddress)
		pubsubClient := capnpclient.NewPubSubCapnpClient(capClient)

		return pubsubClient, func() {
			pubsubClient.Release()
			_ = srvListener.Close()
			// allow ongoing releases to finish
			time.Sleep(time.Millisecond * 1)
			// catch missing releases
			_ = svc.Stop()
		}
	}
	return svc, func() {
		_ = svc.Stop()
	}
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	svc, stopFn := startService(testUseCapnp)
	_ = svc

	stopFn()
}

func TestPubSubAction(t *testing.T) {
	const publisherID = "urn:device1"
	const service1ID = "urn:service1"
	const thing1ID = "urn:thing1"
	const actionName1 = "action1"
	const actionName2 = "action2"
	var deviceAction = 0
	var serviceAction = 0
	var wildcardAction = 0

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)
	defer stopFn()

	devicePS, _ := svc.CapDevicePubSub(ctx, publisherID)
	servicePS, _ := svc.CapServicePubSub(ctx, service1ID)

	// test subscription of a single action by both service and device
	err := devicePS.SubAction(ctx, thing1ID, actionName1, func(val *thing.ThingValue) {
		deviceAction++
	})
	assert.NoError(t, err)
	err = servicePS.SubActions(ctx, publisherID, thing1ID, actionName1, func(val *thing.ThingValue) {
		serviceAction++
	})
	action1Msg := []byte("action1")
	err = servicePS.PubAction(ctx, publisherID, thing1ID, actionName1, action1Msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, deviceAction)
	assert.Equal(t, 1, serviceAction)

	// test subscription of a wildcards action
	err = devicePS.SubAction(ctx, "+", "+", func(val *thing.ThingValue) {
		wildcardAction++
	})
	assert.NoError(t, err)
	action2Msg := []byte("more of action")
	err = servicePS.PubAction(ctx, publisherID, thing1ID, actionName2, action2Msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, deviceAction)
	assert.Equal(t, 1, serviceAction)
	assert.Equal(t, 1, wildcardAction)

	devicePS.Release()
	servicePS.Release()
}

func TestPubSubEvent(t *testing.T) {
	const publisher1ID = "urn:device1"
	const thing1ID = "urn:thing1"
	const user1ID = "urn:user"
	const event1Name = "event1"
	var event1Count = int32(0)

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)
	defer stopFn()

	devicePS, _ := svc.CapDevicePubSub(ctx, publisher1ID)
	userPS, _ := svc.CapUserPubSub(ctx, user1ID)

	// test subscription of a single event by both service and device
	err := userPS.SubEvent(ctx, publisher1ID, thing1ID, event1Name, func(val *thing.ThingValue) {
		atomic.AddInt32(&event1Count, 1)
	})
	assert.NoError(t, err)
	event1Msg := []byte("event one")
	err = devicePS.PubEvent(ctx, thing1ID, event1Name, event1Msg)
	time.Sleep(time.Millisecond)
	assert.NoError(t, err)
	count := atomic.LoadInt32(&event1Count)
	assert.Equal(t, int32(1), count)

	devicePS.Release()
	userPS.Release()
	assert.NoError(t, err)
}
