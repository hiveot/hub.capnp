package td

import "github.com/wostzone/hub/lib/client/pkg/vocab"

// Thing property creation

// CreateProperty creates a new property instance
//  {
//     @type: propType,       // recommended; see vocab.PropertyTypeXyz
//     title: title,          // recommended; for human display
//     description: description, // optional; for additional display
//     writable: true,        // not defined in WoT. true for configuration properties
//     forms: [Form],         // WoT mandatory. unclear on how to use
//     observable: false,     // WoST properties are not directly observable (use message bus instead)
//  }
//  title for human presentation. Highly recommended.
//  description  of what the property does. recommended especially for configuration
//  propType provides '@type' value for a property
//  writable indicates the property is a configuration value and can be modified
func CreateProperty(
	title string,
	description string,
	propType vocab.ThingPropType) map[string]interface{} {

	var writable = (propType == vocab.PropertyTypeConfig)
	prop := make(map[string]interface{})
	// propType must be a string for jsonpath query to succeed
	prop[vocab.WoTAtType] = string(propType)
	prop[vocab.WoTTitle] = title
	if description != "" {
		prop[vocab.WoTDescription] = description
	}
	// default is read-only
	if writable {
		prop["writable"] = writable //
		// see https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#example-29
		prop[vocab.WoTReadOnly] = !writable
		prop[vocab.WoTWriteOnly] = writable
	}
	return prop
}

// GetPropertyValue returns the value of a property in the TD
// If the property doesn't exist in the TD or doesn't contain a value attribute, then found returns false
func GetPropertyValue(thingTD map[string]interface{}, propName string) (value string, found bool) {
	props, found := thingTD[vocab.WoTProperties].(map[string]interface{})
	if !found || props == nil {
		return
	}
	propOfName, found := props[propName].(map[string]interface{})
	if !found || propOfName == nil {
		return // TD does not have this property, or it is not a map[string]interface[}
	}
	valueInterface, found := propOfName[vocab.AttrNameValue]
	if found && valueInterface != nil {
		value = valueInterface.(string)
	}
	return value, found
}

// SetPropertyEnum sets an enumerated list of valid values of a property
func SetPropertyEnum(prop map[string]interface{}, enumValues ...interface{}) {
	prop[vocab.WoTEnum] = enumValues
}

// SetPropertyDataTypeArray sets the property data type as an array (of ?)
// If maxItems is 0, both minItems and maxItems are ignored
//  minItems is the minimum nr of items required
//  maxItems sets the maximum nr of items required
func SetPropertyDataTypeArray(prop map[string]interface{}, minItems uint, maxItems uint) {
	prop[vocab.WoTDataType] = vocab.WoTDataTypeArray
	if maxItems > 0 {
		prop[vocab.WoTMinItems] = minItems
		prop[vocab.WoTMaxItems] = maxItems
	}
}

// SetPropertyDataTypeInteger sets the property data type as an integer
// If min and max are both 0, they are ignored
//  min is the minimum value
//  max sets the maximum value
func SetPropertyDataTypeInteger(prop map[string]interface{}, min int, max int) {
	prop[vocab.WoTDataType] = vocab.WoTDataTypeInteger
	if !(min == 0 && max == 0) {
		prop[vocab.WoTMinimum] = min
		prop[vocab.WoTMaximum] = max
	}
}

// SetPropertyDataTypeNumber sets the property data type as floating point number
// If min and max are both 0, they are ignored
//  min is the minimum value
//  max sets the maximum value
func SetPropertyDataTypeNumber(prop map[string]interface{}, min float64, max float64) {
	prop[vocab.WoTDataType] = vocab.WoTDataTypeNumber
	if !(min == 0 && max == 0) {
		prop[vocab.WoTMinimum] = min
		prop[vocab.WoTMaximum] = max
	}
}

// SetPropertyDataTypeObject sets the property data type as an object
func SetPropertyDataTypeObject(prop map[string]interface{}, object interface{}) {
	prop[vocab.WoTDataType] = vocab.WoTDataTypeObject
	prop[vocab.WoTDataTypeObject] = object
}

// SetPropertyDataTypeString sets the property data type as string
// If minLength and maxLength are both 0, they are ignored
//  minLength is the minimum value
//  maxLength sets the maximum value
func SetPropertyDataTypeString(prop map[string]interface{}, minLength int, maxLength int) {
	prop["type"] = vocab.WoTDataTypeString
	if !(minLength == 0 && maxLength == 0) {
		prop[vocab.WoTMinLength] = minLength
		prop[vocab.WoTMaxLength] = maxLength
	}
}

// SetPropertyUnit sets the unit of a property
func SetPropertyUnit(prop map[string]interface{}, unit string) {
	prop[vocab.WoTUnit] = unit
}

// SetPropertyValue sets the value of a property at the time of TD creation
// Useful for attributes or configuration properties that don't change very often. When a TD is received
// it can be usable immediately without waiting for value updates.
//
// Note1: it is recommended to only set values for properties that rarely change, and when
// they do change to update the TD.
// Note2: This is optional and not part of the WoT specification. It is however allowed
// by the specification (although it might need a schema specified???)
func SetPropertyValue(prop map[string]interface{}, value string) {
	prop[vocab.AttrNameValue] = value
}
