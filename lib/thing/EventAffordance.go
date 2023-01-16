// Package thing with API interface definitions for the ExposedThing and ConsumedThing classes
package thing

// EventAffordance with metadata that describes an event source, which asynchronously pushes
// event data to Consumers (e.g., overheating alerts).
type EventAffordance struct {
	//---InteractionAffordance starts
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

	// Form hypermedia controls to describe how an operation can be performed
	// Forms are serializations of Protocol Bindings.
	Forms []Form `json:"forms"`

	// Define URI template variables according to [RFC6570] as collection based on DataSchema declarations.
	// ... right
	UriVariables map[string]DataSchema `json:"uriVariables,omitempty"`
	//---InteractionAffordance ends

	// Data Schema of the event instance message, eg the event payload
	Data DataSchema `json:"data,omitempty"`

	// subscription is not applicable
	// dataResponse is not applicable
	// cancellation is not applicable
}