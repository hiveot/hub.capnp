package thingdirpb

// handleTDUpdate updates the directory with the updated TD
func (pb *ThingDirPB) handleTDUpdate(thingID string, thingTD map[string]interface{}, publisherID string) {
	pb.dirClient.UpdateTD(thingID, thingTD)
}
