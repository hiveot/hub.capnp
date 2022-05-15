// Package mqttbinding that implements the ExposedThing API using the MQTT protocol binding
// Exposed Things are used by IoT device implementers to provide access to the device.
package mqttbinding

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"strings"
	"sync"
)

// MqttExposedThing is the implementation of an ExposedThing interface using the MQTT protocol binding.
// Thing implementers can use this API to subscribe to actions and publish TDs and events.
//
// This loosely follows the WoT scripting API for ExposedThing as described at
// https://www.w3.org/TR/wot-scripting-api/#the-exposedthing-interface
//
// Differences with the WoT scripting API:
//  1. The WoT scripting API uses ECMAScript with promises for asynchronous results.
//     This implementation uses channels to return async results, which is golang idiomatic.
//  2. The WoT scripting API method names are in 'lowerCase' format.
//     In golang lowerCase makes things private. This implementation uses 'UpperCase' name format.
//  3. Most methods are synchronous instead of asynchronous as the MQTT client is synchronous.
//     The result of actions indicates that it was submitted successfully. Actions do not have
//     a return value (in WoST) as they are not remote procedure calls. If the effect of an
//     action is needed then consumers should subscribe to property changes that are submitted
//     as the action is executed. The results of actions by others then be handled in the same way.
//  4. Actions are only handled by devices that are not asleep as the message bus does not
//     yet support queued actions. This is a limitation of the message bus. Future implementations
//     of the message bus can add queuing to support intermittent connected devices.
//     Use of the 'retain' flag is not recommended for actions on devices that also have a manual input.
//  5. If an action is not allowed then no error is returned. In most cases the MQTT bus won't accept
//     the request in the first place.
//  6. Additional functions UpdatePropertyValue(s) to support sending property change events only
//     when property values change.
//
// Example of properties in a TD with forms for mqtt protocol binding.
// The forms will likely be provided through a @context link to a standardized WoST model, once the semantics are figured out.
// {
//   "properties": {
//        "onoff": {
//            "@type": "iot:SwitchOnOff",
//            "title": "Switch on or off status"
//            "description": "More elaborate description of the onoff property"
//            "observable": true,    // form must provide an observeproperty binding
//            "type": "boolean",
//            "unit": "binary",
//            "readOnly": false,  // property is writable. form must provide a writeproperty binding
//        }
//        // These forms apply to all writable properties
//        "forms": [{
//          	"op": ["writeproperty", "writeproperties"],
//          	"href": "mqtts://{broker}/things/{thingID}/write/properties",
//				"mqv:controlPacketValue": "PUBLISH",
//              "contentType": "application/json"
//          }, {
//              // TBD. MQTT topic. How to parameterize in a generic schema?
//              "op": ["observeproperty"],
//          	"href": "mqtts://{broker}/things/{thingID}/event/properties",
//				"mqv:controlPacketValue": "SUBSCRIBE",
//              "contentType": "application/json"
//         }],
//       }
//    }
// }

type MqttExposedThing struct {
	// ExposedThing extends a ConsumedThing
	MqttConsumedThing

	// deviceID for reverse looking of device by their internal ID
	DeviceID string

	// handler for action requests
	// to set the default handler use name ""
	actionHandlers map[string]func(eThing *MqttExposedThing, actionName string, value InteractionOutput) error

	// handler for writing property requests
	// to set the default handler use name ""
	propertyWriteHandlers map[string]func(eThing *MqttExposedThing, propName string, value InteractionOutput) error
}

// Destroy the exposed thing. This stops serving external requests
func (eThing *MqttExposedThing) Destroy() {
	topic := strings.ReplaceAll(TopicEmitEvent, "{thingID}", eThing.td.ID) + "/#"
	eThing.mqttClient.Unsubscribe(topic)

	eThing.eventSubscriptions = nil
	eThing.actionHandlers = nil
}

