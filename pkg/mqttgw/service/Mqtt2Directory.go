package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hiveot/hub/lib/resolver"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/mqttgw/mqttclient"
)

// Mqtt2Directory handles directory requests over MQTT using the gateway/resolver
type Mqtt2Directory struct {

	// login ID of this client
	clientID string

	// directory capabilities
	readDir directory.IReadDirectory
	writer  *MqttClientWriter
}

// Return the read directory capability. Obtain it from the resolver on first use.
func (m2dir *Mqtt2Directory) getReadDirectory() directory.IReadDirectory {
	if m2dir.readDir == nil {
		m2dir.readDir = resolver.GetCapability[directory.IReadDirectory]()
	}
	return m2dir.readDir
}

// Release the directory capability
func (m2dir *Mqtt2Directory) Release() {
	if m2dir.readDir != nil {
		m2dir.readDir.Release()
	}
}

// handleReadDirectory reads the directory using request parameters from payload and writes a response
//
//	payload contains optional filter parameters for publisherID and limit.
//	The default limit is 1000
func (m2dir *Mqtt2Directory) handleReadDirectory(payload []byte) (err error) {
	req := mqttclient.ReadDirectoryRequest{Limit: 1000}

	if payload != nil && len(payload) > 0 {
		err = json.Unmarshal(payload, &req)
	}
	if err != nil {
		err = fmt.Errorf("directory request invalid parameters: %w", err)
	} else {
		cursor := m2dir.getReadDirectory().Cursor(context.Background())
		values, itemsRemaining := cursor.NextN(req.Limit)
		resp := mqttclient.ReadDirectoryResponse{
			TDs:            values,
			ItemsRemaining: itemsRemaining,
		}
		respJson, _ := json.Marshal(resp)
		err = m2dir.writer.Write(mqttclient.ReadDirectoryResponseTopic, respJson)
	}
	return err
}

// HandleDirectoryRequest handles a directory service request over MQTT
//
//	request topic: mqttclient.ReadDirectoryRequestTopic; payload ReadDirectoryRequest{}
//	return topic:  mqttclient.ReadDirectoryResponseTopic; payload ReadDirectoryResponse
func (m2dir *Mqtt2Directory) HandleDirectoryRequest(topic string, payload []byte) (err error) {
	if topic == mqttclient.ReadDirectoryRequestTopic {
		err = m2dir.handleReadDirectory(payload)
	} else {
		err = fmt.Errorf("unknown directory request from client '%s' on topic %s", m2dir.clientID, topic)
	}
	return err
}

// NewMqtt2Directory returns a handler for directory requests over MQTT
//
//	clientID is the user loginID of the client
//	writer is used to send a response to this client
func NewMqtt2Directory(clientID string, writer *MqttClientWriter) *Mqtt2Directory {
	handler := &Mqtt2Directory{
		clientID: clientID,
		writer:   writer,
	}
	return handler
}
