// Package thing with schema type definitions for the ExposedThing and ConsumedThing classes
// as described here: https://www.w3.org/TR/wot-thing-description/#sec-data-schema-vocabulary-definition
package thing

// DataSchema with metadata  that describes the data format used. It can be used for validation.
// based on https://www.w3.org/TR/wot-thing-description/#dataschema
type DataSchema struct {
	// JSON-LD keyword to label the object with semantic tags (or types)
	AtType string `json:"@type,omitempty"`
	// Provides a human-readable title in the default language
	Title string `json:"title,omitempty"`
	// Provides a multi-language human-readable titles
	Titles []string `json:"titles,omitempty"`
	// Provides additional (human-readable) information based on a default language
	Description string `json:"description,omitempty"`
	// Provides additional nulti-language information
	Descriptions []string `json:"descriptions,omitempty"`
	// Provides a constant value of any type as per data schema
	Const interface{} `json:"const,omitempty"`
	// Provides a default value of any type as per data schema
	Default interface{} `json:"default,omitempty"`
	// Unit as used in international science, engineering, and business.
	// See vocab UnitNameXyz for units in the WoST vocabulary
	Unit string `json:"unit,omitempty"`
	// OneOf provides constraint of data as one of the given data schemas
	OneOf []DataSchema `json:"oneOf,omitempty"`
	// Restricted set of values provided as an array.
	//  for example: ["option1", "option2"]
	Enum []interface{} `json:"enum,omitempty"`
	// Boolean value to indicate whether a property interaction / value is read-only (=true) or not (=false)
	// the value true implies read-only.
	ReadOnly bool `json:"readOnly,omitempty"`
	// Boolean value to indicate whether a property interaction / value is write-only (=true) or not (=false)
	// the value true implies writable, except when ReadOnly is true.
	WriteOnly bool `json:"writeOnly,omitempty"`
	// Allows validation based on a format pattern such as "date-time", "email", "uri", etc.
	// See vocab DataFormXyz "date-time", "email", "uri" (todo)
	Format string `json:"format,omitempty"`
	// Type provides JSON based data type,  one of object, array, string, number, integer, boolean or null
	Type string `json:"type,omitempty"`
}

// ArraySchema with metadata describing data of type Array.
// https://www.w3.org/TR/wot-thing-description/#arrayschema
type ArraySchema struct {
	// subclass of DataSchema
	DataSchema
	// Used to define the characteristics of an array.
	// Note that in golang a field cannot both be a single or an array of items.
	Items DataSchema `json:"items"`
	// Defines the minimum number of items that have to be in the array
	MinItems uint `json:"minItems,omitempty"`
	// Defines the maximum number of items that have to be in the array.
	MaxItems uint `json:"maxItems,omitempty"`
}

// BooleanSchema with metadata describing data of type boolean.
// This Subclass is indicated by the value boolean assigned to type in DataSchema instances.
type BooleanSchema struct {
	DataSchema
}

// NumberSchema with metadata describing data of type number.
// This Subclass is indicated by the value number assigned to type in DataSchema instances.
type NumberSchema struct {
	DataSchema
}

// IntegerSchema with metadata describing data of type integer.
// This Subclass is indicated by the value integer assigned to type in DataSchema instances.
type IntegerSchema struct {
	DataSchema
}

// ObjectSchema with metadata describing data of type object.
// This Subclass is indicated by the value object assigned to type in DataSchema instances.
type ObjectSchema struct {
	DataSchema
}

// StringSchema with metadata describing data of type string.
// This Subclass is indicated by the value string assigned to type in DataSchema instances.
type StringSchema struct {
	DataSchema
}

// NullSchema with metadata describing data of type null.
// This Subclass is indicated by the value null assigned to type in DataSchema instances.
// This Subclass describes only one acceptable value, namely null. It can be used as part of a oneOf declaration,
// where it is used to indicate, that the data can also be null.
type NullSchema struct {
	DataSchema
}
