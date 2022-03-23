// Package mqttbinding with handling of property, event and action values
package mqttbinding

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/thing"
)

// InteractionOutput to expose the data returned from WoT Interactions to applications.
// Use NewInteractionOutput to initialize
type InteractionOutput struct {
	// schema describing the data from property, event or action affordance
	schema *thing.DataSchema
	// raw data from the interaction as described by the Schema
	data []byte
	// parsed json raw data
	value map[string]interface{}
}

// Value returns the parsed value of the interaction
func (io *InteractionOutput) Value() map[string]interface{} {
	return io.value
}

// ValueAsArray returns the value as an array
// The result depends on the schema type
//  array: returns array of values as describe ni the schema
//  boolean: returns a single element true/false
//  bytes: return an array of bytes
//  int: returns a single element with integer
//  object: returns a single element with object
//  string: returns a single element with string
func (io *InteractionOutput) ValueAsArray() []interface{} {
	obj := make([]interface{}, 0)
	_ = json.Unmarshal(io.data, &obj)
	return obj
}

// ValueAsString returns the value as a string
func (io *InteractionOutput) ValueAsString() string {
	s := ""
	err := json.Unmarshal(io.data, &s)
	if err != nil {
		logrus.Errorf("ValueAsBoolean: Can't convert value '%s' to a string", io.value)
	}
	return s
}

// ValueAsBoolean returns the value as a boolean
func (io *InteractionOutput) ValueAsBoolean() bool {
	b := false
	err := json.Unmarshal(io.data, &b)
	if err != nil {
		logrus.Errorf("ValueAsBoolean: Can't convert value '%s' to a boolean", io.value)
	}
	return b
}

// ValueAsInt returns the value as an integer
func (io *InteractionOutput) ValueAsInt() int {
	i := 0
	err := json.Unmarshal(io.data, &i)
	if err != nil {
		logrus.Errorf("ValueAsBoolean: Can't convert value '%s' to a int", io.value)
	}
	return i
}

// ValueAsObject returns the value as an object
func (io *InteractionOutput) ValueAsObject() map[string]interface{} {
	o := make(map[string]interface{})
	err := json.Unmarshal(io.data, &o)
	if err != nil {
		logrus.Errorf("ValueAsBoolean: Can't convert value '%s' to a int", io.value)
	}
	return o
}

// NewInteractionOutput creates a new interaction output for reading output data
// value is stored as a map following the given schema
// schema describes the value. nil in case of unknown schema
func NewInteractionOutput(data []byte, schema *thing.DataSchema) InteractionOutput {
	io := InteractionOutput{
		data:   data,
		schema: schema,
	}
	return io
}
