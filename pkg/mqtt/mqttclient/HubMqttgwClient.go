package mqttclient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hiveot/hub/lib/thing"
)

// HubMqttgwClient to interact with the Hub gateway
// This has a limited feature set and is intended for web browser users.
// Except for 'ping', a successful login is required for most functions
type HubMqttgwClient struct {
	url  string
	paho pahomqtt.Client
}

// Connect to the Hub with credentials
//
//	url contains the connection url, eg "tls://127.0.0.1:4883"
//	loginID Hub login ID
//	password of the hub user
func (cl *HubMqttgwClient) Connect(url string, loginID, password string) error {
	cl.url = url

	//cl := startMqttTestClient()
	opts := pahomqtt.NewClientOptions()
	//opts.SetClientID("testinguser1")
	opts.AddBroker(url)
	opts.SetUsername(loginID)
	opts.SetPassword(password)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts.SetTLSConfig(tlsConfig)
	cl.paho = pahomqtt.NewClient(opts)

	token := cl.paho.Connect()
	<-token.Done()
	return token.Error()
}

// Disconnect from the Hub
func (cl *HubMqttgwClient) Disconnect() {
	cl.paho.Disconnect(1000)
}

// PubEvent publishes the given event to the hub
// The user must be logged in first.
func (cl *HubMqttgwClient) PubEvent() {

}

// SubEvent subscribes to thing events
func (cl *HubMqttgwClient) SubEvent(pubID, thingID, evName string, cb func(tv thing.ThingValue)) error {
	topic := fmt.Sprintf("things/%s/%s/event/%s", pubID, thingID, evName)
	token := cl.paho.Subscribe(topic, 1, func(client pahomqtt.Client, msg pahomqtt.Message) {
		tv := thing.ThingValue{}
		json.Unmarshal(msg.Payload(), &tv)
		cb(tv)
	})
	return token.Error()
}

// NewHubMqttClient provides the client to talk to the Hub via the Hub's Mqtt gateway
func NewHubMqttClient() *HubMqttgwClient {
	cl := &HubMqttgwClient{}
	return cl
}
