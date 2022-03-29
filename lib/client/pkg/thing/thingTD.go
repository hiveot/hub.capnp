package thing

import (
	"encoding/json"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"time"
)

// ThingTD contains the Thing Description document
// Its structure is:
// {
//      @context: "http://www.w3.org/ns/td",
//      @type: <deviceType>,
//      id: <thingID>,
//      title: <human description>,  (why is this not a property?)
//      modified: <iso8601>,
//      actions: {name: ActionAffordance, ...},
//      events:  {name: EventAffordance, ...},
//      properties: {name: PropertyAffordance, ...}
// }
//
type ThingTD struct {
	// JSON-LD keyword to define short-hand names called terms that are used throughout a TD document. Required.
	AtContext []string `json:"@context"`

	// JSON-LD keyword to label the object with semantic tags (or types).
	AtType  string `json:"@type,omitempty"`
	AtTypes string `json:"@types,omitempty"`

	// Identifier of the Thing in form of a URI (RFC3986)
	// Optional in WoT but required in WoST in order to reach the device or service
	ID string `json:"id"`

	// Human-readable title in the default language. Required.
	Title string `json:"title"`
	// Human-readable titles in the different languages
	Titles map[string]string `json:"titles,omitempty"`

	// Provides additional (human-readable) information based on a default language
	Description string `json:"description,omitempty"`
	// Provides additional nulti-language information
	Descriptions []string `json:"descriptions,omitempty"`

	// Version information of the TD document (?not the device??)
	//Version VersionInfo `json:"version,omitempty"` // todo

	// ISO8601 timestamp this document was first created
	Created string `json:"created,omitempty"`
	// ISO8601 timestamp this document was last modified
	Modified string `json:"modified,omitempty"`

	// Information about the TD maintainer as URI scheme (e.g., mailto [RFC6068], tel [RFC3966], https).
	Support string `json:"support,omitempty"`

	// base: Define the base URI that is used for all relative URI references throughout a TD document.
	Base string `json:"base,omitempty"`

	// All properties-based interaction affordances of the thing
	Properties map[string]*PropertyAffordance `json:"properties,omitempty"`
	// All action-based interaction affordances of the thing
	Actions map[string]*ActionAffordance `json:"actions,omitempty"`
	// All event-based interaction affordances of the thing
	Events map[string]*EventAffordance `json:"events,omitempty"`

	// links: todo

	// Form hypermedia controls to describe how an operation can be performed. Forms are serializations of
	// Protocol Bindings. Thing-level forms are used to describe endpoints for a group of interaction affordances.
	Forms []Form `json:"forms,omitempty"`

	// Set of security definition names, chosen from those defined in securityDefinitions
	// In WoST security is handled by the Hub. WoST Things will use the NoSecurityScheme type
	Security string `json:"security"`
	// Set of named security configurations (definitions only).
	// Not actually applied unless names are used in a security name-value pair. (why is this mandatory then?)
	SecurityDefinitions map[string]string `json:"securityDefinitions,omitempty"`

	// profile: todo
	// schemaDefinitions: todo
	// uriVariables: todo
}

// AddProperty provides a simple way to add a property to the TD
// This returns the property affordance that can be augmented/modified directly
// By default the property is a read-only attribute.
//
// name is the name under which it is stored in the property affordance map. Any existing name will be replaced.
// title is the title used in the property. It is okay to use name if not sure.
// dataType is the type of data the property holds, WoTDataTypeNumber, ..Object, ..Array, ..String, ..Integer, ..Boolean or null
func (tdoc *ThingTD) AddProperty(name string, title string, dataType string) *PropertyAffordance {
	prop := &PropertyAffordance{
		DataSchema: DataSchema{
			Title:    title,
			Type:     dataType,
			ReadOnly: true,
		},
	}
	tdoc.UpdateProperty(name, prop)
	return prop
}

// AsMap returns the TD document as a map
func (tdoc *ThingTD) AsMap() map[string]interface{} {
	var asMap map[string]interface{}
	asJSON, _ := json.Marshal(tdoc)
	json.Unmarshal(asJSON, &asMap)
	return asMap
}

// tbd json-ld parsers:
// Most popular; https://github.com/xeipuuv/gojsonschema
// Other:  https://github.com/piprate/json-gold

// GetAction returns the action affordance with schema for the action.
// Returns nil if name is not an action or no affordance is defined.
func (tdoc *ThingTD) GetAction(name string) *ActionAffordance {
	actionAffordance, found := tdoc.Actions[name]
	if !found {
		return nil
	}
	return actionAffordance
}

