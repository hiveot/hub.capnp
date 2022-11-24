package service

import (
	"github.com/hiveot/hub/pkg/pubsub"
)

// MakeThingTopic makes a new topic address for publishing or subscribing to things.
// Thing address topics are: thingAddress/messageType/name, where thingAddress is publisherID/thingID
func MakeThingTopic(thingAddr, messageType, name string) string {
	if thingAddr == "" {
		thingAddr = "+/+"
	}
	if messageType == "" {
		messageType = "+"
	}
	if name == "" {
		name = "+"
	}
	return pubsub.ThingsPrefix + "/" + thingAddr + "/" + messageType + "/" + name
}

// MakePublisherThingTopic makes a new topic address from gateway and thingID for publishing or subscribing to things.
//func MakePublisherThingTopic(publisherID, thingID, messageType, name string) string {
//	if publisherID == "" {
//		publisherID = "+"
//	}
//	if thingID == "" {
//		thingID = "+"
//	}
//	if messageType == "" {
//		messageType = "+"
//	}
//	if name == "" {
//		name = "+"
//	}
//	return pubsub.ThingsPrefix + "/" + publisherID + "/" + thingID + "/" + messageType + "/" + name
//}

// SplitTopic breaks a topic into publisherID, thingID, message type and name
//func SplitTopic(topic string) (publisherID, thingID, messageType, name string, err error) {
//	// Require at least 4 parts
//	parts := strings.Split(topic, "/")
//	if len(parts) < 4 {
//		err = fmt.Errorf("invalid topic address: %s", topic)
//		return
//	}
//	publisherID = parts[1]
//	thingID = parts[2]
//	messageType = parts[3]
//	if len(parts) > 3 {
//		name = parts[4]
//	}
//	return
//}