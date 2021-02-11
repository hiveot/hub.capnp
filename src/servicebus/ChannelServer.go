// Package servicebus for servicing channel connections
package servicebus

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Default nr of messages that can be queued per channel before the sender blocks
const defaultChannelQueueDepth = 10

// Connection headers
const (
	AuthorizationHeader = "Authorization"
	ClientHeader        = "Client"
)

// Mux variables of incoming connections
const (
	MuxChannel = "channel"
	MuxStage   = "stage"
)

// publisher and subscribe stages used in establishing connections
const (
	PublisherStage  = "pub"
	SubscriberStage = "sub"
)

// ChannelServer provides a lightweight server for managing multiple pub/sub channels
type ChannelServer struct {
	httpServer    *http.Server
	channels      map[string]*Channel // channels by ID
	updateMutex   *sync.Mutex
	upgrader      websocket.Upgrader // use default options
	ServerCertPEM []byte             // server certificate PEM. For testing
}

// Channel holding publisher and subscriber connections to this channel
type Channel struct {
	ID           string            // ID of the channel
	publishers   []*websocket.Conn // publisher connections
	subscribers  []*websocket.Conn // subscriber connections
	jobQueue     chan []byte       // job queue for processing
	MessageCount int32             // nr of published messages
}

// AddPublisher adds a publisher to a channel
// c contains the connection to add
func (cs *ChannelServer) AddPublisher(channel *Channel, c *websocket.Conn) *Channel {
	cs.updateMutex.Lock()
	defer cs.updateMutex.Unlock()
	channel.publishers = append(channel.publishers, c)
	return channel
}

// AddSubscriber adds a subscriber connection to a channel
// c contains the connection to add
func (cs *ChannelServer) AddSubscriber(channel *Channel, c *websocket.Conn) *Channel {
	cs.updateMutex.Lock()
	defer cs.updateMutex.Unlock()
	channel.subscribers = append(channel.subscribers, c)
	return channel
}

// Authenticate the connectino and return the clientID
func (cs *ChannelServer) authenticateConnection(request *http.Request) (string, error) {
	clientID, authToken, ok := request.BasicAuth()
	_ = authToken

	if !ok {
		return clientID, errors.New("Invalid authorization header for client '" + clientID + "'")
	}
	return clientID, nil
}

// read a message from the channel's job queue and send it to subscribers
// this function never returns, it reads as long as the channel queue exists
func (cs *ChannelServer) channelMessageWorker(channel *Channel) {
	for message := range channel.jobQueue {
		cs.processChannelMessage(channel.ID, message)
	}
}

// GetPublishers returns a list of channel subscriber connections
func (cs *ChannelServer) GetPublishers(channelID string) []*websocket.Conn {
	cs.updateMutex.Lock()
	defer cs.updateMutex.Unlock()
	channel := cs.channels[channelID]
	if channel == nil {
		return make([]*websocket.Conn, 0)
	}
	return channel.publishers
}

// GetSubscribers returns a list of channel subscriber connections
func (cs *ChannelServer) GetSubscribers(channelID string) []*websocket.Conn {
	cs.updateMutex.Lock()
	defer cs.updateMutex.Unlock()
	channel := cs.channels[channelID]
	if channel == nil {
		return make([]*websocket.Conn, 0)
	}
	return channel.subscribers
}

// GetChannel returns the channel for the given ID
func (cs *ChannelServer) GetChannel(channelID string) *Channel {
	cs.updateMutex.Lock()
	defer cs.updateMutex.Unlock()
	channel := cs.channels[channelID]
	return channel
}

// NewChannel creates a new container for channel connections
// channelID is the unique ID of the channel
// queueDepth is the number of messages that can be queued
func (cs *ChannelServer) NewChannel(channelID string, queueDepth int) *Channel {
	channel := &Channel{
		ID:          channelID,
		publishers:  make([]*websocket.Conn, 0),
		subscribers: make([]*websocket.Conn, 0),
		jobQueue:    make(chan []byte, queueDepth),
	}
	cs.updateMutex.Lock()
	defer cs.updateMutex.Unlock()
	cs.channels[channelID] = channel

	return channel
}

// processChannelMessage passes a message to all subscribers
// If a subscriber fails, it is removed from the channel
func (cs *ChannelServer) processChannelMessage(channelID string, message []byte) {
	consumers := cs.GetSubscribers(channelID)
	// logrus.Infof("processChannelMessage: Sending message to %d subscribers of channel %s", len(consumers), channelID)
	for _, c := range consumers {
		err := cs.sendMessage(c, message)
		if err != nil {
			logrus.Warningf("processChannelMessage: failed sending 1 message to a subscriber of %s", channelID)
			cs.RemoveConnection(c)
		}
	}
}

// receiveMessage from a socket connection
// Wait until message received or timeoutSec has passed, use 0 to wait indefinitely
// Note that if the connection times out, the connectino must be discarded.
// func (cs *ChannelServer) receiveMessage(connection *websocket.Conn, timeout time.Duration) ([]byte, error) {
// 	connection.SetReadDeadline(time.Now().Add(timeout))
// 	_, message, err := connection.ReadMessage()
// 	return message, err
// }

