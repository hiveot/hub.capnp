package td

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

type ThingTD map[string]interface{}

// tbd json-ld parsers:
// Most popular; https://github.com/xeipuuv/gojsonschema
// Other:  https://github.com/piprate/json-gold

// AddTDAction adds or replaces an action in the TD
//  td is a TD created with 'CreateTD'
//  name of action to add
//  action created with 'CreateAction'
func AddTDAction(td ThingTD, name string, action interface{}) {
	actions := td[vocab.WoTActions].(map[string]interface{})
	if action == nil {
		logrus.Errorf("Add action '%s' to TD. Action is nil", name)
	} else {
		actions[name] = action
	}
}

// AddTDEvent adds or replaces an event in the TD
//  td is a TD created with 'CreateTD'
//  name of action to add
//  event created with 'CreateEvent'
func AddTDEvent(td ThingTD, name string, event interface{}) {
	events := td[vocab.WoTEvents].(map[string]interface{})
	if event == nil {
		logrus.Errorf("Add event '%s' to TD. Event is nil.", name)
	} else {
		events[name] = event
	}
}

// AddTDProperty adds or replaces a property in the TD
//  td is a TD created with 'CreateTD'
//  name of property to add
//  property created with 'CreateProperty'
func AddTDProperty(td ThingTD, name string, property interface{}) {
	props := td[vocab.WoTProperties].(map[string]interface{})
	if property == nil {
		logrus.Errorf("Add property '%s' to TD. Propery is nil.", name)
	} else {
		props[name] = property
	}
}

// CreateThingID creates a ThingID from the zone it belongs to, the hardware device ID and device Type
// This creates a Thing ID: URN:zone:deviceID:deviceType.
//  zone is the name of the zone the device is part of
//  deviceID is the ID of the device to use as part of the Thing ID
func CreateThingID(zone string, deviceID string, deviceType vocab.DeviceType) string {
	thingID := fmt.Sprintf("urn:%s:%s:%s", zone, deviceID, deviceType)
	return thingID
}

// return the ID of the given thing TD
func GetID(td ThingTD) string {
	if td == nil {
		return ""
	}
	id := td["id"].(string)
	return id
}

// RemoveTDProperty removes a property from the TD.
func RemoveTDProperty(td ThingTD, name string) {
	props := td[vocab.WoTProperties].(map[string]interface{})
	if props == nil {
		logrus.Errorf("RemoveTDProperty: TD does not have any properties.")
		return
	}
	props[name] = nil

}

// SetThingVersion adds or replace Thing version info in the TD
//  td is a TD created with 'CreateTD'
//  version with map of 'name: version'
func SetThingVersion(td ThingTD, version map[string]string) {
	td[vocab.WoTVersion] = version
}

// SetThingTitle sets the title and description of the Thing in the TD
//  td is a TD created with 'CreateTD'
//  title of the Thing
//  description of the Thing
func SetThingDescription(td ThingTD, title string, description string) {
	td[vocab.WoTTitle] = title
	td[vocab.WoTDescription] = description
}

// SetThingErrorStatus sets the error status of a Thing
// This is set under the 'status' property, 'error' subproperty
//  td is a TD created with 'CreateTD'
//  status is a status tring
func SetThingErrorStatus(td ThingTD, errorStatus string) {
	// FIXME:is this a property
	status := td["status"]
	if status == nil {
		status = make(map[string]interface{})
		td["status"] = status
	}
	status.(map[string]interface{})["error"] = errorStatus
}

// SetTDForm sets the top level forms section of the TD
// NOTE: In WoST actions are always routed via the Hub using the Hub's protocol binding.
// Under normal circumstances forms are therefore not needed.
//  td to add form to
//  forms with list of forms to add. See also CreateForm to create a single form
func SetTDForms(td ThingTD, formList []map[string]interface{}) {
	td[vocab.WoTForms] = formList
}

// CreatePublisherThingID creates a globally unique Thing ID that includes the zone and publisher
// name where the Thing originates from. The publisher is especially useful where protocol
// bindings create thing IDs. In this case the publisher is the gateway used by the protocol binding
// or the PB itself.  See also SplitThingID.
//
// This creates a Thing ID: URN:zone:publisher:deviceID:deviceType
//  zone is the name of the zone the device is part of
//  publisher is the deviceID of the publisher of the thing.
//  deviceID is the ID of the device to use as part of the Thing ID
func CreatePublisherThingID(zone string, publisher string, deviceID string, deviceType vocab.DeviceType) string {
	thingID := fmt.Sprintf("urn:%s:%s:%s:%s", zone, publisher, deviceID, deviceType)
	return thingID
}

// CreateTD creates a new Thing Description document with properties, events and actions
// This creates a structure:
//   {
//      @context: "http://www.w3.org/ns/td",
//      id: <thingID>,
//      @type: <deviceType>,
//      created: <iso8601>,
//      actions: {},
//      events: {},
//      properties: {}
//   }
func CreateTD(thingID string, deviceType vocab.DeviceType) ThingTD {
	td := make(ThingTD)
	td[vocab.WoTAtContext] = "http://www.w3.org/ns/td"
	td[vocab.WoTID] = thingID
	// TODO @type is a JSON-LD keyword to label using semantic tags, eg it needs a schema
	if deviceType != "" {
		// deviceType must be a string for serialization and querying
		td[vocab.WoTAtType] = string(deviceType)
	}
	td[vocab.WoTCreated] = time.Now().Format(vocab.TimeFormat)
	td[vocab.WoTActions] = make(map[string]interface{})
	td[vocab.WoTEvents] = make(map[string]interface{})
	td[vocab.WoTProperties] = make(map[string]interface{})
	return td
}

// SplitThingID takes a ThingID and breaks it down into individual parts. Supported formats:
//  A thingID without anything specific: URN:deviceID
//  A thingID without zone: URN:deviceID:deviceType
//  A thingID without publisher: URN:zone:deviceID:deviceType.
//  A thingID with publisher: URN:zone:publisherID:deviceID:deviceType.
//  thingID is the multi-part identified of the thing and the device it is part of
func SplitThingID(thingID string) (
	zone string, publisherID string, deviceID string, deviceType vocab.DeviceType) {
	parts := strings.Split(thingID, ":")
	if len(parts) < 2 || strings.ToLower(parts[0]) != "urn" {
		// not a conventional thing ID
		return "", "", "", ""
	} else if len(parts) == 5 {
		zone = parts[1]
		publisherID = parts[2]
		deviceID = parts[3]
		deviceType = vocab.DeviceType(parts[4])
	} else if len(parts) == 4 {
		zone = parts[1]
		deviceID = parts[2]
		deviceType = vocab.DeviceType(parts[3])
	} else if len(parts) == 3 {
		deviceID = parts[1]
		deviceType = vocab.DeviceType(parts[2])
	} else if len(parts) == 2 {
		deviceID = parts[1]
	}
	return
}