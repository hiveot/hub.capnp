package thingdir

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/wost-go/pkg/consumedthing"
)

// handleEvent stores the last event or property values
func (tDir *ThingDir) handleEvent(topic string, message []byte) {
	thingID, topicType, eventName := consumedthing.SplitTopic(topic)
	_ = topicType
	if eventName == consumedthing.TopicSubjectProperties {
		props := make(map[string]interface{})
		err := json.Unmarshal(message, &props)
		if err != nil {
			logrus.Warningf("ThingDirPB.handleEvent. Invalid payload for topic '%s': %s", topic, err)
			return
		}
		tDir.dirServer.UpdatePropertyValues(thingID, props)
	} else {
		var eventValue interface{}
		err := json.Unmarshal(message, &eventValue)
		if err != nil {
			logrus.Warningf("ThingDirPB.handleEvent. Invalid payload for topic '%s': %s", topic, err)
			return
		}
		tDir.dirServer.UpdateEventValue(thingID, eventName, eventValue)
	}
}
