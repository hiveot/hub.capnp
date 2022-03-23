package thing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

const zone = "test"

func TestCreateTD(t *testing.T) {
	deviceID := zone + "Thing1"
	thing := thing.CreateTD(deviceID, "test TD", vocab.DeviceTypeSensor)
	assert.NotNil(t, thing)

	// Set version
	versions := map[string]string{"Software": "v10.1", "Hardware": "v2.0"}
	thing.SetThingVersion(thing, versions)

	// Define TD property
	prop := thing.CreateProperty("Prop1", "First property", vocab.PropertyTypeOutput)
	enumValues := make([]string, 0) //{"value1", "value2"}
	thing.SetPropertyEnum(prop, enumValues)
	thing.SetPropertyUnit(prop, "C")
	thing.SetPropertyDataTypeInteger(prop, 1, 10)
	thing.SetPropertyDataTypeNumber(prop, 1, 10)
	thing.SetPropertyDataTypeString(prop, 1, 10)
	thing.SetPropertyDataTypeObject(prop, nil)
	thing.SetPropertyDataTypeArray(prop, 3, 10)
	thing.AddTDProperty(thing, "prop1", prop)
	// invalid prop should not blow up
	thing.AddTDProperty(thing, "prop2", nil)

	// Define event
	ev1 := thing.CreateTDEvent("ev1", "First event")
	thing.AddTDEvent(thing, "ev1", ev1)
	// invalid event should not blow up
	thing.AddTDEvent(thing, "ev1", nil)

	// Set error status
	thing.SetThingErrorStatus(thing, "there is an error")

	// Define action
	action1 := thing.CreateTDAction("setChannel", "Change the channel")
	actionProp := thing.CreateProperty("channel", "Select channel", "input")
	required := []string{"channel"}
	thing.SetTDActionInput(action1, "string", actionProp, required)
	thing.SetTDActionOutput(action1, "string")

	thing.AddTDAction(thing, "action1", action1)
	// invalid action should not blow up
	thing.AddTDAction(thing, "action1", nil)

	// Define form
	f1 := thing.CreateTDForm("form1", "", "application/json", "GET")
	formList := make([]map[string]interface{}, 0)
	formList = append(formList, f1)
	thing.SetTDForms(thing, formList)
	// invalid form should not blow up
	thing.SetTDForms(thing, nil)
}

func TestCreateAction(t *testing.T) {
	thing := thing.CreateTDAction("Action1", "Do stuff")
	assert.NotNil(t, thing)
}

func TestCreateActionRequest(t *testing.T) {
	param1 := map[string]interface{}{}
	thing := thing.CreateActionRequest("Action1", param1)
	assert.NotNil(t, thing)
}
