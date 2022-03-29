// Package mqttbinding with messaging topics for the MQTT protocol binding
package mqttbinding

import "strings"

// TopicMessageTD topic for thing publishing its TD
const TopicMessageTD = "td"
const TopicThingTD = "things/{thingID}/" + TopicMessageTD

// TopicMessageEvent root topic for thing publishing its Thing events
const TopicMessageEvent = "event"
const TopicThingEvent = "things/{thingID}/" + TopicMessageEvent

// TopicMessageAction root topic request to start action
const TopicMessageAction = "action"
const TopicThingAction = "things/{thingID}/" + TopicMessageAction

// TopicMessageProperty root topic for publishing property value updates
const TopicMessageProperty = "property"
const TopicThingProperty = "things/{thingID}/TopicMessageProperty"

// TopicProvisionRequest topic requesting to provision of a thing device
// const TopicProvisionRequest = "provisioning" + "/{thingID}/request"

// TopicProvisionResponse topic for provisioning of a thing device
// const TopicProvisionResponse = "provisioning" + "/{thingID}/response"

// CreateTopic creates a new topic for publishing or subscribing to a message of type
// td, action, event, property
func CreateTopic(thingID string, topicMessage string) string {
	return "things/" + thingID + "/" + topicMessage
}

// SplitTopic breaks a MQTT topic into thingID and message type (td, event, action, property value)
func SplitTopic(topic string) (thingID string, topicMessage string) {
	parts := strings.Split(topic, "/")
	if len(parts) < 2 {
		return
	}
	thingID = parts[1]
	if len(parts) > 2 {
		topicMessage = parts[2]
	}
	return
}
