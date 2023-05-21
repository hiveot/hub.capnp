package mqttclient

import "github.com/hiveot/hub/lib/thing"

const ReadDirectoryRequestTopic = "services/directory/action/directory"
const ReadDirectoryResponseTopic = "services/directory/event/directory"
const ReadHistoryRequestTopic = "services/history/action/history"
const ReadHistoryResponseTopic = "services/history/event/history"
const ReadLatestRequestTopic = "services/history/action/latest"
const ReadLatestResponseTopic = "services/history/event/latest"

type ReadDirectoryRequest struct {
	PublisherID string `json:"publisherID,omitempty"`
	Limit       uint   `json:"limit,omitempty"`
}

type ReadDirectoryResponse struct {
	// TBD: should this just contain the value result of the service?
	TDs            []thing.ThingValue `json:"tds"`
	ItemsRemaining bool               `json:"itemsRemaining,omitempty"`
}

type ReadHistoryRequest struct {
	PublisherID string `json:"publisherID,omitempty"`
	ThingID     string `json:"thingID,omitempty"`
	Name        string `json:"name,omitempty"`
	StartTime   string `json:"startTime,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

type ReadHistoryResponse struct {
	ItemsRemaining bool               `json:"itemsRemaining"`
	Name           string             `json:"name"`
	PublisherID    string             `json:"publisherID"`
	ThingID        string             `json:"thingID"`
	Values         []thing.ThingValue `json:"history"`
}
type ReadLatestRequest struct {
	PublisherID string `json:"publisherID,omitempty"`
	ThingID     string `json:"thingID,omitempty"`
	// optional list of property and event values to return
	Names []string `json:"names,omitempty"`
}
type ReadLatestResponse struct {
	PublisherID string             `json:"publisherID,omitempty"`
	ThingID     string             `json:"thingID,omitempty"`
	Values      []thing.ThingValue `json:"values"`
}
