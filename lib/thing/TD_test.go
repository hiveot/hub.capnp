package thing_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hiveot/hub.capnp/go/vocab"
	"github.com/hiveot/hub/lib/thing"
)

func TestCreateTD(t *testing.T) {
	thingID := "urn:thing1"
	tdoc := thing.NewTD(thingID, "test TD", vocab.DeviceTypeSensor)
	assert.NotNil(t, tdoc)

	// Set version
	//versions := map[string]string{"Software": "v10.1", "Hardware": "v2.0"}
	propAffordance := &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type:  vocab.WoTDataTypeArray,
			Title: "version",
		},
	}
	tdoc.UpdateProperty(vocab.VocabSoftwareVersion, propAffordance)

	// Define TD property
	propAffordance = &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type: vocab.WoTDataTypeString,
			Enum: make([]interface{}, 0), //{"value1", "value2"},
			Unit: "C",
		},
	}

	// created time must be set to ISO8601
	assert.NotEmpty(t, tdoc.Created)
	t1, err := time.Parse(vocab.ISO8601Format, tdoc.Created)
	assert.NoError(t, err)
	assert.NotNil(t, t1)

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

func TestMissingAffordance(t *testing.T) {
	thingID := "urn:thing1"

	// test return nil if no affordance is found
	tdoc := thing.NewTD(thingID, "test TD", vocab.DeviceTypeSensor)
	assert.NotNil(t, tdoc)

	prop := tdoc.GetProperty("prop1")
	assert.Nil(t, prop)

	action := tdoc.GetAction("action1")
	assert.Nil(t, action)

	ev := tdoc.GetEvent("event1")
	assert.Nil(t, ev)
}

func TestAddProp(t *testing.T) {
	thingID := "urn:thing1"
	tdoc := thing.NewTD(thingID, "test TD", vocab.DeviceTypeSensor)
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
	thingID := "urn:thing1"
	tdoc := thing.NewTD(thingID, "test TD", vocab.DeviceTypeSensor)
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
	thingID := "urn:thing1"
	tdoc := thing.NewTD(thingID, "test TD", vocab.DeviceTypeSensor)
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
