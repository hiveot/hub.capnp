package td

// Thing property definition
// Credits: https://github.com/dravenk/webthing-go/blob/master/property.go

import (
	"encoding/json"
)

// Property Initialize the object.
//
// @param thing    Thing this property belongs to
// @param name     Name of the property
// @param value    Value object to hold the property value
// @param metadata Property metadata, i.e. type, description, unit, etc., as
//                 a Map
type Property struct {
	thing      *Thing
	name       string
	value      *Value
	hrefPrefix string
	href       string
	metadata   json.RawMessage
}

// PropertyObject A property object describes an attribute of a Thing and is indexed by a property id.
// See https://iot.mozilla.org/wot/#property-object
type PropertyObject struct {
	AtType      string      `json:"@type,omitempty"`
	Title       string      `json:"title,omitempty"`
	Type        string      `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	Unit        string      `json:"unit,omitempty"`
	ReadOnly    bool        `json:"readOnly,omitempty"`
	Minimum     json.Number `json:"minimum,omitempty"`
	Maximum     json.Number `json:"maximum,omitempty"`
	Links       []Link      `json:"links,omitempty"`
}