// EmitEvent publishes a single event to subscribers.
// The topic will be things/{thingID}/event/{name} and payload will be the event data.
// If the event cannot be published, for example it is not defined, an error is returned.
//
// name is the name of the event as described in the TD, or one of the general purpose events.
// data is the event value as defined in the TD events schema and used as the payload
// Returns an error if the event is not found or cannot be published
func (eThing *MqttExposedThing) EmitEvent(name string, data interface{}) error {
	_, found := eThing.td.Events[name]
	if !found {
		logrus.Errorf("EmitEvent. Event '%s' not defined for thing '%s'", name, eThing.td.ID)
		err := errors.New("NotFoundError")
		return err
	}

	topic := strings.ReplaceAll(TopicEmitEvent, "{thingID}", eThing.td.ID) + "/" + name
	err := eThing.mqttClient.PublishObject(topic, data)
	return err
}

// EmitPropertyChange publishes a property value change event, which in turn will notify all
// observers (subscribers) of the change.
//
// The topic will be things/{thingID}/event/properties and payload will be
// a map with the property name-value pair.
// propName is the name of the property in the TD.
// newRawValue is the new raw value of the property. This will be also be stored in the valueStore.
// Returns an error if the property value cannot be published
func (eThing *MqttExposedThing) EmitPropertyChange(propName string, newRawValue interface{}) error {
	propAffordance, found := eThing.td.Properties[propName]
	if !found {
		logrus.Errorf("EmitPropertyChange. Property '%s' not found on thing '%s'", propName, eThing.td.ID)
		err := errors.New("NotFoundError")
		return err
	}

	// update the cached value in ConsumedThing base class
	io := NewInteractionOutput(newRawValue, &propAffordance.DataSchema)
	eThing._writeValue(propName, io)

	// protocol binding -> things/{thingID}/event/{propName}
	propMap := map[string]interface{}{propName: newRawValue}
	topic := strings.ReplaceAll(TopicEmitPropertiesChange, "{thingID}", eThing.td.ID)
	err := eThing.mqttClient.PublishObject(topic, propMap)
	return err
}

// EmitPropertyChanges sends a properties change event for multiple properties
// and if the property name matches an event name, an event with the property name
// is sent, if the value changed.
// This will remove properties that do not have an affordance.
// This uses the 'TopicEmitPropertiesChange' topic, eg 'things/{thingID}/event/properties'.
// propMap is a map of property name to raw value. This will be converted to json as-is.
//
// For property names that are defined as events, an event is sent for each property in the event list.
//
// @param onlyChanges: include only those properties whose value have changed (recommended)
// Returns an error if submitting an event fails
func (eThing *MqttExposedThing) EmitPropertyChanges(
	propMap map[string]interface{}, onlyChanges bool) error {
	//logrus.Infof("EmitPropertyChanges: %s", propMap)
	var err error
	changedProps := make(map[string]interface{})

	// filter properties that have no affordance or haven't changed
	for propName, newVal := range propMap {
		lastVal, found := eThing._readValue(propName)

		// In order to be included as a property it must have a propertyAffordance
		if !found || !onlyChanges || lastVal.Value != newVal {
			propAffordance, found := eThing.td.Properties[propName]
			// only include values that are in the properties map
			if found {
				changedProps[propName] = newVal
				newIO := NewInteractionOutput(newVal, &propAffordance.DataSchema)
				eThing._writeValue(propName, newIO)
			}

			// to be sent as an event it must have an event affordance
			eventAffordance, found := eThing.td.Events[propName]
			if found {
				_ = eventAffordance
				topic := strings.ReplaceAll(TopicEmitEvent, "{thingID}", eThing.td.ID)
				topic += "/" + propName
				err = eThing.mqttClient.PublishObject(topic, newVal)
				if err != nil {
					logrus.Warningf("MqqExposedThing.EmitPropertyChanges: Failed %s", err)
					return err
				}
			}
		}
	}
	// only publish if there are properties left
	if len(changedProps) > 0 {
		topic := strings.ReplaceAll(TopicEmitPropertiesChange, "{thingID}", eThing.td.ID)
		err := eThing.mqttClient.PublishObject(topic, changedProps)
		if err != nil {
			logrus.Warningf("MqqExposedThing.EmitPropertyChanges: Failed %s", err)
			return err
		}
		cpAsText, _ := json.Marshal(changedProps)
		logrus.Infof("MqqExposedThing.EmitPropertyChanges: submitted %d properties for thing %s: %s",
			len(changedProps), eThing.td.ID, cpAsText)
	}
	return err
}

