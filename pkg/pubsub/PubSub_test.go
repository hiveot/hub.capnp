package pubsub_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.capnp/go/vocab"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/pubsub/service"
)

func createService() *service.PubSubService {
	svc := service.NewPubSubService()
	return svc
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	ctx := context.Background()
	svc := createService()
	err := svc.Start(ctx)
	assert.NoError(t, err)

	err = svc.Stop(ctx)
	assert.NoError(t, err)
}

func TestMissingRelease(t *testing.T) {
	ctx := context.Background()
	svc := createService()
	err := svc.Start(ctx)
	assert.NoError(t, err)

	devicePS := svc.CapDevicePubSub(ctx, "deviceID")
	_ = devicePS

	// test subscription of a single action by both service and device
	err = devicePS.SubAction(ctx, "", "", func(val *thing.ThingValue) {
		// subscriber
	})
	// having unreleased subscription is an error
	err = svc.Stop(ctx)
	assert.Error(t, err)
}

func TestPubSubAction(t *testing.T) {
	const deviceID = "urn:device1"
	const serviceID = "urn:service1"
	const userID = "urn:user1"
	const thing1ID = "urn:thing1"
	const actionName1 = "action1"
	const actionName2 = "action2"
	var deviceAction = 0
	var serviceAction = 0
	var wildcardAction = 0

	ctx := context.Background()
	svc := createService()
	err := svc.Start(ctx)
	assert.NoError(t, err)

	devicePS := svc.CapDevicePubSub(ctx, deviceID)
	servicePS := svc.CapServicePubSub(ctx, serviceID)

	// test subscription of a single action by both service and device
	err = devicePS.SubAction(ctx, thing1ID, actionName1, func(val *thing.ThingValue) {
		deviceAction++
	})
	assert.NoError(t, err)
	err = servicePS.SubActions(ctx, deviceID, thing1ID, actionName1, func(val *thing.ThingValue) {
		serviceAction++
	})

	action1Msg := thing.ThingValue{
		GatewayID: deviceID,
		ThingID:   thing1ID,
		Name:      actionName1,
		ValueJSON: []byte("lots of action"),
		Created:   "",
	}
	err = servicePS.PubAction(ctx, &action1Msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, deviceAction)
	assert.Equal(t, 1, serviceAction)

	// test subscription of a wildcards action
	err = devicePS.SubAction(ctx, "+", "+", func(val *thing.ThingValue) {
		wildcardAction++
	})
	assert.NoError(t, err)
	action2Msg := thing.ThingValue{
		GatewayID: deviceID,
		ThingID:   thing1ID,
		Name:      actionName2,
		ValueJSON: []byte("more of action"),
		Created:   "",
	}
	err = servicePS.PubAction(ctx, &action2Msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, deviceAction)
	assert.Equal(t, 1, serviceAction)
	assert.Equal(t, 1, wildcardAction)

	devicePS.Release()
	servicePS.Release()
	err = svc.Stop(ctx)
	assert.NoError(t, err)
}

func TestPubSubEvent(t *testing.T) {
	const device1ID = "urn:device1"
	const service1ID = "urn:service1"
	const thing1ID = "urn:thing1"
	const user1ID = "urn:user"
	const event1Name = "event1"
	var event1Count = 0

	ctx := context.Background()
	svc := createService()
	err := svc.Start(ctx)
	assert.NoError(t, err)

	devicePS := svc.CapDevicePubSub(ctx, device1ID)
	userPS := svc.CapUserPubSub(ctx, user1ID)

	// test subscription of a single event by both service and device
	err = userPS.SubEvent(ctx, device1ID, thing1ID, event1Name, func(val *thing.ThingValue) {
		event1Count++
	})
	assert.NoError(t, err)
	event1Msg := thing.ThingValue{
		GatewayID: device1ID,
		ThingID:   thing1ID,
		Name:      event1Name,
		ValueJSON: []byte("event one"),
		Created:   "",
	}
	err = devicePS.PubEvent(ctx, &event1Msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, event1Count)

	devicePS.Release()
	userPS.Release()
	err = svc.Stop(ctx)
	assert.NoError(t, err)

}

func TestPubSubTD(t *testing.T) {
	const deviceID = "urn:device1"
	const serviceID = "urn:service1"
	const userID = "urn:user1"
	const thing1ID = "urn:thing1"
	var rxTD *thing.ThingValue

	ctx := context.Background()
	svc := createService()
	err := svc.Start(ctx)
	assert.NoError(t, err)

	devicePS := svc.CapDevicePubSub(ctx, deviceID)
	servicePS := svc.CapServicePubSub(ctx, serviceID)

	err = servicePS.SubTDs(ctx, deviceID, func(val *thing.ThingValue) {
		rxTD = val
	})
	assert.NoError(t, err)
	td1 := []byte("hi")
	err = devicePS.PubTD(ctx, thing1ID, vocab.DeviceTypeButton, td1)
	assert.NoError(t, err)
	require.NotNil(t, rxTD)
	assert.Equal(t, deviceID, rxTD.GatewayID)
	assert.Equal(t, thing1ID, rxTD.ThingID)
	assert.Equal(t, td1, rxTD.ValueJSON)

	devicePS.Release()
	servicePS.Release()
	err = svc.Stop(ctx)
	assert.NoError(t, err)
}

func TestPubSubProperties(t *testing.T) {
	const device1ID = "urn:device1"
	const thing1ID = "urn:thing1"
	const user1ID = "urn:user"
	const propName = "event1"
	var event1Count = 0
	var rxPropsEvent *thing.ThingValue
	var rxProps map[string]string

	ctx := context.Background()
	svc := createService()
	err := svc.Start(ctx)
	assert.NoError(t, err)

	devicePS := svc.CapDevicePubSub(ctx, device1ID)
	userPS := svc.CapUserPubSub(ctx, user1ID)

	// test subscription of a single event by both service and device
	err = userPS.SubEvent(ctx, device1ID, "", "properties",
		func(val *thing.ThingValue) {
			event1Count++
			rxPropsEvent = val
		})
	assert.NoError(t, err)
	propsIn := map[string]string{
		"prop1": "prop1value",
	}
	err = devicePS.PubProperties(ctx, thing1ID, propsIn)
	assert.NoError(t, err)
	assert.Equal(t, 1, event1Count)
	assert.Equal(t, device1ID, rxPropsEvent.GatewayID)
	assert.Equal(t, thing1ID, rxPropsEvent.ThingID)
	err = json.Unmarshal(rxPropsEvent.ValueJSON, &rxProps)
	assert.NoError(t, err)
	assert.Equal(t, "prop1value", rxProps["prop1"])

	devicePS.Release()
	userPS.Release()
	err = svc.Stop(ctx)
	assert.NoError(t, err)

}
