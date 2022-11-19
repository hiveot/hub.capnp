package service

import (
	"github.com/hiveot/hub/pkg/pubsub"
)

// MakeThingTopic makes a new topic address for publishing or subscribing to things.
func MakeThingTopic(gatewayID, thingID, messageType, name string) string {
	if gatewayID == "" {
		gatewayID = "+"
	}
	if thingID == "" {
		thingID = "+"
	}
	if messageType == "" {
		messageType = "+"
	}
	if name == "" {
		name = "+"
	}
	return pubsub.ThingsPrefix + "/" + gatewayID + "/" + thingID + "/" + messageType + "/" + name
}

// SplitTopic breaks a topic into gatewayID, thingID, message type and name
//func SplitTopic(topic string) (gatewayID, thingID, messageType, name string, err error) {
//	// Require at least 4 parts
//	parts := strings.Split(topic, "/")
//	if len(parts) < 4 {
//		err = fmt.Errorf("invalid topic address: %s", topic)
//		return
//	}
//	gatewayID = parts[1]
//	thingID = parts[2]
//	messageType = parts[3]
//	if len(parts) > 3 {
//		name = parts[4]
//	}
//	return
//}
