package thingdir

import (
	"encoding/json"

	"github.com/wostzone/wost-go/pkg/consumedthing"
)

// handleTDUpdate updates the directory with the received TD
func (tDir *ThingDir) handleTDUpdate(topic string, message []byte) {
	thingID, _, _ := consumedthing.SplitTopic(topic)

	tdoc := make(map[string]interface{})
	err := json.Unmarshal(message, &tdoc)
	if err == nil {
		_ = tDir.dirServer.UpdateTD(thingID, tdoc)
	}
}
