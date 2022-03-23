// Package mqttbinding that implements the MQTT protocol binding for the ConsumedThing API
// Consumed Things are remote representations of Things used by consumers.
package mqttbinding

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"strings"
	"sync"
)

// MqttConsumedThing is the implementation of an ConsumedThing interface using the MQTT protocol binding.
// Thing consumers can use this API to subscribe to events and actions.
//
// The protocol binding schema is work in progress, as is the WoT MQTT protocol binding. For now this takes
// a best guess approach based on "https://w3c.github.io/wot-binding-templates"
//
// This closely follows the WoT scripting API for ConsumedThing as described at
// https://www.w3.org/TR/wot-scripting-api/#the-consumedthing-interface
type MqttConsumedThing struct {
	// Client for MQTT message bus access
	mqttClient *mqttclient.MqttClient
	// Internal slot with Thing Description document this Thing exposes
	td *thing.ThingTD
	// internal slot for Subscriptions by event name
	eventSubscriptions map[string]Subscription
	// mutex for async updating of subscriptions
	subscriptionMutex sync.Mutex
	// valueStore holds property values as received by events
	valueStore map[string]InteractionOutput
}

// GetThingDescription returns the TD document of this exposed Thing
// This returns the cached version of the TD
func (cThing *MqttConsumedThing) GetThingDescription() *thing.ThingTD {
	return cThing.td
}

// Handle incoming events.
// If the event name is that of a property then use the property data schema, otherwise look for
// the data schema in the events map of the TD.
// Store the resulting event data as ActionOutput
// Invoke the subscriber to the event name, if any, or the default subscriber
//  address is the MQTT topic that the event is published on as: things/{id}/event/{eventName}
//  whereas message is the body of the event.
func (cThing *MqttConsumedThing) handleEvent(address string, message []byte) {
	var evData InteractionOutput

	// the event topic is "things/id/event/name"
	parts := strings.Split(address, "/")
	if len(parts) < 4 {
		logrus.Warningf("MqttConsumedThing.handleEvent: EventName is missing in topic %s", address)
		return
	}
	eventName := parts[3]

	//evData := InteractionOutput{}
	//err := json.Unmarshal(message, &evData)
	//if err != nil {
	//	logrus.Warningf("MqttConsumedThing.handleEvent: Unable to unmarshal event on topic %s", address)
	//	return
	//}

	// handle property events
	propAffordance := cThing.td.GetProperty(eventName)
	if propAffordance != nil {
		evData = NewInteractionOutput(message, &propAffordance.DataSchema)

		logrus.Infof("MqttConsumedThing.handleEvent: Event with topic %s is a property event", address)
		// TODO validate the data
	} else {
		// not a property event
		eventAffordance := cThing.td.GetEvent(eventName)
		if eventAffordance != nil {
			evData = NewInteractionOutput(message, &eventAffordance.Data)
			logrus.Infof("MqttConsumedThing.handleEvent: Event with topic %s is not a property event", address)
		} else {
			// unknown schema
			evData = NewInteractionOutput(message, nil)
		}
	}
	// property or event, it is stored in the valueStore
	cThing.valueStore[eventName] = evData

	// notify subscribers if any
	subscription, found := cThing.eventSubscriptions[eventName]
	if !found || subscription.Handler == nil {
		subscription, found = cThing.eventSubscriptions[""] // default subscriber
	}
	if subscription.Handler != nil {
		subscription.Handler(eventName, evData)
	}
}

// InvokeAction makes a request for invoking an Action and returns once the
// request is submitted.
//
// WoST actions are used to update properties and control inputs indicated by @type.
// The TD action schema describes the inputs and protocol to submit the action. This
// is still work in progress.
//
// TD example:
// {
//   "actions": {
//        "actionName": {
//            // type of action with namespace, eg iot:switch or wost:configuration
//            "@type": "iot:SwitchOnOff",
//            // input parameters of the action
//            "input": {
//                "title": "Switch on or off value"
//                "type": "boolean",
//             },
//          },
//          // protocol binding for the actions
//          "forms": [{
//             // TBD. MQTT topic. How to parameterize in a generic schema?
//             "mqtt": "things/${thingID}/action/${actionName}",
//             "op": ["invokeaction"],
//             "mqv:controlPacketValue": "PUBLISH",
//             "contentType": "application/json"
//           }],
//        }
//    }
// }
//
// This will be posted as:
// MQTT publish on topic "things/thingID/action/actionName"
// data: { "input": true|false }
//
// Takes as arguments actionName, optionally action data as defined in the TD.
// Returns nil if the action request was submitted successfully or an error if failed
func (cThing *MqttConsumedThing) InvokeAction(actionName string, data interface{}) error {
	topic := strings.ReplaceAll(TopicAction, "{id}", cThing.td.ID)
	topic += "/" + actionName
	message := data

	return cThing.mqttClient.PublishObject(topic, message)
}

