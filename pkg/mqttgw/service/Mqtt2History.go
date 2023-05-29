package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hiveot/hub/api/go/vocab"
	"github.com/hiveot/hub/lib/resolver"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/mqttgw/mqttclient"
	"github.com/mochi-co/mqtt/v2"
	"time"
)

// Mqtt2History handles history service requests over MQTT
type Mqtt2History struct {
	// the logged in client's ID
	clientID string

	// history capabilities
	readHist history.IReadHistory

	// the mqttgw client this serves
	mqttClient *mqtt.Client

	writer *MqttClientWriter
}

// Return the device ReadHistory capability. Obtain it from the resolver on first use.
func (m2hist *Mqtt2History) getReadHistory() history.IReadHistory {
	// FIXME: apply publisherID and thingID
	if m2hist.readHist == nil {
		m2hist.readHist = resolver.GetCapability[history.IReadHistory]()
	}
	return m2hist.readHist
}

func (m2hist *Mqtt2History) handleReadHistory(payload []byte) (err error) {
	req := mqttclient.ReadHistoryRequest{}
	err = json.Unmarshal(payload, &req)
	if req.StartTime == "" {
		ago := time.Now().Add(-time.Hour * 24)
		req.StartTime = ago.Format(vocab.ISO8601Format)
	}
	if req.Duration == 0 {
		req.Duration = 24 * 3600
	}

	rh := m2hist.getReadHistory()
	cursor := rh.GetEventHistory(context.Background(), req.PublisherID, req.ThingID, req.Name)
	val1, isValid := cursor.Seek(req.StartTime)
	results := []thing.ThingValue{val1}
	if isValid {
		var resp = mqttclient.ReadHistoryResponse{}
		batch, itemsRemaining := cursor.NextN(uint(req.Limit))

		// todo: truncate on duration
		results = append(results, batch...)
		//
		resp.Name = req.Name
		resp.PublisherID = req.PublisherID
		resp.ThingID = req.ThingID
		resp.ItemsRemaining = itemsRemaining
		resp.Values = results
		//
		respJson, _ := json.Marshal(resp)
		m2hist.writer.Write(mqttclient.ReadHistoryResponseTopic, respJson)
	}
	return err
}

func (m2hist *Mqtt2History) handleReadLatest(payload []byte) (err error) {
	req := mqttclient.ReadLatestRequest{}
	err = json.Unmarshal(payload, &req)
	if err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	rh := m2hist.getReadHistory()
	thingValues := rh.GetProperties(context.Background(), req.PublisherID, req.ThingID, req.Names)
	if thingValues != nil {
		var resp = mqttclient.ReadLatestResponse{
			PublisherID: req.PublisherID,
			ThingID:     req.ThingID,
			Values:      thingValues,
		}
		//
		respJson, _ := json.Marshal(resp)
		err = m2hist.writer.Write(mqttclient.ReadLatestResponseTopic, respJson)
	}
	return err
}

// HandleHistoryRequest handles a request over MQTT to read history
//
// Topics:
//
//	read history: services/history/action/history
//	reply: services/history/event/history
//
//	read latest: services/history/action/latest
//	reply:  services/history/event/latest
func (m2hist *Mqtt2History) HandleHistoryRequest(topic string, payload []byte) (err error) {

	if topic == mqttclient.ReadHistoryRequestTopic {
		err = m2hist.handleReadHistory(payload)
	} else if topic == mqttclient.ReadLatestRequestTopic {
		err = m2hist.handleReadLatest(payload)
	}
	return err
}

// Release the history capability
func (m2hist *Mqtt2History) Release() {
	if m2hist.readHist != nil {
		m2hist.readHist.Release()
	}
}

// NewMqtt2History returns a handler for history requests over MQTT
func NewMqtt2History(clientID string, writer *MqttClientWriter) *Mqtt2History {

	m2hist := &Mqtt2History{
		clientID: clientID,
		writer:   writer,
	}
	return m2hist
}
