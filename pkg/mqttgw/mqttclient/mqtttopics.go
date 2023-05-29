package mqttclient

import (
	"errors"
	"strings"
)

// topics for passing messages
const (
	MessageTypeAction          = "action"
	MessageTypeEvent           = "event"
	ThingsTopicPrefix          = "things"
	DirectoryTopicPrefix       = "services/directory"
	HistoryTopicPrefix         = "services/history"
	ReadDirectoryRequestTopic  = DirectoryTopicPrefix + "/action/directory"
	ReadDirectoryResponseTopic = DirectoryTopicPrefix + "/event/directory"
	ReadHistoryRequestTopic    = HistoryTopicPrefix + "/action/history"
	ReadHistoryResponseTopic   = HistoryTopicPrefix + "/event/history"
	ReadLatestRequestTopic     = HistoryTopicPrefix + "/action/latest"
	ReadLatestResponseTopic    = HistoryTopicPrefix + "/event/latest"
)

// IsThingsTopic test if the given topic is a thing pub/sub topic
func IsThingsTopic(topic string) bool {
	return strings.HasPrefix(topic, ThingsTopicPrefix)
}

// IsDirectoryTopic test if the given topic is a directory service topic
func IsDirectoryTopic(topic string) bool {
	return strings.HasPrefix(topic, DirectoryTopicPrefix)
}

// IsHistoryTopic test if the given topic is a history service topic
func IsHistoryTopic(topic string) bool {
	return strings.HasPrefix(topic, HistoryTopicPrefix)
}

// MakeActionTopic constructs a mqttgw topic for publishing Thing actions
// This builds a topic with the format: things/publisherID/thingID/action/name
//
//	publisherID is the publisher
//	thingID is the Thing's TD
//	name is the event or action name, or td device type
func MakeActionTopic(publisherID, thingID, name string) string {
	parts := []string{ThingsTopicPrefix, publisherID, thingID, MessageTypeAction, name}
	return strings.Join(parts, "/")
}

// MakeEventTopic constructs a mqttgw topic for addressing Things
// This builds a topic with the format: things/publisherID/thingID/event/name
//
//	publisherID is the publisher
//	thingID is the Thing's TD
//	name is the event or action name
func MakeEventTopic(publisherID, thingID, name string) string {
	parts := []string{ThingsTopicPrefix, publisherID, thingID, MessageTypeEvent, name}
	return strings.Join(parts, "/")
}

// SplitThingsTopic into its parts and check for errors
// This splits a MQTT topic into its parts:
//
//	things/publisherID/thingID/msgType/name
//
// Where:
//   - publisherID is the publisher or service to be addressed
//   - thingID is the Thing's ID to be address
//   - msgType is action or event
//   - name is the event or action name
//
// Returns the topic parts or an error if it is invalid
func SplitThingsTopic(topic string) (publisherID, thingID, msgType, name string, err error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 5 {
		err = errors.New("invalid things topic format: " + topic)
		return
	}
	if parts[0] != ThingsTopicPrefix {
		err = errors.New("not a things topic: " + topic)
		return
	}
	publisherID = parts[1]
	thingID = parts[2]
	msgType = parts[3]
	name = parts[4]
	return
}
