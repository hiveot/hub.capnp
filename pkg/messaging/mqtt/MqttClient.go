package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// ConnectionTimeoutSec constant with connection and reconnection timeouts
const ConnectionTimeoutSec = 20

// TLSPort is the default secure port to connect to mqtt
const TLSPort = 8883

// MqttClient client that implements IGatewayMessenger
type MqttClient struct {
	clientID string // unique ID of the client
	hostPort string // host:port of server to connect to
	pubQos   byte
	subQos   byte
	//
	isRunning           bool                          // listen for messages while running
	pahoClient          pahomqtt.Client               // Paho MQTT Client
	subscriptions       map[string]*TopicSubscription // map of TopicSubscription for re-subscribing after reconnect
	tlsVerifyServerCert bool                          // verify the server certificate, this requires a Root CA signed cert
	tlsCACertFile       string                        // path to CA certificate
	updateMutex         *sync.Mutex                   // mutex for async updating of subscriptions
}

// TopicSubscription holds subscriptions to restore after disconnect
type TopicSubscription struct {
	topic     string
	handler   func(address string, message []byte)
	handlerID reflect.Value
	// token     pahomqtt.Token // for debugging
	// client *MqttMessenger //
}

// Connect to the MQTT broker
// If a previous connection exists then it is disconnected first.
// serverAddr contains the hostname:port of the server
// timeout in seconds after which to give up
func (messenger *MqttClient) Connect(clientID string, timeout int) error {
	// set config defaults
	// ClientID defaults to hostname-secondsSinceEpoc
	if clientID == "" {
		hostName, _ := os.Hostname()
		clientID = fmt.Sprintf("%s-%d", hostName, time.Now().Unix())
	}

	// close existing connection
	if messenger.pahoClient != nil && messenger.pahoClient.IsConnected() {
		messenger.pahoClient.Disconnect(10 * ConnectionTimeoutSec)
	}

	brokerURL := fmt.Sprintf("tls://%s/", messenger.hostPort) // tcp://host:1883 ws://host:1883 tls://host:8883, tcps://awshost:8883/mqtt
	// brokerURL := fmt.Sprintf("tls://mqtt.eclipse.org:8883/")
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(messenger.clientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetMaxReconnectInterval(60 * time.Second) // max wait 1 minute for a reconnect
	// Do not use MQTT persistence as not all brokers support it, and it causes problems on the broker if the client ID is
	// randomly generated. CleanSession disables persistence.
	opts.SetCleanSession(true)
	opts.SetKeepAlive(ConnectionTimeoutSec * time.Second) // pings to detect a disconnect. Use same as reconnect interval
	//opts.SetKeepAlive(60) // keepalive causes deadlock in v1.1.0. See github issue #126

	opts.SetOnConnectHandler(func(client pahomqtt.Client) {
		logrus.Warningf("MqttMessenger.onConnect: Connected to server at %s. Connected=%v. ClientId=%s",
			brokerURL, client.IsConnected(), messenger.clientID)
		// Subscribe to addresss already registered by the app on connect or reconnect
		messenger.resubscribe()
	})
	opts.SetConnectionLostHandler(func(client pahomqtt.Client, err error) {
		logrus.Warningf("MqttMessenger.onConnectionLost: Disconnected from server %s. Error %s, ClientId=%s",
			brokerURL, err, messenger.clientID)
	})
	// if lastWillAddress != "" {
	// 	opts.SetWill(lastWillAddress, lastWillValue, 1, false)
	// }
	// Use TLS if a CA certificate is given
	var rootCA *x509.CertPool
	if messenger.tlsCACertFile != "" {
		rootCA = x509.NewCertPool()
		caFile, err := ioutil.ReadFile(messenger.tlsCACertFile)
		if err != nil {
			logrus.Errorf("MqttMessenger.Connect: Unable to read CA certificate chain: %s", err)
		}
		rootCA.AppendCertsFromPEM([]byte(caFile))
	}
	opts.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: !messenger.tlsVerifyServerCert,
		RootCAs:            rootCA, // include the zcas cert in the host root ca set
		// https://opium.io/blog/mqtt-in-go/
		ServerName: "", // hostname on the server certificate. How to get this?
	})

	logrus.Infof("MqttMessenger.Connect: Connecting to MQTT server: %s with clientID %s"+
		" AutoReconnect and CleanSession are set.",
		brokerURL, messenger.clientID)

	// FIXME: PahoMqtt disconnects when sending a lot of messages, like on startup of some adapters.
	messenger.pahoClient = pahomqtt.NewClient(opts)

	// start listening for messages
	messenger.isRunning = true
	//go messenger.messageChanLoop()

	// Auto reconnect doesn't work for initial attempt: https://github.com/eclipse/paho.mqtt.golang/issues/77
	retryDelaySec := 1
	retryDuration := 0
	var err error
	for timeout == 0 || retryDuration < timeout {
		token := messenger.pahoClient.Connect()
		token.Wait()
		// Wait to give connection time to settle. Sending a lot of messages causes the connection to fail. Bug?
		time.Sleep(1000 * time.Millisecond)
		err = token.Error()
		if err == nil {
			break
		}
		retryDuration++

		logrus.Errorf("MqttMessenger.Connect: Connecting to broker on %s failed: %s. retrying in %d seconds.",
			brokerURL, token.Error(), retryDelaySec)
		sleepDuration := time.Duration(retryDelaySec)
		retryDuration += int(sleepDuration)
		time.Sleep(sleepDuration * time.Second)
		// slowly increment wait time
		if retryDelaySec < 120 {
			retryDelaySec++
		}
	}
	return err
}

