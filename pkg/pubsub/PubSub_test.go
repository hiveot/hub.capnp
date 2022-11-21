package pubsub_test

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.capnp/go/vocab"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/capnpclient"
	"github.com/hiveot/hub/pkg/pubsub/capnpserver"
	"github.com/hiveot/hub/pkg/pubsub/service"
)

const testAddress = "/tmp/pubsub_test.socket"
const testUseCapnp = true

func startService(useCapnp bool) (pubsub.IPubSubService, func() error) {
	ctx := context.Background()
	svc, err := service.StartPubSubService()
	if err != nil {
		logrus.Panicf("not happy: %s", err)
	}
	if useCapnp {
		_ = os.Remove(testAddress)
		srvListener, err := net.Listen("unix", testAddress)
		if err != nil {
			logrus.Panic("Unable to create a listener, can't run test")
		}
		go capnpserver.StartPubSubCapnpServer(ctx, srvListener, svc)

		// connect the client to the server above
		clConn, _ := net.Dial("unix", testAddress)
		capClient, err := capnpclient.StartPubSubCapnpClient(ctx, clConn)

		return capClient, func() error {
			// allow ongoing releases to finish
			time.Sleep(time.Millisecond * 10)
			err = capClient.Release()
			err2 := svc.Release()
			if err == nil {
				err = err2
			}
			return err
		}
	}
	return svc, svc.Release
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	svc, stopFn := startService(testUseCapnp)
	_ = svc

	err := stopFn()
	assert.NoError(t, err)

}

func TestMissingRelease(t *testing.T) {
	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	devicePS := svc.CapDevicePubSub(ctx, "deviceID")
	_ = devicePS

	// test subscription of a single action by both service and device
	err := devicePS.SubAction(ctx, "", "", func(val *thing.ThingValue) {
		// subscriber
	})
	// having unreleased subscription is an error
	err = stopFn()
	// it seems to be a race. not clear why
	_ = err
	assert.Error(t, err, "all subscriptions were released. this is unexpected")
}

func TestPubSubAction(t *testing.T) {
	const device1ID = "urn:device1"
	const service1ID = "urn:service1"
	const userID = "urn:user1"
	const thing1ID = "urn:thing1"
	var thing1Addr = thing.MakeThingAddr(device1ID, thing1ID)
	const actionName1 = "action1"
	const actionName2 = "action2"
	var deviceAction = 0
	var serviceAction = 0
	var wildcardAction = 0

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	devicePS := svc.CapDevicePubSub(ctx, device1ID)
	servicePS := svc.CapServicePubSub(ctx, service1ID)

	// test subscription of a single action by both service and device
	err := devicePS.SubAction(ctx, thing1ID, actionName1, func(val *thing.ThingValue) {
		deviceAction++
	})
	assert.NoError(t, err)
	err = servicePS.SubActions(ctx, thing1Addr, actionName1, func(val *thing.ThingValue) {
		serviceAction++
	})
	action1Msg := []byte("action1")
	err = servicePS.PubAction(ctx, thing1Addr, actionName1, action1Msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, deviceAction)
	assert.Equal(t, 1, serviceAction)

	// test subscription of a wildcards action
	err = devicePS.SubAction(ctx, "+", "+", func(val *thing.ThingValue) {
		wildcardAction++
	})
	assert.NoError(t, err)
	action2Msg := []byte("more of action")
	err = servicePS.PubAction(ctx, thing1Addr, actionName2, action2Msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, deviceAction)
	assert.Equal(t, 1, serviceAction)
	assert.Equal(t, 1, wildcardAction)

	devicePS.Release()
	servicePS.Release()
	err = stopFn()
	assert.NoError(t, err)
}

func TestPubSubEvent(t *testing.T) {
	const device1ID = "urn:device1"
	const service1ID = "urn:service1"
	const thing1ID = "urn:thing1"
	const user1ID = "urn:user"
	const event1Name = "event1"
	var thing1Addr = thing.MakeThingAddr(device1ID, thing1ID)
	var event1Count = int32(0)

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	devicePS := svc.CapDevicePubSub(ctx, device1ID)
	userPS := svc.CapUserPubSub(ctx, user1ID)

	// test subscription of a single event by both service and device
	err := userPS.SubEvent(ctx, thing1Addr, event1Name, func(val *thing.ThingValue) {
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
	err = stopFn()
	assert.NoError(t, err)

}

func TestPubSubTD(t *testing.T) {
	const device1ID = "urn:device1"
	const serviceID = "urn:service1"
	const userID = "urn:user1"
	const thing1ID = "urn:thing1"
	var thing1Addr = thing.MakeThingAddr(device1ID, thing1ID)
	var rxTD *thing.ThingValue

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	devicePS := svc.CapDevicePubSub(ctx, device1ID)
	servicePS := svc.CapServicePubSub(ctx, serviceID)

	err := servicePS.SubTDs(ctx, func(val *thing.ThingValue) {
		rxTD = val
	})
	assert.NoError(t, err)
	td1 := []byte("hi")
	err = devicePS.PubTD(ctx, thing1ID, vocab.DeviceTypeButton, td1)
	assert.NoError(t, err)
	require.NotNil(t, rxTD)
	assert.Equal(t, thing1Addr, rxTD.ThingAddr)
	assert.Equal(t, td1, rxTD.ValueJSON)

	devicePS.Release()
	servicePS.Release()
	err = stopFn()
	assert.NoError(t, err)
}

func TestPubSubProperties(t *testing.T) {
	const device1ID = "urn:device1"
	const thing1ID = "urn:thing1"
	const user1ID = "urn:user"
	const propName = "event1"
	var thing1Addr = thing.MakeThingAddr(device1ID, thing1ID)
	var event1Count = 0
	var rxPropsEvent *thing.ThingValue
	var rxProps map[string][]byte

	ctx := context.Background()
	svc, stopFn := startService(testUseCapnp)

	devicePS := svc.CapDevicePubSub(ctx, device1ID)
	userPS := svc.CapUserPubSub(ctx, user1ID)

	// test subscription of a single event by both service and device
	err := userPS.SubEvent(ctx, thing1Addr, "properties",
		func(val *thing.ThingValue) {
			event1Count++
			rxPropsEvent = val
		})
	assert.NoError(t, err)
	propsIn := map[string][]byte{
		"prop1": []byte("prop1value"),
	}
	err = devicePS.PubProperties(ctx, thing1ID, propsIn)
	assert.NoError(t, err)
	assert.Equal(t, 1, event1Count)
	assert.Equal(t, thing1Addr, rxPropsEvent.ThingAddr)

	err = json.Unmarshal(rxPropsEvent.ValueJSON, &rxProps)
	assert.NoError(t, err)
	assert.Equal(t, "prop1value", string(rxProps["prop1"]))

	devicePS.Release()
	userPS.Release()
	err = stopFn()
	assert.NoError(t, err)

}
