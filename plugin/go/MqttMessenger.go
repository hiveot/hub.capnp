// Package plugin with connection management for the gateway that use MQTT to communicate
package plugin

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// ConnectionTimeoutSec constant with connection and reconnection timeouts
const ConnectionTimeoutSec = 20

// TLSPort is the default secure port to connect to mqtt
const TLSPort = 8883

// MqttMessenger that implements IMessenger
type MqttMessenger struct {
	clientID      string // unique ID of the client
	serverAddress string // host:port of server
	pubQos        byte
	subQos        byte
	//
	isRunning           bool                // listen for messages while running
	pahoClient          pahomqtt.Client     // Paho MQTT Client
	subscriptions       []TopicSubscription // list of TopicSubscription for re-subscribing after reconnect
	tlsVerifyServerCert bool                // verify the server certificate, this requires a Root CA signed cert
	tlsCACertFile       string              // path to CA certificate
	updateMutex         *sync.Mutex         // mutex for async updating of subscriptions
}

// TopicSubscription holds subscriptions to restore after disconnect
type TopicSubscription struct {
	address string
	handler func(address string, message []byte)
	token   pahomqtt.Token // for debugging
	client  *MqttMessenger //
}

// Connect to the MQTT broker
// If a previous connection exists then it is disconnected first.
// serverAddr contains the hostname:port of the server
func (messenger *MqttMessenger) Connect(serverAddr string) error {

	messenger.serverAddress = serverAddr
	// close existing connection
	if messenger.pahoClient != nil && messenger.pahoClient.IsConnected() {
		messenger.pahoClient.Disconnect(10 * ConnectionTimeoutSec)
	}

	// set config defaults
	// ClientID defaults to hostname-secondsSinceEpoc
	hostName, _ := os.Hostname()
	if messenger.clientID == "" {
		messenger.clientID = fmt.Sprintf("%s-%d", hostName, time.Now().Unix())
	}

	brokerURL := fmt.Sprintf("tls://%s/", messenger.serverAddress) // tcp://host:1883 ws://host:1883 tls://host:8883, tcps://awshost:8883/mqtt
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
	for {
		token := messenger.pahoClient.Connect()
		token.Wait()
		// Wait to give connection time to settle. Sending a lot of messages causes the connection to fail. Bug?
		time.Sleep(1000 * time.Millisecond)
		err := token.Error()
		if err == nil {
			break
		}

		logrus.Errorf("MqttMessenger.Connect: Connecting to broker on %s failed: %s. retrying in %d seconds.",
			brokerURL, token.Error(), retryDelaySec)
		time.Sleep(time.Duration(retryDelaySec) * time.Second)
		// slowly increment wait time
		if retryDelaySec < 120 {
			retryDelaySec++
		}
	}
	return nil
}

// Disconnect from the MQTT broker and unsubscribe from all addresss and set
// device state to disconnected
func (messenger *MqttMessenger) Disconnect() {
	messenger.updateMutex.Lock()
	messenger.isRunning = false
	messenger.updateMutex.Unlock()

	if messenger.pahoClient != nil {
		logrus.Warningf("MqttMessenger.Disconnect: Set state to disconnected and close connection")
		//messenger.publish("$state", "disconnected")
		time.Sleep(time.Second / 10) // Disconnect doesn't seem to wait for all messages. A small delay ahead helps
		messenger.pahoClient.Disconnect(10 * ConnectionTimeoutSec * 1000)
		messenger.pahoClient = nil

		messenger.subscriptions = nil
		//close(messenger.messageChannel)     // end the message handler loop
	}
}

// Wrapper for message handling.
// Use a channel to handle the message in a gorouting.
// This fixes a problem with losing context in callbacks. Not sure what is going on though.
func (subscription *TopicSubscription) onMessage(c pahomqtt.Client, msg pahomqtt.Message) {
	// NOTE: Scope in this callback is not always retained. Pipe notifications through a channel and handle in goroutine
	address := msg.Topic()
	payload := msg.Payload()

	logrus.Infof("MqttMessenger.onMessage. address=%s, subscription=%s", address, subscription.address)
	subscription.handler(address, payload)
}

