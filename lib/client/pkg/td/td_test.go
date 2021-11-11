package td_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/td"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

const zone = "test"

func TestCreateTD(t *testing.T) {
	deviceID := zone + "Thing1"
	thing := td.CreateTD(deviceID, vocab.DeviceTypeSensor)
	assert.NotNil(t, thing)

	// Set version
	versions := map[string]string{"Software": "v10.1", "Hardware": "v2.0"}
	td.SetThingVersion(thing, versions)

	// Define TD property
	prop := td.CreateProperty("Prop1", "First property", vocab.PropertyTypeSensor)
	enumValues := make([]string, 0) //{"value1", "value2"}
	td.SetPropertyEnum(prop, enumValues)
	td.SetPropertyUnit(prop, "C")
	td.SetPropertyDataTypeInteger(prop, 1, 10)
	td.SetPropertyDataTypeNumber(prop, 1, 10)
	td.SetPropertyDataTypeString(prop, 1, 10)
	td.SetPropertyDataTypeObject(prop, nil)
	td.SetPropertyDataTypeArray(prop, 3, 10)
	td.AddTDProperty(thing, "prop1", prop)
	// invalid prop should not blow up
	td.AddTDProperty(thing, "prop2", nil)

	// Define event
	ev1 := td.CreateTDEvent("ev1", "First event")
	td.AddTDEvent(thing, "ev1", ev1)
	// invalid event should not blow up
	td.AddTDEvent(thing, "ev1", nil)

	// Set error status
	td.SetThingErrorStatus(thing, "there is an error")

	// Define action
	action1 := td.CreateTDAction("setChannel", "Change the channel")
	actionProp := td.CreateProperty("channel", "Select channel", "input")
	required := []string{"channel"}
	td.SetTDActionInput(action1, "string", actionProp, required)
	td.SetTDActionOutput(action1, "string")

	td.AddTDAction(thing, "action1", action1)
	// invalid action should not blow up
	td.AddTDAction(thing, "action1", nil)

	// Define form
	f1 := td.CreateTDForm("form1", "", "application/json", "GET")
	formList := make([]map[string]interface{}, 0)
	formList = append(formList, f1)
	td.SetTDForms(thing, formList)
	// invalid form should not blow up
	td.SetTDForms(thing, nil)
}

func TestCreateAction(t *testing.T) {
	thing := td.CreateTDAction("Action1", "Do stuff")
	assert.NotNil(t, thing)
}

func TestCreateActionRequest(t *testing.T) {
	param1 := map[string]interface{}{}
	thing := td.CreateActionRequest("Action1", param1)
	assert.NotNil(t, thing)
}
