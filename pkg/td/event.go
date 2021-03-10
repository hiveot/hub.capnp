package td

// Thing event definition
// Credit: https://github.com/dravenk/webthing-go/blob/master/event.go

import (
	"encoding/json"
)

// Event An Event represents an individual event from a thing.
type Event struct {
	thing *Thing
	name  string
	data  json.RawMessage
	time  string
}

// EventObject An event object describes a kind of event which may be emitted by a device.
// See https://iot.mozilla.org/wot/#event-object
type EventObject struct {
	AtType      string `json:"@type,omitempty"`
	Title       string `json:"title,omitempty"`
	ObjectType  string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Unit        string `json:"unit,omitempty"`
	Links       []Link `json:"links,omitempty"`
}