// Publish value using the device address as base
// address to publish on.
// payload is converted to string if it isn't a byte array, as Paho doesn't handle int and bool
func (messenger *MqttMessenger) Publish(topic string, message []byte) error {
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
func (messenger *MqttMessenger) resubscribe() {
	// prevent simultaneous access to subscriptions
	messenger.updateMutex.Lock()
	defer messenger.updateMutex.Unlock()

	logrus.Infof("MqttMessenger.resubscribe to %d addresess", len(messenger.subscriptions))
	for _, subscription := range messenger.subscriptions {
		// clear existing subscription
		messenger.pahoClient.Unsubscribe(subscription.address)

		logrus.Infof("MqttMessenger.resubscribe: address %s", subscription.address)
		// create a new variable to hold the subscription in the closure
		newSubscr := subscription
		token := messenger.pahoClient.Subscribe(newSubscr.address, messenger.pubQos, newSubscr.onMessage)
		//token := messenger.pahoClient.Subscribe(newSubscr.address, newSubscr.qos, func (c pahomqtt.Client, msg pahomqtt.Message) {
		//logrus.Infof("mqtt.resubscribe.onMessage: address %s, subscription %s", msg.Topic(), newSubscr.address)
		//newSubscr.onMessage(c, msg)
		//})
		newSubscr.token = token
	}
	logrus.Infof("MqttMessenger.resubscribe complete")
}

// Subscribe to a address
// Subscribers are automatically resubscribed after the connection is restored
// If no connection exists, then subscriptions are stored until a connection is established.
// address: address to subscribe to. This can contain wildcards.
// qos: Quality of service for subscription: 0, 1, 2
// handler: callback handler.
func (messenger *MqttMessenger) Subscribe(
	topic string, handler func(address string, message []byte)) error {
	subscription := TopicSubscription{
		address: topic,
		handler: handler,
		token:   nil,
		client:  messenger,
	}
	messenger.updateMutex.Lock()
	defer messenger.updateMutex.Unlock()
	messenger.subscriptions = append(messenger.subscriptions, subscription)

	logrus.Infof("MqttMessenger.Subscribe: topic %s, qos %d", topic, messenger.subQos)
	//messenger.pahoClient.Subscribe(address, qos, addressSubscription.onMessage) //func(c pahomqtt.Client, msg pahomqtt.Message) {
	if messenger.pahoClient != nil {
		messenger.pahoClient.Subscribe(topic, messenger.subQos, subscription.onMessage) //func(c pahomqtt.Client, msg pahomqtt.Message) {
	}
	return nil
}

// Unsubscribe an address and handler
// if handler is nil then only the address needs to match
func (messenger *MqttMessenger) Unsubscribe(
	address string, handler func(address string, message []byte)) {
	// messenger.publishMutex.Lock()
	var onMessageID = handler
	// onMessageStr := fmt.Sprintf("%v", &onMessage)
	for i, sub := range messenger.subscriptions {
		// can't compare addresses directly so convert to string
		// handlerStr := fmt.Sprintf("%v", &sub.handler)
		var handlerID = sub.handler
		// if sub.address == address && handlerStr == onMessageStr {
		if sub.address == address && (handler == nil || &onMessageID == &handlerID) {
			// shift remainder left one index
			if i < len(messenger.subscriptions) {
				copy(messenger.subscriptions[i:], messenger.subscriptions[i+1:])
			}
			messenger.subscriptions = messenger.subscriptions[:len(messenger.subscriptions)-1]
			if handler != nil {
				break
			}
		}
	}
	// messenger.publishMutex.Unlock()
}

// NewMqttMessenger creates a new MQTT messenger instance
// clientID must be unique
// certFolder must contain the CaCertFile
func NewMqttMessenger(clientID string, certFolder string) *MqttMessenger {
	messenger := &MqttMessenger{
		clientID:   clientID,
		pubQos:     1,
		subQos:     1,
		pahoClient: nil,
		//messageChannel: make(chan *IncomingMessage),
		tlsCACertFile:       certFolder + "/" + CaCertFile,
		tlsVerifyServerCert: true,
		updateMutex:         &sync.Mutex{},
	}
	return messenger
}