// ObserveProperty makes a request for Property value change notifications.
// Takes as arguments propertyName, listener.
func (cThing *MqttConsumedThing) ObserveProperty(
	name string, handler func(name string, data InteractionOutput)) error {
	var err error = nil
	sub := Subscription{
		SubType: SubscriptionTypeProperty,
		Name:    name,
		Handler: handler,
	}
	cThing.eventSubscriptions[name] = sub
	return err
}

// ReadProperty reads a Property value
// Returns the last known property value as a string or an error if
// the name is not a known property.
func (cThing *MqttConsumedThing) ReadProperty(name string) (InteractionOutput, error) {
	//return res, errors.New("'"+name + "' is not a known property" )
	value, found := cThing.valueStore[name]
	if !found {
		return value, errors.New("Property " + name + " does not exist on thing " + cThing.td.ID)
	}
	return value, nil
}

// ReadMultipleProperties reads multiple Property values with one request.
// propertyNames is an array with names of properties to return
// Returns a PropertyMap object that maps keys from propertyNames to InteractionOutput of that property.
func (cThing *MqttConsumedThing) ReadMultipleProperties(names []string) map[string]InteractionOutput {
	res := make(map[string]InteractionOutput, 0)
	for _, name := range names {
		output, _ := cThing.ReadProperty(name)
		res[name] = output
	}
	return res
}

// ReadAllProperties reads all properties of the Thing with one request.
// Returns a PropertyMap object that maps keys from all Property names to InteractionOutput
// of the properties.
func (cThing *MqttConsumedThing) ReadAllProperties() map[string]InteractionOutput {
	res := make(map[string]InteractionOutput, 0)

	for name := range cThing.td.Properties {
		output, _ := cThing.ReadProperty(name)
		res[name] = output
	}
	return res
}

// Stop delivering notifications for event subscriptions
func (cThing *MqttConsumedThing) Stop() {
	topic := strings.ReplaceAll(TopicThingEvent, "{thingID}", cThing.td.ID)
	cThing.mqttClient.Unsubscribe(topic)
}

// SubscribeEvent makes a request for subscribing to Event or property change notifications.
// Takes as arguments eventName, listener and optionally onerror and option
// When eventName is a propertyName the event is a property value update.
// The engine already subscribes to events for updating properties, use this subscription to be notified of
// a change to a particular property or a specific event.
// Returns nil if subscription is successful or error if failed
func (cThing *MqttConsumedThing) SubscribeEvent(
	eventName string, handler func(eventName string, data InteractionOutput)) error {
	sub := Subscription{
		SubType: SubscriptionTypeEvent, // what is the purpose of capturing this?
		Name:    eventName,
		Handler: handler,
	}
	cThing.eventSubscriptions[eventName] = sub
	return nil
}

// WriteProperty submit a request to change a property value.
// Takes as arguments propertyName and value, and sends a property update to the exposedThing that in turn
// updates the actual device.
// This does not update the property immediately. It is up to the exposedThing to perform necessary validation
// and notify subscribers with an event after the change has been applied.
//
// This will be posted as topic "things/thingID/property"
// { "propertyName": true|false }
//
// It returns an error if the property update could not be sent and nil if it is successfully
//  published. Final confirmation is obtained if an event is received with the updated property value.
func (cThing *MqttConsumedThing) WriteProperty(name string, value interface{}) error {
	messageObject := map[string]interface{}{name: value}

	topic := strings.ReplaceAll(TopicThingProperty, "{id}", cThing.td.ID)
	err := cThing.mqttClient.PublishObject(topic, messageObject)
	return err
}

// WriteMultipleProperties writes multiple property values with one request.
// Takes as arguments properties - as a map keys being Property names and values as Property values.
//
// This will be posted as:
// MQTT publish on topic "things/thingID/property"
// {
//     "propertyName1": value1,
//     "propertyName2": value2,
//     ...
// }
//
// It returns an error if the action could not be sent and nil if the action is successfully
//  published. Final success is achieved if the property value will be updated through an event.
func (cThing *MqttConsumedThing) WriteMultipleProperties(
	properties map[string]interface{}) error {

	topic := strings.ReplaceAll(TopicThingProperty, "{id}", cThing.td.ID)
	err := cThing.mqttClient.PublishObject(topic, properties)
	return err
}

// Consume constructs a consumed thing instance from a TD.
// A consumed Thing is a remote instance of a thing for the purpose of interaction by remote consumers.
// This subscribes to events from the remote thing.
//
// tdDoc is a Thing Description document of the Thing to expose.
// mqttClient client for binding to the MQTT protocol
func Consume(tdDoc *thing.ThingTD, mqttClient *mqttclient.MqttClient) *MqttConsumedThing {
	cThing := MqttConsumedThing{
		mqttClient:         mqttClient,
		td:                 tdDoc,
		eventSubscriptions: make(map[string]Subscription),
		subscriptionMutex:  sync.Mutex{},
		valueStore:         make(map[string]InteractionOutput),
	}
	// in order to keep props up to date, subscribe to all events
	topic := strings.ReplaceAll(TopicThingEvent, "{id}", cThing.td.ID) + "/#"
	cThing.mqttClient.Subscribe(topic, cThing.handleEvent)

	return &cThing
}
