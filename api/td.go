// Package api with WoT definitions in golang
// Credits: https://github.com/draggett/arena-webhub
package api

// ThingDescription defining a thing from the Web of Things
// See also [5.3.1.1 Thing](https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#thing)
type ThingDescription struct {
	Context []string `json:"@context"` // JSON-LD context: http://www.w3.org/ns/td
	// Type    []string `json:"@type,omitempty"`    // JSON-LD context

	ID          *string                  `json:"id,omitempty"`     // URI with unique identifier for the thing
	Title       string                   `json:"title"`            // human friendly name for the thing
	Titles      *map[string]string       `json:"titles,omitempty"` // Multi language titles
	Description *string                  `json:"description,omitempty"`
	Version     *string                  `json:"version,,omitempty"`
	Created     *string                  `json:"created,omitempty"`    // DateTime
	Modified    *string                  `json:"modified,omitempty"`   // DateTime
	Security    []string                 `json:"security"`             // ???
	Properties  map[string]ThingProperty `json:"properties,omitempty"` // thing property map
	Actions     map[string]ThingAction   `json:"actions,omitempty"`    // thing action map
	Events      map[string]ThingEvent    `json:"events,omitempty"`     // thing event map
	Links       []string                 `json:"links,omitempty"`
}

// ThingProperty defining a property of a Thing
type ThingProperty struct {
	// thingID string   // ID of the thing this property belongs to
	// Name        string `json:"name"`        // name of the property
	Description string `json:"description"` // description of the property
	Type        string `json:"type"`        // property type
	// Value       string `json:"value"`       // current value of the property
	Minimum int `json:"minimum"` // optional minimum value
	Maximum int `json:"maximum"` // optional maximum value
}

// ThingAction defining an action of a Thing
type ThingAction struct {
	// Title       string `json:"title"`       // human name of the action
	Description string `json:"description"` // description of the action
	Input       struct {
		Type       string                   `json:"type"` // input type: Object
		Properties map[string]ThingProperty `json:"properties"`
	} `json:"input"` // accepted input for the action
	Output struct {
		Type       string                   `json:"type"` // input type: Object
		Properties map[string]ThingProperty `json:"properties"`
	} `json:"output"` // accepted outputs for the action
	Safe       bool `json:"safe,omitempty"` // no internal state involved
	Idempotent bool `json:"idempotennt"`    // Action can be repeated with same result
}

// ThingEvent defining an event of a Thing
type ThingEvent struct {
	// Name string `json:"name"` // name of the event
	Description string `json:"description"` // description of the event
	Data        struct {
		Type string `json:"type"` // type of data
	} `json:"data,omitempty"` // data schema of event ???
	// Subscription string `json:"subscription,omitempty"`  // data to be passed upon subscription ???
	// Cancellation string `json:"canbcellation,omitempty"` // data to pass to cancel a subscription ???
}
