package mqttclient

import (
	"errors"
	"strings"
)

type TopicType string

const (
	ThingsTopic   TopicType = "things"
	ServicesTopic TopicType = "services"
)

// MakeThingTopic constructs a mqtt topic for addressing Things
// This builds a topic with the format: things/pubID/thingID/msgType/name
//
//	pubID is the publisher
//	thingID is the Thing's TD
//	msgType is "action", "event", or "td"
//	name is the event or action name, or td device type
func MakeThingTopic(pubID, thingID, msgType, name string) string {
	parts := []string{"things", pubID, thingID, msgType, name}
	return strings.Join(parts, "/")
}

// MakeServiceTopic constructs a mqtt topic for addressing services
// This builds a topic with the format: services/serviceID/msgType/name
func MakeServiceTopic(serviceID, msgType, name string) string {
	parts := []string{"things", serviceID, msgType, name}
	return strings.Join(parts, "/")
}

// SplitServiceTopic into its parts and check for errors
// This splits a MQTT topic
//
//	   services/serviceID/msgType/name, or
//
//		serviceID is the type of service to address, eg "directory"...
//		msgType is action, event, or td
//		name is the event or action name, or td device type
//
// Returns the topic parts or an error if it is invalid
func SplitServiceTopic(topic string) (serviceID, msgType, name string, err error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 4 {
		err = errors.New("invalid services topic format: " + topic)
		return
	}
	serviceID = parts[1]
	msgType = parts[2]
	name = parts[3]
	return
}

// SplitThingsTopic into its parts and check for errors
// This splits a MQTT topic
//
//	   things/pubID/thingID/msgType/name, or
//
//		pubID is the publisher or service to be addressed
//		thingID is the Thing's ID, or "" for services
//		msgType is action, event, or td
//		name is the event or action name, or td device type
//
// Returns the topic parts or an error if it is invalid
func SplitThingsTopic(topic string) (pubID, thingID, msgType, name string, err error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 5 {
		err = errors.New("invalid things topic format: " + topic)
		return
	}
	pubID = parts[1]
	thingID = parts[2]
	msgType = parts[3]
	name = parts[4]
	return
}