// GetEvent returns the schema for the event or nil if the event doesn't exist
func (tdoc *ThingTD) GetEvent(name string) *EventAffordance {
	eventAffordance, found := tdoc.Events[name]
	if !found {
		return nil
	}
	return eventAffordance
}

// GetProperty returns the schema and value for the property or nil if name is not a property
func (tdoc *ThingTD) GetProperty(name string) *PropertyAffordance {
	propAffordance, found := tdoc.Properties[name]
	if !found {
		return nil
	}
	return propAffordance
}

// GetID returns the ID of the thing TD
func (tdoc *ThingTD) GetID() string {
	return tdoc.ID
}

// UpdateAction adds a new or replaces an existing action affordance (schema) of name. Intended for creating TDs
// Use UpdateProperty if name is a property name.
// Returns the added affordance to support chaining
func (tdoc *ThingTD) UpdateAction(name string, affordance *ActionAffordance) *ActionAffordance {
	tdoc.Actions[name] = affordance
	return affordance
}

// UpdateEvent adds a new or replaces an existing event affordance (schema) of name. Intended for creating TDs
// Returns the added affordance to support chaining
func (tdoc *ThingTD) UpdateEvent(name string, affordance *EventAffordance) *EventAffordance {
	tdoc.Events[name] = affordance
	return affordance
}

// UpdateForms sets the top level forms section of the TD
// NOTE: In WoST actions are always routed via the Hub using the Hub's protocol binding.
// Under normal circumstances forms are therefore not needed.
func (tdoc *ThingTD) UpdateForms(formList []Form) {
	tdoc.Forms = formList
}

// UpdateProperty adds or replaces a property affordance in the TD. Intended for creating TDs
// Returns the added affordance to support chaining
func (tdoc *ThingTD) UpdateProperty(name string, affordance *PropertyAffordance) *PropertyAffordance {
	tdoc.Properties[name] = affordance
	return affordance
}

// UpdateTitleDescription sets the title and description of the Thing in the default language
func (tdoc *ThingTD) UpdateTitleDescription(title string, description string) {
	tdoc.Title = title
	tdoc.Description = description
}

//// UpdateStatus sets the status property of a Thing
//// The status property is an object that holds possible status values
//// For example, an error status can be set using the 'error' field of the status property
//func (tdoc *ThingTD) UpdateStatus(statusName string, value string) {
//	sprop := tdoc.GetProperty("status")
//	if sprop == nil {
//		sprop = &PropertyAffordance{}
//		sprop.Title = "Status"
//		sprop.Description = "Device status info"
//		sprop.Type = vocab.WoTDataTypeObject
//	}
//	tdoc.UpdatePropertyValue("status", errorStatus)
//	// FIXME:is this a property
//	status := td["status"]
//	if status == nil {
//		status = make(map[string]interface{})
//		td["status"] = status
//	}
//	status.(map[string]interface{})["error"] = errorStatus
//}

// CreateTD creates a new Thing Description document with properties, events and actions
// Its structure:
// {
//      @context: "http://www.w3.org/ns/td",
//      id: <thingID>,      		// required in WoST. See CreateThingID for recommended format
//      title: string,              // required. Human description of the thing
//      @type: <deviceType>,        // required in WoST. See WoST DeviceType vocabulary
//      created: <iso8601>,         // will be the current timestamp. See vocabulary TimeFormat
//      actions: {name:TDAction, ...},
//      events:  {name: TDEvent, ...},
//      properties: {name: TDProperty, ...}
// }
func CreateTD(thingID string, title string, deviceType vocab.DeviceType) *ThingTD {
	td := ThingTD{
		AtContext:  []string{"http://www.w3.org/ns/thing"},
		Actions:    map[string]*ActionAffordance{},
		Created:    time.Now().Format(vocab.TimeFormat),
		Events:     map[string]*EventAffordance{},
		Forms:      nil,
		ID:         thingID,
		Modified:   time.Now().Format(vocab.TimeFormat),
		Properties: map[string]*PropertyAffordance{},
		// security schemas don't apply to WoST devices, except services exposed by the hub itself
		Security: vocab.WoTNoSecurityScheme,
		Title:    title,
	}

	// TODO @type is a JSON-LD keyword to label using semantic tags, eg it needs a schema
	if deviceType != "" {
		// deviceType must be a string for serialization and querying
		td.AtType = string(deviceType)
	}
	return &td
}
