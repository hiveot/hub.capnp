// Package mqttbinding with Subscription definitions for the ExposedThing and ConsumedThing classes
package mqttbinding

// PropertyMap represents a map of Property names as strings to a value that the Property can take.
// It is used as a property bag for interactions that involve multiple Properties at once.
//type PropertyMap map[string]interface{}

// AsyncResult from a channel that carries a result value or an error
//type AsyncResult struct {
//	result interface{}
//	err    error
//}

//type ProtocolBinding interface

const (
	SubscriptionTypeAction   = "action"
	SubscriptionTypeEvent    = "event"
	SubscriptionTypeProperty = "property"
)

// Subscription describes the type of observed property, event or action
type Subscription struct {
	SubType string // "property" | "event" | "action" | nil
	Name    string // property, event or action name, or "" for all properties, events or actions
	//interaction ThingInteraction // not clear what the purpose of this is. Validation? tbd
	//form        ThingForm        // not clear what the purpose of this is. Validation? tbd
	Handler func(name string, message InteractionOutput)
}

// Stop delivering notifications for this subscription
func (sub *Subscription) Stop() {
	// set handler to inactive
	sub.Handler = nil
}
