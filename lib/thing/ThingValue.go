package thing

// ThingValue contains an event, action value or TD of a thing
type ThingValue struct {
	// Address of the thing related to the value
	// Addresses are publisherID/thingID
	ThingAddr string

	// Name of event or action as described in the thing TD
	Name string

	// Event Value, JSON encoded
	ValueJSON []byte

	// Timestamp the value was created, in ISO8601 UTC format. Default "" is 'now'
	Created string
	// Timestamp in unix time, msec since Epoch.
	//CreatedMsec int64

	// Expiry time of the message in seconds since epoc.
	// Events expire based on their update interval.
	// Actions expiry is used for queueing. 0 means the action expires immediately after receiving it and is not queued.
	//Expiry int64

	// Sequence of the message from its creator. Intended to prevent replay attacks.
	//Sequence int64
}

// NewThingValue creates a new ThingValue object with the address of the thing, the value name and the serialized value
// This copies the value buffer.
func NewThingValue(thingAddr, name string, value []byte) *ThingValue {
	return &ThingValue{
		ThingAddr: thingAddr,
		Name:      name,
		// DO NOT REMOVE THE TYPE CONVERSION
		// this clones the value so the value buffer can be reused
		ValueJSON: []byte(string(value)),
	}
}
