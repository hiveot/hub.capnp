package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hiveot/hub/lib/thing"
	"github.com/sirupsen/logrus"
	"time"
)

// MqttGwClient to interact with the Hub gateway
// This has a limited feature set and is intended for web browser users.
// Except for 'ping', a successful login is required for most functions
type MqttGwClient struct {
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
func (cl *MqttGwClient) Connect(
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
func (cl *MqttGwClient) Disconnect() {
	cl.paho.Disconnect(10)
	time.Sleep(time.Millisecond * 10)
}

// PubAction publishes the given action to the hub.
// The client must be logged in first and must have permission to send action requests to the Thing
// Intended for end users and services.
//
//	publisherID that handles the action request
//	thingID this action controls
//	actionName of the action as defined in the TD
//	actionValue to publish
//
// Returns an error if publication fails
func (cl *MqttGwClient) PubAction(publisherID string, thingID string, actionName string, actionValue []byte) error {
	topic := MakeActionTopic(publisherID, thingID, actionName)
	token := cl.paho.Publish(topic, 1, false, actionValue)
	return token.Error()
}

// PubEvent publishes the given event to the hub.
// The client must be logged in first and will be used as publisher of the event.
// Intended for devices and services to publish their events, properties, or td's
// as defined in hubapi.EventNameXyz.
//
//	thingID this event is from
//	eventName to publish as defined in the TD, "td" or "properties"
//	eventValue to publish as per TD
//
// Returns an error if publication fails
func (cl *MqttGwClient) PubEvent(thingID string, eventName string, eventValue []byte) error {

	topic := MakeEventTopic(cl.clientID, thingID, eventName)
	logrus.Infof("clientID=%s, topic=%s", cl.clientID, topic)
	token := cl.paho.Publish(topic, 1, false, eventValue)
	return token.Error()
}

// PubReadDirectory requests thing TD documents from the directory service.
//
// A publisherID can be given to limit the amount of results.
//
// To receive the directory TD's, clients should subscribe to directory events using SubDirectory.
// To receive live updates of TD's, subscribe using SubReadDirectory.
//
//	publisherID optionally only return the directory of this publisher, "" for all
//
// This returns nil on success or an error on failure
func (cl *MqttGwClient) PubReadDirectory(publisherID string) error {
	topic := ReadDirectoryRequestTopic
	req := ReadDirectoryRequest{
		PublisherID: publisherID,
		Limit:       0, // not used
	}
	payload, _ := json.Marshal(req)
	token := cl.paho.Publish(topic, 1, false, payload)
	return token.Error()
}

// PubReadHistory requests reading of thing values from the history service.
//
// To receive the values, clients should subscribe to history events using SubReadHistory.
// To receive live updates of events subscribe to events using SubReadHistory.
//
//	publisherID is the publisher of the thing
//	thingID is the thing whose history to get
//	name is the name of the event whose history to get
//	startTime is optional ISO8601 time or "" for 24 hours ago
//	duration is the number of seconds to get or 0 for 3600*24
//	limit is the maximum number of results
//
// This returns nil on success or an error on failure
func (cl *MqttGwClient) PubReadHistory(
	publisherID, thingID string, name string, startTime string, duration int, limit int) error {

	req := ReadHistoryRequest{
		PublisherID: publisherID,
		ThingID:     thingID,
		Name:        name,
		StartTime:   startTime,
		Duration:    duration,
		Limit:       limit,
	}
	payload, _ := json.Marshal(req)
	topic := ReadHistoryRequestTopic
	token := cl.paho.Publish(topic, 1, false, payload)
	return token.Error()
}

// PubReadLatest publishes the requests the latest property/event values of a thing
//
// To receive the values, clients should subscribe to history events using SubLatest.
// To receive live updates of events subscribe to events using SubReadProperties.
//
// This posts topic services/{historyID}/action/properties
//
//	publisherID is the publisher of the thing
//	thingID is the thing whose values to get
//	names are the event names whose history to get or nil to get all
//
// This returns nil on success or error on failure
func (cl *MqttGwClient) PubReadLatest(
	publisherID, thingID string, names []string) error {

	topic := ReadLatestRequestTopic
	req := ReadLatestRequest{
		PublisherID: publisherID,
		ThingID:     thingID,
		Names:       names,
	}
	payload, _ := json.Marshal(req)

	token := cl.paho.Publish(topic, 1, false, payload)
	return token.Error()
}

// SubReadDirectory subscribes to the response of PubReadDirectory
//
// cb is the callback invoked when a response is received
func (cl *MqttGwClient) SubReadDirectory(cb func(*ReadDirectoryResponse)) error {
	topic := ReadDirectoryResponseTopic
	token := cl.paho.Subscribe(topic, 1,
		func(client pahomqtt.Client, message pahomqtt.Message) {
			resp := ReadDirectoryResponse{}
			err := json.Unmarshal(message.Payload(), &resp)
			if err != nil {
				logrus.Errorf("response on topic '%s' is unexpected json: %s", topic, err)
			} else {
				cb(&resp)
			}
		})
	return token.Error()
}

// SubReadHistory subscribes to the response of PubReadHistory
//
//	cb is the callback invoked when a response is received
func (cl *MqttGwClient) SubReadHistory(cb func(*ReadHistoryResponse)) error {
	topic := ReadHistoryResponseTopic
	token := cl.paho.Subscribe(topic, 1,
		func(client pahomqtt.Client, message pahomqtt.Message) {
			resp := ReadHistoryResponse{}
			err := json.Unmarshal(message.Payload(), &resp)
			if err != nil {
				logrus.Errorf("response on topic '%s' is unexpected json: %s", topic, err)
			} else {
				cb(&resp)
			}
		})
	return token.Error()
}

// SubReadLatest subscribes to receive responses to PubReadLatest
//
// This subscribes to the topic services/{historyID}/event/properties
//
//	cb is the callback invoked when properties are received. The publisher and thingID
//	fields of ThingValue are not populated
func (cl *MqttGwClient) SubReadLatest(
	cb func(*ReadLatestResponse)) error {

	topic := ReadLatestResponseTopic
	token := cl.paho.Subscribe(topic, 1,
		func(client pahomqtt.Client, message pahomqtt.Message) {
			resp := ReadLatestResponse{}
			err := json.Unmarshal(message.Payload(), &resp)
			if err != nil {
				logrus.Errorf("response on topic '%s' is unexpected json: %s", topic, err)
			} else {
				cb(&resp)
			}
		})
	return token.Error()
}

// SubAction subscribes to thing action requests for use by publishers of Things.
//
//	thingID whose action to subscribe to or "" for all things of this publisher
//	actionName to subscribe to or "" for all
//
// Returns an error subscription fails
func (cl *MqttGwClient) SubAction(thingID, actionName string, cb func(tv thing.ThingValue)) error {
	if thingID == "" {
		thingID = "+"
	}
	if actionName == "" {
		actionName = "+"
	}
	topic := MakeActionTopic(cl.clientID, thingID, actionName)
	token := cl.paho.Subscribe(topic, 1, func(client pahomqtt.Client, msg pahomqtt.Message) {
		tv := thing.ThingValue{}
		err := json.Unmarshal(msg.Payload(), &tv)
		if err != nil {
			logrus.Errorf("Action with invalid payload on topic '%s': %s", msg.Topic(), err)
			return
		}
		cb(tv)
	})
	return token.Error()
}

// SubEvent subscribes to thing events
//
//	publisherID to subscribe to or "" for all publishers
//	thingID to subscribe to or "" for all things the user has access to (group membership)
//	eventName to subscribe to or "" for all
//
// Returns an error if subscription fails
func (cl *MqttGwClient) SubEvent(publisherID, thingID, eventName string, cb func(tv thing.ThingValue)) error {
	if publisherID == "" {
		publisherID = "+"
	}
	if thingID == "" {
		thingID = "+"
	}
	if eventName == "" {
		eventName = "+"
	}
	topic := MakeEventTopic(publisherID, thingID, eventName)
	token := cl.paho.Subscribe(topic, 1, func(client pahomqtt.Client, msg pahomqtt.Message) {
		tv := thing.ThingValue{}
		err := json.Unmarshal(msg.Payload(), &tv)
		if err != nil {
			logrus.Errorf("Event with invalid payload for topic %s: %s", msg.Topic(), err)
		}
		cb(tv)
	})
	return token.Error()
}

// NewHubMqttClient provides the client to talk to the Hub via the Hub's Mqtt gateway
func NewHubMqttClient() *MqttGwClient {
	cl := &MqttGwClient{}
	return cl
}
