// Package thing with API interface definitions for the ExposedThing and ConsumedThing classes
package thing

// ActionAffordance metadata that defines how to invoke a function of a Thing to manipulate
// its state, eg toggle lamp on/off or trigger a process
type ActionAffordance struct {
	InteractionAffordance

	// Define the input data schema of the action
	Input DataSchema `json:"input,omitempty"`

	// Defines the output data schema of the action
	Output DataSchema `json:"output,omitempty"`

	// Signals if the Action is state safe (=true) or not
	// Safe actions do not change the internal state of a resource
	Safe bool `json:"safe,omitempty" default:"false"`

	// Indicate whether the action is idempotent, eg repeated calls with the same result
	Idempotent bool `json:"idempotent,omitempty" default:"false"`
}