// Expose starts serving external requests for the Thing so that WoT Interactions using Properties and Actions
// will be possible. This also publishes the TD document of this Thing.
func (eThing *MqttExposedThing) Expose() error {
	// Actions and Properties are handled the same.
	// An action with a property name will update the property.
	topic := strings.ReplaceAll(TopicInvokeAction, "{thingID}", eThing.td.ID) + "/#"
	eThing.mqttClient.Subscribe(topic, eThing.handleActionRequest)

	// Also publish this Thing's TD document
	topic = strings.ReplaceAll(TopicThingTD, "{thingID}", eThing.td.ID)
	err := eThing.mqttClient.PublishObject(topic, eThing.td)
	return err
}

// Handle action requests for this Thing
// This passes the request to the registered handler
// If no specific handler is set then the default handler with name "" is invoked.
func (eThing *MqttExposedThing) handleActionRequest(address string, message []byte) {
	var actionData InteractionOutput
	var err error

	logrus.Infof("MqttExposedThing.handleActionRequest: address '%s', message: '%s'", address, message)

	// the topic is "things/id/action|property/name"
	thingID, messageType, subject := SplitTopic(address)
	if thingID == "" || messageType == "" {
		logrus.Warningf("MqttExposedThing.handleActionRequest: subject is missing in topic %s", address)
		return
	}

	// determine the action/property schema
	logrus.Infof("MqttExposedThing.handleActionRequest: Received action request with topic %s", address)
	actionAffordance := eThing.td.GetAction(subject)
	if actionAffordance != nil {
		// this is a registered action
		actionData = NewInteractionOutputFromJson(message, &actionAffordance.Input)
		// TODO validate the data against the schema

		// action specific handlers takes precedence
		handler, _ := eThing.actionHandlers[subject]
		if handler != nil {
			err = handler(eThing, subject, actionData)
		} else {
			// default handler is a fallback
			defaultHandler, _ := eThing.actionHandlers[""]
			if defaultHandler != nil {
				err = defaultHandler(eThing, subject, actionData)
			} else {
				err = errors.New("no handler for action request")
			}
		}
	} else {
		// properties are written using actions
		propAffordance := eThing.td.GetProperty(subject)
		if propAffordance == nil {
			// this is a registered property
			err = errors.New("not a registered action or property")
		} else {
			eThing.handlePropertyWriteRequest(subject, propAffordance, message)
		}
	}
	if err != nil {
		logrus.Errorf("MqttExposedThing.handleActionRequest: request failed for topic %s: %s", address, err)
	}
}

// handlePropertyWriteRequest for updating a property
// This invokes the property update handler with the value of the new property.
//
// It is up to the handler to invoke emitPropertyChange and update the property in the valueStore
// after the change takes effect.
//
// There is currently no error feedback in case the request cannot be handled. The requester will receive a
// property change event when the request has completed successfully.
// Failure to complete the request can be caused by an invalid value or if the IoT device is not
// in a state to accept changes.
//
// TBD: if there is a need to be notified of failure then a future update can add a write-property failed event.
//
// If no specific handler is set for the property then the default handler with name "" is invoked.
func (eThing *MqttExposedThing) handlePropertyWriteRequest(propName string, propAffordance *thing.PropertyAffordance, message []byte) {
	var err error
	logrus.Infof("MqttExposedThing.handlePropertyWriteRequest for '%s'. property '%s'", eThing.td.ID, propName)
	var propValue interface{}

	err = json.Unmarshal(message, &propValue)
	if err != nil {
		logrus.Warningf("MqttExposedThing.handlePropertyWriteRequest: missing property value for %s: %s", propName, err)
		// TBD: reply with a failed event
		return
	}

	if propAffordance == nil {
		err = errors.New("property '%s' is not a valid name")
		logrus.Warningf("MqttExposedThing.handlePropertyWriteRequest: %s. Request ignored.", err)
	} else if propAffordance.ReadOnly {
		err = errors.New("property '" + propName + "' is readonly")
		logrus.Warningf("MqttExposedThing.handlePropertyWriteRequest: %s", err)
	} else {
		propValue := NewInteractionOutput(propValue, &propAffordance.DataSchema)
		// property specific handler takes precedence
		handler, _ := eThing.propertyWriteHandlers[propName]
		if handler != nil {
			err = handler(eThing, propName, propValue)
		} else {
			// default handler is a fallback
			defaultHandler, _ := eThing.propertyWriteHandlers[""]
			if defaultHandler == nil {
				err = errors.New("no handler for property write request")
				logrus.Warningf("MqttExposedThing.handlePropertyWriteRequest: No handler for property '%s' on thing '%s'", propName, eThing.td.ID)
			} else {
				err = defaultHandler(eThing, propName, propValue)
			}
		}
	}

}