// Disconnect from the MQTT broker and unsubscribe from all addresss and set
// device state to disconnected
func (messenger *MqttClient) Disconnect() {
	messenger.updateMutex.Lock()
	messenger.isRunning = false
	messenger.updateMutex.Unlock()

	if messenger.pahoClient != nil {
		logrus.Warningf("MqttMessenger.Disconnect: Set state to disconnected and close connection")
		//messenger.publish("$state", "disconnected")
		time.Sleep(time.Second / 10) // Disconnect doesn't seem to wait for all messages. A small delay ahead helps
		messenger.pahoClient.Disconnect(10 * ConnectionTimeoutSec * 1000)
		messenger.pahoClient = nil

		messenger.subscriptions = make(map[string]*TopicSubscription, 0)
		//close(messenger.messageChannel)     // end the message handler loop
	}
}

// Wrapper for message handling to support multiple subscribers to one topic
func (messenger *MqttClient) onMessage(c pahomqtt.Client, msg pahomqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	logrus.Infof("MqttMessenger.onMessage. address=%s", topic)
	subscription := messenger.subscriptions[topic]
	if subscription == nil {
		logrus.Errorf("onMessage: no subscription for topic %s", topic)
		return
	}
	subscription.handler(topic, payload)
}

// Publish a message to a topic address
func (messenger *MqttClient) Publish(topic string, message []byte) error {
	var err error

	if messenger.pahoClient == nil || !messenger.pahoClient.IsConnected() {
		logrus.Warnf("MqttMessenger.Publish: Unable to publish. No connection with server.")
		return errors.New("no connection with server")
	}
	logrus.Debugf("MqttMessenger.Publish []byte: topic=%s, qos=%d",
		topic, messenger.pubQos)
	token := messenger.pahoClient.Publish(topic, messenger.pubQos, false, message)

	err = token.Error()
	if err != nil {
		// TODO: confirm that with qos=1 the message is sent after reconnect
		logrus.Warnf("MqttMessenger.Publish: Error during publish on address %s: %v", topic, err)
		//return err
	}
	return err
}

// subscribe to addresss after establishing connection
// The application can already subscribe to addresss before the connection is established. If connection is lost then
// this will re-subscribe to those addresss as PahoMqtt drops the subscriptions after disconnect.
//
func (messenger *MqttClient) resubscribe() {
	// prevent simultaneous access to subscriptions
	messenger.updateMutex.Lock()
	defer messenger.updateMutex.Unlock()

	logrus.Infof("MqttMessenger.resubscribe to %d addresess", len(messenger.subscriptions))
	for topic := range messenger.subscriptions {
		// clear existing subscription in case it is still there
		messenger.pahoClient.Unsubscribe(topic)

		logrus.Infof("MqttMessenger.resubscribe: address %s", topic)
		// create a new variable to hold the subscription in the closure
		messenger.pahoClient.Subscribe(topic, messenger.pubQos, messenger.onMessage)
		// token := messenger.pahoClient.Subscribe(newSubscr.topic, messenger.pubQos, newSubscr.onMessage)
		//token := messenger.pahoClient.Subscribe(newSubscr.address, newSubscr.qos, func (c pahomqtt.Client, msg pahomqtt.Message) {
		//logrus.Infof("mqtt.resubscribe.onMessage: address %s, subscription %s", msg.Topic(), newSubscr.address)
		//newSubscr.onMessage(c, msg)
		//})
		// newSubscr.token = token
	}
	logrus.Infof("MqttMessenger.resubscribe complete")
}

// Subscribe to a address
// If a subscription already exists, it is replaced.
// topic: address to subscribe to. This supports mqtt wildcards such as + and #
// handler: callback handler.
func (messenger *MqttClient) Subscribe(
	topic string, handler func(address string, message []byte)) {
	handlerID := reflect.ValueOf(handler)
	subscription := &TopicSubscription{
		topic:     topic,
		handler:   handler,
		handlerID: handlerID,
	}
	logrus.Infof("MqttMessenger.Subscribe: topic %s, qos %d", topic, messenger.subQos)

	messenger.updateMutex.Lock()
	defer messenger.updateMutex.Unlock()

	if messenger.subscriptions[topic] != nil {
		logrus.Warningf("Subscribe: Existing subscription to %s is replaced", topic)
	} else {
		if messenger.pahoClient != nil {
			messenger.pahoClient.Subscribe(topic, messenger.subQos, messenger.onMessage) //func(c pahomqtt.Client, msg pahomqtt.Message) {
		}
	}
	messenger.subscriptions[topic] = subscription
}

// Unsubscribe a topic and handler
// if handler is nil then all subscribers to the topic are removed
func (messenger *MqttClient) Unsubscribe(topic string) {
	// messenger.publishMutex.Lock()
	subscription := messenger.subscriptions[topic]
	if subscription == nil {
		// nothing to unsubscribe
		logrus.Warningf("Unsubscribe: Subscription on topic %s didn't exist. Ignored", topic)
		return
	}

	messenger.pahoClient.Unsubscribe(topic)
	messenger.subscriptions[topic] = nil
}

// NewMqttMessenger creates a new MQTT messenger instance
// clientID must be unique
// certFolder must contain the server certificate mqtt_srv.crt
func NewMqttMessenger(certFolder string, hostPort string) *MqttClient {

	messenger := &MqttClient{
		hostPort:      hostPort,
		pubQos:        1,
		subQos:        1,
		pahoClient:    nil,
		subscriptions: make(map[string]*TopicSubscription, 0),
		//messageChannel: make(chan *IncomingMessage),
		tlsCACertFile:       certFolder + "/mqtt_srv.crt",
		tlsVerifyServerCert: true,
		updateMutex:         &sync.Mutex{},
	}
	return messenger
}
