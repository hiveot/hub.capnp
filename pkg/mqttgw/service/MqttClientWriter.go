package service

import (
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

// MqttClientWriter writes messages directly to MQTT client
type MqttClientWriter struct {
	cl *mqtt.Client
}

// write a message directly to the mqttgw client
func (wr *MqttClientWriter) Write(mqttTopic string, payload []byte) error {
	newPk := packets.Packet{}
	newPk.Payload = payload
	newPk.FixedHeader.Type = packets.Publish
	newPk.TopicName = mqttTopic
	err := wr.cl.WritePacket(newPk)
	return err
}

// NewMqttClientWriter resturn a new instance of the mqttgw client writer
func NewMqttClientWriter(cl *mqtt.Client) *MqttClientWriter {
	wr := &MqttClientWriter{cl: cl}
	return wr
}
