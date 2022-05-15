// Package mqttbinding with messaging topics for the MQTT protocol binding
package mqttbinding

import (
	"strings"
)

// TopicTypeTD topic for thing publishing its TD
const TopicTypeTD = "td"
const TopicThingTD = "things/{thingID}/" + TopicTypeTD

// TopicTypeEvent base topic for thing publishing its Thing events
const TopicTypeEvent = "event"
const TopicEmitEvent = "things/{thingID}/" + TopicTypeEvent

// TopicTypeAction base topic request to start action
const TopicTypeAction = "action"
const TopicInvokeAction = "things/{thingID}/" + TopicTypeAction

// TopicSubjectProperties base topic for publishing a map of property values updates
const TopicSubjectProperties = "properties"

// TopicEmitPropertiesChange base topic for publishing property value updates
const TopicEmitPropertiesChange = "things/{thingID}/" + TopicTypeEvent + "/" + TopicSubjectProperties

// TopicReadProperties topic to submit request to receive a property event with property values
const (
	TopicTypeRead       = "read"
	TopicReadProperties = "things/{thingID}/" + TopicTypeRead + "/" + TopicSubjectProperties
)

// TopicProvisionRequest topic requesting to provision of a thing device
// const TopicProvisionRequest = "provisioning" + "/{thingID}/request"

// TopicProvisionResponse topic for provisioning of a thing device
// const TopicProvisionResponse = "provisioning" + "/{thingID}/response"

// CreateTopic creates a new topic for publishing or subscribing to a message of type
// td, action, event, property
// thingID to listen on. "" or "+" for any thingID
func CreateTopic(thingID string, topicMessageType string) string {
	if thingID == "" {
		thingID = "+"
	}
	return "things/" + thingID + "/" + topicMessageType
}

// SplitTopic breaks a MQTT topic into thingID, topic type (td, event, action, property value)
// and optionally a subject like for example 'properties' in 'things/event/properties'
func SplitTopic(topic string) (thingID string, topicType string, subject string) {
	parts := strings.Split(topic, "/")
	if len(parts) < 2 {
		//err = errors.New("Topic too short")
		return
	}
	thingID = parts[1]
	if len(parts) > 2 {
		topicType = parts[2]
	}
	if len(parts) > 3 {
		subject = parts[3]
	}
	return
}
