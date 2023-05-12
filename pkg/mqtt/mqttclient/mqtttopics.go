package mqttclient

import (
	"errors"
	"strings"
)

// MakeTopic constructs a mqtt topic for hiveot
// This builds a topic with the format: things/pubID/thingID/msgType/name
//
//	pubID is the publisher
//	thingID is the Thing's TD
//	msgType is "action", "event", or "td"
//	name is the event or action name, or td device type
//
// Returns the topic
func MakeTopic(pubID, thingID, msgType, name string) string {
	parts := []string{"things", pubID, thingID, msgType, name}
	return strings.Join(parts, "/")
}

// SplitTopic into its parts and check for errors
// This splits a MQTT topic things/pubID/thingID/msgType/name into its parts
//
//	pubID is the publisher
//	thingID is the Thing's TD
//	msgType is action, event, or td
//	name is the event or action name, or td device type
//
// Returns the topic parts or an error if it is invalid
func SplitTopic(topic string) (pubID, thingID, msgType, name string, err error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 5 {
		err = errors.New("invalid topic format: " + topic)
		return
	}
	//
	pubID = parts[1]
	thingID = parts[2]
	msgType = parts[3]
	name = parts[4]
	return
}
