package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hiveot/hub/lib/thing"
	"github.com/sirupsen/logrus"
	"time"
)

// HubMqttGWClient to interact with the Hub gateway
// This has a limited feature set and is intended for web browser users.
// Except for 'ping', a successful login is required for most functions
type HubMqttGWClient struct {
	clientID string
	url      string
	paho     pahomqtt.Client
}

// Connect to the Hub with credentials
//
//	url contains the connection url, for example "tls://127.0.0.1:4883" or "wss://127.0.0.4884"
//	loginID Hub login ID and used as the clientID for publications
//	password of the hub user. This can be a password or refresh token
//	clientCert in case of certificate based authentication or nil for password auth
//	caCert to verify the server. Highly recommended but not required.
//
// Returns an error if connecting fails.
func (cl *HubMqttGWClient) Connect(
	url string, loginID, password string, clientCert *tls.Certificate, caCert *x509.Certificate) error {

	cl.url = url
	cl.clientID = loginID

	//cl := startMqttTestClient()
	opts := pahomqtt.NewClientOptions()
	//opts.SetClientID("testinguser1")
	opts.AddBroker(url)
	opts.SetUsername(loginID)
	if password != "" {
		opts.SetPassword(password)
	}
	// TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: caCert == nil,
	}
	if caCert != nil {
		caCertPool := x509.NewCertPool()
		caCertPool.AddCert(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	// TLS setup for mutual tls client authentication
	if clientCert != nil {
		clientCertList := []tls.Certificate{*clientCert}
		tlsConfig.Certificates = clientCertList
	}
	// connect
	opts.SetTLSConfig(tlsConfig)
	cl.paho = pahomqtt.NewClient(opts)

	token := cl.paho.Connect()
	<-token.Done()
	return token.Error()
}

// Disconnect from the Hub
func (cl *HubMqttGWClient) Disconnect() {
	cl.paho.Disconnect(10)
	time.Sleep(time.Millisecond * 10)
}

// PubAction publishes the given action to the hub.
// The client must be logged in first a must have permission to send action requests to the Thing
// Intended for end users and services.
//
//	 publisherID that handles the action request
//		thingID this action controls
//		actionName of the action as defined in the TD
//		actionValue to publish
//
// Returns an error if publication fails
func (cl *HubMqttGWClient) PubAction(publisherID, thingID, actionName, actionValue string) error {

	topic := fmt.Sprintf("things/%s/%s/action/%s", publisherID, thingID, actionName)
	token := cl.paho.Publish(topic, 1, false, []byte(actionValue))
	return token.Error()
}

// PubEvent publishes the given event to the hub.
// The client must be logged in first and will be used as publisher of the event.
// Intended for devices and services.
//
//	thingID this event is from
//	eventName of the event as per TD
//	eventValue to publish
//
// Returns an error if publication fails
func (cl *HubMqttGWClient) PubEvent(thingID string, eventName string, eventValue []byte) error {

	topic := fmt.Sprintf("things/%s/%s/event/%s", cl.clientID, thingID, eventName)
	logrus.Infof("clientID=%s, topic=%s", cl.clientID, topic)
	token := cl.paho.Publish(topic, 1, false, eventValue)
	return token.Error()
}

// SubAction subscribes to thing action requests
// Intended for services and bindings.
//
//	thingID whose action to subscribe to or "" for all things of this publisher
//	actionName to subscribe to or "" for all
//
// Returns an error subscription fails
func (cl *HubMqttGWClient) SubAction(thingID, actionName string, cb func(tv thing.ThingValue)) error {
	if thingID == "" {
		thingID = "+"
	}
	if actionName == "" {
		actionName = "+"
	}
	topic := fmt.Sprintf("things/%s/%s/event/%s", cl.clientID, thingID, actionName)
	token := cl.paho.Subscribe(topic, 1, func(client pahomqtt.Client, msg pahomqtt.Message) {
		tv := thing.ThingValue{}
		tv.Data = msg.Payload()
		cb(tv)
	})
	return token.Error()
}

// SubEvent subscribes to thing events
//
//	pubID to subscribe to or "" for all publishers
//	thingID to subscribe to or "" for all things the user has access to (group membership)
//	evName to subscribe to or "" for all
//
// Returns an error if subscription fails
func (cl *HubMqttGWClient) SubEvent(pubID, thingID, name string, cb func(tv thing.ThingValue)) error {
	if pubID == "" {
		pubID = "+"
	}
	if thingID == "" {
		thingID = "+"
	}
	if name == "" {
		name = "+"
	}
	topic := fmt.Sprintf("things/%s/%s/event/%s", pubID, thingID, name)
	token := cl.paho.Subscribe(topic, 1, func(client pahomqtt.Client, msg pahomqtt.Message) {

		//evPubID, evThingID, msgType, evName, err := SplitTopic(topic)
		//_ = msgType // should be 'event'
		//if err != nil {
		//	logrus.Errorf("Received invalid topic '%s': %s", topic, err)
		//}
		tv := thing.ThingValue{}
		err := json.Unmarshal(msg.Payload(), &tv)
		if err != nil {
			logrus.Errorf("Received invalid payload for topic %s. Expected ThingValue: %s", topic, err)
		}
		cb(tv)
	})
	return token.Error()
}

// NewHubMqttClient provides the client to talk to the Hub via the Hub's Mqtt gateway
func NewHubMqttClient() *HubMqttGWClient {
	cl := &HubMqttGWClient{}
	return cl
}
