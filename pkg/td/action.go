package td

import "encoding/json"

// Thing action definition
// credit: https://github.com/dravenk/webthing-go/blob/master/action.go

// Action An Action represents an individual action on a thing.
type Action struct {
	id            string
	thing         *Thing
	name          string
	input         *json.RawMessage
	hrefPrefix    string
	href          string
	status        string
	timeRequested string
	timeCompleted string

	// Override this with the code necessary to perform the action.
	PerformAction func() *Action

	// Override this with the code necessary to cancel the action.
	Cancel func()
}