// RemoveConnection a channel connection while retaining order
// Returns true if remove successful, false if connection not found
func (cs *ChannelServer) RemoveConnection(c *websocket.Conn) bool {
	cs.updateMutex.Lock()
	defer cs.updateMutex.Unlock()
	// slow way is okay as there won't be that many connections and they don't change often
	for _, channel := range cs.channels {
		for index, connection := range channel.subscribers {
			if connection == c {
				if index == len(channel.subscribers)-1 {
					channel.subscribers = channel.subscribers[:index]
				} else {
					channel.subscribers = append(channel.subscribers[:index], channel.subscribers[index+1:]...)
				}
				return true
			}
		}
		for index, connection := range channel.publishers {
			if connection == c {
				if index == len(channel.publishers)-1 {
					channel.publishers = channel.publishers[:index]
				} else {
					channel.publishers = append(channel.publishers[:index], channel.publishers[index+1:]...)
				}
				return true
			}
		}
	}
	return false
}

// sendMessage publishes a message into the channel
func (cs *ChannelServer) sendMessage(connection *websocket.Conn, message []byte) error {
	w, err := connection.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.Write(message)
	w.Close()
	return nil
}

// ServeChannel handles new pub/sub channel connections
// The http header must contain a client ID otherwise the connection will be rejected.
func (cs *ChannelServer) ServeChannel(response http.ResponseWriter, request *http.Request) {
	chID := mux.Vars(request)[MuxChannel]
	pubOrSub := mux.Vars(request)[MuxStage]

	clientID, err := cs.authenticateConnection(request)
	if err != nil {
		http.Error(response, "Invalid client authorization", 401)
		logrus.Warningf("ServeChannel: Unauthorized. Rejected connection from client '%s' for channel %s", clientID, chID)
		return
	}
	logrus.Infof("ServeChannel accepted incoming connection for channel %s, stage %s", chID, pubOrSub)

	// upgrade the HTTP connection to a websocket connection
	c, err := cs.upgrader.Upgrade(response, request, nil)
	if err != nil {
		logrus.Warningf("ServeChannel upgrade error for client %s: %s", clientID, err)
		return
	}
	// logrus.Warningf("ServeChannel accepted connection from client %s", clientID)

	channel := cs.GetChannel(chID)
	if channel == nil {
		// Create a new channel with worker to process messages
		channel = cs.NewChannel(chID, defaultChannelQueueDepth)
		go cs.channelMessageWorker(channel)
	}
	if pubOrSub == SubscriberStage {
		// record subscriber connections
		cs.AddSubscriber(channel, c)
	} else { // else if pubOrSub == PublisherStage {
		// Also read on subscribe connections to detect if the client has closed it
		cs.AddPublisher(channel, c)
		// publisher connections are closed on exit
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Errorf("ServeChannel connection error to %s/%s by client: %s", chID, pubOrSub, err)
				} else {
					logrus.Warningf("ServeChannel connection to %s/%s dropped by client: %s", chID, pubOrSub, err)
				}
				cs.RemoveConnection(c)
				break
			}
			atomic.AddInt32(&channel.MessageCount, 1)
			logrus.Infof("ServeChannel received publication on channel ID %s", chID)
			channel.jobQueue <- message
		}
	}
}

// Start the server and listen for incoming connection on /channel/#
// Returns the mux router to allow for additional listeners such as /home
func (cs *ChannelServer) Start(host string) (*mux.Router, error) {
	var err error
	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/channel/{%s}/{%s}", MuxChannel, MuxStage), cs.ServeChannel)

	go func() {
		// cs.updateMutex.Lock()
		cs.httpServer = &http.Server{
			Addr:    host,
			Handler: router,
		}
		// cs.updateMutex.Unlock()
		err = cs.httpServer.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Start: ListenAndServe error ", err)
		}
	}()
	return router, err
}

// StartTLS starts the server and listen for incoming connection on /channel/# using TLS
// This expects a certfile and keyfile.
// listenAddress is the address and port the server listens on
// Returns the mux router to allow for additional listeners such as /home
func (cs *ChannelServer) StartTLS(listenAddress string, caCertFile string, serverCertFile string, serverKeyFile string) (*mux.Router, error) {
	logrus.Infof("StartTLS: Serving on address %s", listenAddress)
	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/channel/{%s}/{%s}", MuxChannel, MuxStage), cs.ServeChannel)

	// The server certificate and key is needed
	cs.ServerCertPEM, _ = ioutil.ReadFile(serverCertFile)
	serverKeyPEM, _ := ioutil.ReadFile(serverKeyFile)
	serverCert, err := tls.X509KeyPair(cs.ServerCertPEM, serverKeyPEM)
	if err != nil {
		return router, err
	}
	// To verify clients, the client CA must be provided
	caCertPEM, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return router, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)

	serverTLSConf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAnyClientCert,
		ClientCAs:    caCertPool,
	}

	cs.httpServer = &http.Server{
		Addr: listenAddress,
		// ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		// WriteTimeout: 10 * time.Second,
		Handler:   router,
		TLSConfig: serverTLSConf,
	}
	go func() {
		err := cs.httpServer.ListenAndServeTLS("", "")
		// err := cs.httpServer.ListenAndServeTLS(serverCertFile, serverKeyFile)
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Start: ListenAndServeTLS error ", err)
		}
	}()
	return router, err
}

// Stop the server and close all connections
func (cs *ChannelServer) Stop() {
	// cs.updateMutex.Lock()
	// defer cs.updateMutex.Unlock()

	cs.httpServer.Shutdown(nil)
	for _, channel := range cs.channels {
		for _, pubs := range channel.publishers {
			pubs.Close()
		}
		for _, subs := range channel.subscribers {
			subs.Close()
		}
	}
}

// NewChannelServer creates an instance of the lightweight channel server
func NewChannelServer() *ChannelServer {
	p := &ChannelServer{
		channels:    make(map[string]*Channel, 0),
		upgrader:    websocket.Upgrader{},
		updateMutex: &sync.Mutex{},
	}
	return p
}
