// Package td with TD event creation
package td

import "github.com/wostzone/hub/lib/client/pkg/vocab"

// CreateTDEvent creates a new TD event description
//  title title for presentation
//  description optional extra description of what the event represents
func CreateTDEvent(title string, description string) map[string]interface{} {
	event := make(map[string]interface{}, 0)
	event[vocab.WoTTitle] = title
	event[vocab.WoTDescription] = description

	return event
}

// CreateThingEvent creates a new event for publication
//  name of the event
//  params optional parameters
func CreateThingEvent(name string, params map[string]interface{}) map[string]interface{} {
	event := make(map[string]interface{}, 0)
	event[name] = params
	return event
}
