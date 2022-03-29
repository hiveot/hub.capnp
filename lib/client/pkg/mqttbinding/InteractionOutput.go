// Package mqttbinding with handling of property, event and action values
package mqttbinding

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

// InteractionOutput to expose the data returned from WoT Interactions to applications.
// Use NewInteractionOutput to initialize
type InteractionOutput struct {
	// schema describing the data from property, event or action affordance
	schema *thing.DataSchema
	// raw data from the interaction as described by the Schema
	jsonEncoded []byte
	// decoded data in their native format, eg string, int, array, object
	value interface{}
}

// Value returns the parsed value of the interaction
func (io *InteractionOutput) Value() interface{} {
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
	_ = json.Unmarshal(io.jsonEncoded, &obj)
	return obj
}

// ValueAsString returns the value as a string
func (io *InteractionOutput) ValueAsString() string {
	s := ""
	err := json.Unmarshal(io.jsonEncoded, &s)
	if err != nil {
		logrus.Errorf("ValueAsString: Can't convert value '%s' to a string", io.value)
	}
	return s
}

// ValueAsBoolean returns the value as a boolean
func (io *InteractionOutput) ValueAsBoolean() bool {
	b := false
	err := json.Unmarshal(io.jsonEncoded, &b)
	if err != nil {
		logrus.Errorf("ValueAsBoolean: Can't convert value '%s' to a boolean", io.value)
	}
	return b
}

// ValueAsInt returns the value as an integer
func (io *InteractionOutput) ValueAsInt() int {
	i := 0
	err := json.Unmarshal(io.jsonEncoded, &i)
	if err != nil {
		logrus.Errorf("ValueAsInt: Can't convert value '%s' to a int", io.value)
	}
	return i
}

// ValueAsMap returns the value as a key-value map
// Returns nil if no data was provided.
func (io *InteractionOutput) ValueAsMap() map[string]interface{} {
	o := make(map[string]interface{})
	err := json.Unmarshal(io.jsonEncoded, &o)
	if err != nil {
		logrus.Errorf("ValueAsMap: Can't convert value '%s' to a map", io.value)
	}
	return o
}

// NewInteractionOutputFromJson creates a new interaction output for reading output data
// jsonEncoded is raw data that will be json parsed using the given schema
// schema describes the value. nil in case of unknown schema
func NewInteractionOutputFromJson(jsonEncoded []byte, schema *thing.DataSchema) InteractionOutput {
	var err error
	var val interface{}
	if schema != nil && schema.Type == vocab.WoTDataTypeObject {
		// If this is an object use a map
		val := make(map[string]interface{})
		err = json.Unmarshal(jsonEncoded, &val)
	} else {
		var sVal interface{}
		err = json.Unmarshal(jsonEncoded, &sVal)
		if err == nil {
			val = sVal
		} else {
			// otherwise keep native type in its string format
			val = string(jsonEncoded)
		}
	}
	if err != nil {
		logrus.Errorf("NewInteractionOutputFromJson. Error unmarshalling data: '%s'", err)
	}
	io := InteractionOutput{
		jsonEncoded: jsonEncoded,
		schema:      schema,
		value:       val,
	}
	return io
}

// NewInteractionOutput creates a new interaction output from object data
// data is native that will be json encoded using the given schema
// schema describes the value. nil in case of unknown schema
func NewInteractionOutput(data interface{}, schema *thing.DataSchema) InteractionOutput {
	jsonEncoded, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("NewInteractionOutput. Unable to marshal data: '%s'", data)
	}
	io := InteractionOutput{
		jsonEncoded: jsonEncoded,
		schema:      schema,
		value:       data,
	}
	return io
}
