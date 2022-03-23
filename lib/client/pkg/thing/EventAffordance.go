// Package thing with API interface definitions for the ExposedThing and ConsumedThing classes
package thing

// EventAffordance with metadata that describes an event source, which asynchronously pushes
// event data to Consumers (e.g., overheating alerts).
type EventAffordance struct {
	InteractionAffordance

	// Data schema of the event instance message, eg the event payload
	Data DataSchema `json:"data,omitempty"`

	// subscription is not applicable
	// dataResponse is not applicable
	// cancellation is not applicable
}
