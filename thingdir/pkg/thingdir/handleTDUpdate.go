package thingdir

import (
	"encoding/json"

	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
)

// handleTDUpdate updates the directory with the received TD
func (tDir *ThingDir) handleTDUpdate(topic string, message []byte) {
	thingID, _, _ := mqttbinding.SplitTopic(topic)

	tdoc := make(map[string]interface{})
	err := json.Unmarshal(message, &tdoc)
	if err == nil {
		tDir.dirServer.UpdateTD(thingID, tdoc)
	}
}
