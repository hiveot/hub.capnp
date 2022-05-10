package thing_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

const zone = "test"

func TestCreateTD(t *testing.T) {
	thingID := thing.CreateThingID("", "thing1", vocab.DeviceTypeUnknown)
	tdoc := thing.CreateTD(thingID, "test TD", vocab.DeviceTypeSensor)
	assert.NotNil(t, tdoc)

	// Set version
	//versions := map[string]string{"Software": "v10.1", "Hardware": "v2.0"}
	propAffordance := &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type:  vocab.WoTDataTypeArray,
			Title: "version",
		},
	}
	tdoc.UpdateProperty(vocab.PropNameSoftwareVersion, propAffordance)

	// Define TD property
	propAffordance = &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type: vocab.WoTDataTypeString,
			Enum: make([]interface{}, 0), //{"value1", "value2"},
			Unit: "C",
		},
	}
	tdoc.UpdateProperty("prop1", propAffordance)
	prop := tdoc.GetProperty("prop1")
	assert.NotNil(t, prop)

	tdoc.UpdateTitleDescription("title", "description")

	tdoc.UpdateAction("action1", &thing.ActionAffordance{})
	action := tdoc.GetAction("action1")
	assert.NotNil(t, action)

	tdoc.UpdateEvent("event1", &thing.EventAffordance{})
	ev := tdoc.GetEvent("event1")
	assert.NotNil(t, ev)

	tdoc.UpdateForms([]thing.Form{})

	tid2 := tdoc.GetID()
	assert.Equal(t, thingID, tid2)

	asMap := tdoc.AsMap()
	assert.NotNil(t, asMap)
}

func TestCreateThingID(t *testing.T) {
	// A full ID with publisher
	thingID1 := thing.CreatePublisherID("a", "pub1", "device2", vocab.DeviceTypeButton)
	zone, pub, did, dtype := thing.SplitThingID(thingID1)
	assert.Equal(t, "a", zone)
	assert.Equal(t, "pub1", pub)
	assert.Equal(t, "device2", did)
	assert.Equal(t, vocab.DeviceTypeButton, dtype)

	// An ID without publisher
	thingID2 := thing.CreateThingID("", "device2", vocab.DeviceTypeUnknown)
	zone, pub, did, dtype = thing.SplitThingID(thingID2)
	assert.Equal(t, "local", zone)
	assert.Equal(t, "", pub)
	assert.Equal(t, "device2", did)
	assert.Equal(t, vocab.DeviceTypeUnknown, dtype)

	// An ID without zone or publisher
	thingID3 := "urn:device3:pushbutton"
	zone, pub, did, dtype = thing.SplitThingID(thingID3)
	assert.Equal(t, "", zone)
	assert.Equal(t, "", pub)
	assert.Equal(t, "device3", did)
	assert.Equal(t, "pushbutton", string(dtype))

	// An ID without zone, publisher or device type
	thingID4 := "urn:device4"
	zone, pub, did, dtype = thing.SplitThingID(thingID4)
	assert.Equal(t, "", zone)
	assert.Equal(t, "", pub)
	assert.Equal(t, "device4", did)
	assert.Equal(t, "", string(dtype))
}

func TestMissingAffordance(t *testing.T) {
	thingID := thing.CreateThingID("", "thing1", vocab.DeviceTypeUnknown)

	// test return nil if no affordance is found
	tdoc := thing.CreateTD(thingID, "test TD", vocab.DeviceTypeSensor)
	assert.NotNil(t, tdoc)

	prop := tdoc.GetProperty("prop1")
	assert.Nil(t, prop)

	action := tdoc.GetAction("action1")
	assert.Nil(t, action)

	ev := tdoc.GetEvent("event1")
	assert.Nil(t, ev)
}

func TestAddProp(t *testing.T) {
	thingID := thing.CreateThingID("", "thing1", vocab.DeviceTypeUnknown)
	tdoc := thing.CreateTD(thingID, "test TD", vocab.DeviceTypeSensor)
	tdoc.AddProperty("prop1", "test property", vocab.WoTDataTypeBool)

	go func() {
		tdoc.AddProperty("prop2", "test property2", vocab.WoTDataTypeString)
	}()

	prop := tdoc.GetProperty("prop1")
	assert.NotNil(t, prop)
	time.Sleep(time.Millisecond)
	prop = tdoc.GetProperty("prop2")
	assert.NotNil(t, prop)
}

func TestAddEvent(t *testing.T) {
	thingID := thing.CreateThingID("", "thing1", vocab.DeviceTypeUnknown)
	tdoc := thing.CreateTD(thingID, "test TD", vocab.DeviceTypeSensor)
	tdoc.AddEvent("event1", "test event", vocab.WoTDataTypeBool)

	go func() {
		tdoc.AddEvent("event2", "test event 2", vocab.WoTDataTypeBool)
	}()

	ev := tdoc.GetEvent("event1")
	assert.NotNil(t, ev)
	time.Sleep(time.Millisecond)
	ev = tdoc.GetEvent("event2")
	assert.NotNil(t, ev)
}

func TestAddAction(t *testing.T) {
	thingID := thing.CreateThingID("", "thing1", vocab.DeviceTypeUnknown)
	tdoc := thing.CreateTD(thingID, "test TD", vocab.DeviceTypeSensor)
	tdoc.AddAction("action1", "test action", vocab.WoTDataTypeBool)

	go func() {
		tdoc.AddAction("action2", "test action", vocab.WoTDataTypeBool)
	}()

	action := tdoc.GetAction("action1")
	assert.NotNil(t, action)
	time.Sleep(time.Millisecond)
	action = tdoc.GetAction("action2")
	assert.NotNil(t, action)
}
