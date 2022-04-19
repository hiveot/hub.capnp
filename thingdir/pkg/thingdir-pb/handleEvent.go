package thingdirpb

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
)

// handleEvent stores the last event or property values
func (pb *ThingDirPB) handleEvent(topic string, message []byte) {
	thingID, topicType, eventName := mqttbinding.SplitTopic(topic)
	_ = topicType
	if eventName == mqttbinding.TopicSubjectProperties {
		props := make(map[string]interface{})
		err := json.Unmarshal(message, &props)
		if err != nil {
			logrus.Warningf("ThingDirPB.handleEvent. Invalid payload for topic '%s': %s", topic, err)
			return
		}
		pb.dirServer.UpdatePropertyValues(thingID, props)
	} else {
		var eventValue interface{}
		err := json.Unmarshal(message, &eventValue)
		if err != nil {
			logrus.Warningf("ThingDirPB.handleEvent. Invalid payload for topic '%s': %s", topic, err)
			return
		}
		pb.dirServer.UpdateEventValue(thingID, eventName, eventValue)
	}
}