//func (eThing *MqttExposedThing) SetPropertyReadHandler(func(name string) string) error {
//	return errors.New("not implemented")
//}

// SetActionHandler sets the handler for handling an action for the IoT device.
//  Only a single handler is active. If a handler is set when a previous handler was already set then the
//  latest handler will be used.
//
// The device code should implement this handler to updated configuration of the device.
//
// actionName is the action name this handler is for. If a single handler can take care of most actions
//  then use "" as the name to indicate it is the default handler.
//
// The handler should return nil if the write is accepted or an error if not accepted. The property value
// in the TD will be updated after the property has changed through the change notification handler.
func (eThing *MqttExposedThing) SetActionHandler(
	actionName string, actionHandler func(eThing *MqttExposedThing, actionName string, value InteractionOutput) error) {

	eThing.actionHandlers[actionName] = actionHandler
}

// SetPropertyObserveHandler sets the handler for subscribing to properties
// Not implemented as subscriptions are handled by the MQTT message bus
//func (eThing *MqttExposedThing) SetPropertyObserveHandler(handler func(name string) InteractionOutput) error {
//	_ = handler
//	return errors.New("not implemented")
//}

// SetPropertyUnobserveHandler sets the handler for unsubscribing to properties
// Not implemented as subscriptions are handled by the MQTT message bus
//func (eThing *MqttExposedThing) SetPropertyUnobserveHandler(handler func(name string) InteractionOutput) error {
//	_ = handler
//	return errors.New("not implemented")
//}

// SetPropertyReadHandler sets the handler for reading a property of the IoT device
// Not implemented as property values are updated with events and not requested.
// The latest property value can be found with the TD properties.
//func (eThing *MqttExposedThing) SetPropertyReadHandler(handler func(name string) string) error {
//	_ = handler
//	return errors.New("not implemented")
//}

// SetPropertyWriteHandler sets the handler for writing a property of the IoT device.
// This is intended to update device configuration. If the property is read-only the handler must return an error.
// Only a single handler is active. If a handler is set when a previous handler was already
//  set then the latest handler will be used.
//
// The device code should implement this handler to updated configuration of the device.
//
// propName is the property name this handler is for. Use "" for a default handler
//
// The handler should return nil if the request is accepted or an error if not accepted. The property value
// in the TD will be updated after the property has changed through the change notification handler.
func (eThing *MqttExposedThing) SetPropertyWriteHandler(
	propName string,
	writeHandler func(eThing *MqttExposedThing, propName string, value InteractionOutput) error) {

	eThing.propertyWriteHandlers[propName] = writeHandler
}

// CreateExposedThing constructs an exposed thing from a TD.
// An exposed Thing is a local instance of a thing for the purpose of interaction with remote consumers.
// Call 'Expose' to publish the TD of the thing and to start listening for actions and property write requests.
//
// tdDoc is a Thing Description document of the Thing to expose.
// mqttClient client for binding to the MQTT protocol
func CreateExposedThing(deviceID string, tdDoc *thing.ThingTD, mqttClient *mqttclient.MqttClient) *MqttExposedThing {
	eThing := &MqttExposedThing{
		DeviceID: deviceID,
		// the consumed thing does not subscribe to property events, just initialize the fields
		// TBD should an exposed thing support property change subscription?
		MqttConsumedThing: MqttConsumedThing{
			mqttClient:         mqttClient,
			td:                 tdDoc,
			eventSubscriptions: make(map[string]Subscription),
			subscriptionMutex:  sync.Mutex{},
			valueStore:         make(map[string]InteractionOutput),
			storeMutex:         sync.RWMutex{},
		},
		actionHandlers:        make(map[string]func(eThing *MqttExposedThing, actionName string, value InteractionOutput) error),
		propertyWriteHandlers: make(map[string]func(eThing *MqttExposedThing, actionName string, value InteractionOutput) error),
	}
	return eThing
}
