// Package smbserver with simple internal message bus for serving plugins pub/sub
package internal

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/messaging/smbus"
)

// Default nr of messages that can be queued per channel before the sender blocks
const defaultChannelQueueDepth = 10

// ServeSmbus provides a lightweight server for managing multiple pub/sub channels
type ServeSmbus struct {
	httpServer  *http.Server
	connections []*websocket.Conn // pub/sub connections

	subscriptions map[string]*Channel // channel subscriptions by channel ID
	updateMutex   *sync.Mutex
	upgrader      websocket.Upgrader // use default options
	ServerCertPEM []byte             // server certificate PEM. For testing
}

// Channel holding subscription handlers and job queue for a channel
type Channel struct {
	ID           string // ID of the channel
	subscribers  []*websocket.Conn
	jobQueue     chan []byte // job queue for processing
	MessageCount int32       // nr of published messages on this channel
}

// AddConnection adds a new client connection
// c contains the connection to add
func (mbs *ServeSmbus) AddConnection(c *websocket.Conn) {
	mbs.updateMutex.Lock()
	defer mbs.updateMutex.Unlock()
	logrus.Infof("ServeMsgBus.AddConnection: adding connection from '%s'", c.RemoteAddr().String())
	mbs.connections = append(mbs.connections, c)
}

// AddChannel setups a job queue with listener to send messages to subscribers
// If the channel already exists this is ignored
//  channelID to add
func (mbs *ServeSmbus) AddChannel(channelID string) *Channel {
	mbs.updateMutex.Lock()
	defer mbs.updateMutex.Unlock()

	channel := mbs.subscriptions[channelID]
	if channel == nil {
		logrus.Infof("ServeMsgBus.AddChannel: adding new channel '%s'", channelID)
		channel = &Channel{
			ID:          channelID,
			subscribers: make([]*websocket.Conn, 0),
			jobQueue:    make(chan []byte, defaultChannelQueueDepth),
		}
		// worker to process channel messages
		go func() {
			for message := range channel.jobQueue {
				mbs.sendChannelMessageToSubscribers(channel.ID, message)
			}
		}()
	}
	return channel
}

// AddSubscriber creates a new subscription for the given channel
// channelID is the unique ID of the channel
// conn is the websocket connection that subscribes to the channel
func (mbs *ServeSmbus) AddSubscriber(channelID string, conn *websocket.Conn) *Channel {
	logrus.Infof("ServeMsgBus.AddSubscriber: adding subscription to channel '%s'", channelID)

	channel := mbs.AddChannel(channelID)

	mbs.updateMutex.Lock()
	defer mbs.updateMutex.Unlock()
	mbs.subscriptions[channelID] = channel
	channel.subscribers = append(channel.subscribers, conn)

	return channel
}

// GetChannel returns the channel for the given ID
// This returns nil if the channel doesn't exist
func (mbs *ServeSmbus) GetChannel(channelID string) *Channel {
	mbs.updateMutex.Lock()
	defer mbs.updateMutex.Unlock()
	channel := mbs.subscriptions[channelID]
	return channel
}

// GetConnections returns a list of client connections
// func (mbs *ServeMsgBus) GetConnections() []*websocket.Conn {
// 	mbs.updateMutex.Lock()
// 	defer mbs.updateMutex.Unlock()
// 	connectionList := mbs.connections
// 	return connectionList
// }

// GetSubscribers that have subscribed to a channel
// This returns a shallow copy of subscriber list
func (mbs *ServeSmbus) GetSubscribers(channelID string) []*websocket.Conn {
	mbs.updateMutex.Lock()
	defer mbs.updateMutex.Unlock()
	channel := mbs.subscriptions[channelID]
	if channel == nil {
		return make([]*websocket.Conn, 0)
	}
	subscribers := append([]*websocket.Conn(nil), channel.subscribers...) // copy the list, yuk
	return subscribers
}

// PublishToSubscribers sends a message to all subscribers of a channel
// This handles the message through a worker thread
func (mbs *ServeSmbus) PublishToSubscribers(channelID string, message []byte) {
	ch := mbs.GetChannel(channelID)
	if ch == nil {
		logrus.Warningf("ServeMsgBus.PublishToSubscribers: No subscribers for channel %s", channelID)
	} else {
		atomic.AddInt32(&ch.MessageCount, 1)
		ch.jobQueue <- message
	}
}

// Remove a connection from a list of connections
// Returns a new list, or the old list if the connection wasn't in the list
func removeConnectionFromList(clist []*websocket.Conn, c *websocket.Conn) []*websocket.Conn {

	// slow way is okay as there won't be that many connections and they don't change often
	for index, connection := range clist {
		if connection == c {
			if index == len(clist)-1 {
				clist = clist[:index]
			} else {
				clist = append(clist[:index], clist[index+1:]...)
			}
			return clist
		}
	}
	return clist
}

// RemoveConnection remove connection from subscriptions
// The caller must make sure it is closed
func (mbs *ServeSmbus) RemoveConnection(c *websocket.Conn) {
	logrus.Infof("ServeMsgBus.RemoveConnection: Removing closed connection")
	mbs.updateMutex.Lock()
	defer mbs.updateMutex.Unlock()

	mbs.connections = removeConnectionFromList(mbs.connections, c)

	for _, channel := range mbs.subscriptions {
		channel.subscribers = removeConnectionFromList(channel.subscribers, c)
	}
}

// RemoveSubscriber a connection from a channel
func (mbs *ServeSmbus) RemoveSubscriber(channelID string, c *websocket.Conn) {
	logrus.Infof("ServeMsgBus.RemoveSubscriber: Remove subscription to channel '%s'", channelID)

	channel := mbs.GetChannel(channelID)
	if channel != nil {
		mbs.updateMutex.Lock()
		defer mbs.updateMutex.Unlock()
		channel.subscribers = removeConnectionFromList(channel.subscribers, c)
	}
}

// sendChannelMessageToSubscribers passes a message to all subscribers of a channel
// If a connection fails, it is removed
func (mbs *ServeSmbus) sendChannelMessageToSubscribers(channelID string, message []byte) {
	consumers := mbs.GetSubscribers(channelID)

	logrus.Infof("ServeMsgBus.sendChannelMessageToSubscribers send message to %d subscribers of channel %s", len(consumers), channelID)
	// logrus.Infof("processChannelMessage: Sending message to %d subscribers of channel %s", len(consumers), channelID)
	// logrus.Infof("--- sending message to %d subscribers of channel %s", len(consumers), channelID)
	for _, c := range consumers {
		err := smbus.Send(c, smbus.MsgBusCommandReceive, channelID, message)
		if err != nil {
			logrus.Warningf("ServeMsgBus.processChannelMessage: failed sending 1 message to a subscriber of %s", channelID)
			mbs.RemoveConnection(c)
		}
	}
}

// ServeConnection handles new pub/sub connections
// The http header must contain a client ID otherwise the connection will be rejected.
func (mbs *ServeSmbus) serveConnection(response http.ResponseWriter, request *http.Request) {
	// chID := mux.Vars(request)[MuxChannel]
	// pubOrSub := mux.Vars(request)[MuxStage]

	// clientID, err := cs.authenticateConnection(request)
	clientID := request.Header.Get(smbus.ClientHeader)
	if clientID == "" {
		http.Error(response, "Invalid client. A clientID is required.", 401)
		logrus.Warningf("ServeMsgBus.serveConnection: Missing clientID from client '%s'", request.RemoteAddr)
		return
	}
	logrus.Infof("ServeMsgBus.serveConnection accepted incoming connection from %s", clientID)

	// upgrade the HTTP connection to a websocket connection
	c, err := mbs.upgrader.Upgrade(response, request, nil)
	if err != nil {
		http.Error(response, err.Error(), 401)
		logrus.Warningf("ServeMsgBus.serveConnection upgrade error for client %s: %s", clientID, err)
		return
	}
	// logrus.Warningf("ServeChannel accepted connection from client %s", clientID)
	mbs.AddConnection(c)

	// channel = cs.NewChannel(chID, defaultChannelQueueDepth)
	// go cs.channelMessageWorker(channel)

	//listen connections are closed on exit
	go func() {
		smbus.Listen(c, func(command string, topic string, data []byte) {
			if data == nil {
				mbs.RemoveConnection(c)
			} else if command == smbus.MsgBusCommandPublish {
				mbs.PublishToSubscribers(topic, data)
			} else if command == smbus.MsgBusCommandSubscribe {
				mbs.AddSubscriber(topic, c)
			} else if command == smbus.MsgBusCommandUnsubscribe {
				mbs.RemoveSubscriber(topic, c)
			} else {
				logrus.Warningf("ServeMsgBus.ServeConnection: Ignored unknown command '%s'", command)
			}
		})
		// c.Close()
	}()
}

// Start the server and listen for incoming connections
// Returns the mux router to allow for additional listeners such as /home
func (mbs *ServeSmbus) Start(host string) (*mux.Router, error) {
	var err error
	router := mux.NewRouter()
	router.HandleFunc(smbus.MsgbusAddress, mbs.serveConnection)

	go func() {
		// cs.updateMutex.Lock()
		mbs.httpServer = &http.Server{
			Addr:    host,
			Handler: router,
		}
		// cs.updateMutex.Unlock()
		logrus.Infof("ServeMsgBus.Start: ListenAndServe on %s", host)
		err = mbs.httpServer.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Start: ListenAndServe error: %s", err)
		}
	}()

	// FIXME: this doesn't catch error "bind: address already in use"
	//  Wait until running or failure?

	return router, err
}

// StartTLS starts listing for incoming connection on via TLS.
// This uses both client and server certificates
//  listenAddress is the address and port the server listens on
//  caCertFile path to CA certificate, required
//  serverCertFile path to MsgBusServer certificate, required
//  serverKeyFile path to MsgBusServer private key, required
// Returns the mux router to allow for additional listeners such as /home
func (mbs *ServeSmbus) StartTLS(listenAddress string, caCertFile string, serverCertFile string,
	serverKeyFile string) (router *mux.Router, err error) {

	logrus.Infof("ServeMsgBus.StartTLS: Serving on address %s", listenAddress)

	router = mux.NewRouter()
	router.HandleFunc(smbus.MsgbusAddress, mbs.serveConnection)

	// The server certificate and key is needed
	mbs.ServerCertPEM, err = ioutil.ReadFile(serverCertFile)
	serverKeyPEM, err2 := ioutil.ReadFile(serverKeyFile)
	serverCert, err3 := tls.X509KeyPair(mbs.ServerCertPEM, serverKeyPEM)
	if err != nil || err2 != nil || err3 != nil {
		logrus.Errorf("StartTLS: certificates not found")
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

	mbs.httpServer = &http.Server{
		Addr: listenAddress,
		// ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		// WriteTimeout: 10 * time.Second,
		Handler:   router,
		TLSConfig: serverTLSConf,
	}
	go func() {
		err := mbs.httpServer.ListenAndServeTLS("", "")
		// err := cs.httpServer.ListenAndServeTLS(serverCertFile, serverKeyFile)
		if err != nil && err != http.ErrServerClosed {
			logrus.Errorf("ServeMsgBus.Start: ListenAndServeTLS error: %s", err)
		}
	}()
	return router, err
}

// Stop the server and close all connections
func (mbs *ServeSmbus) Stop() {
	// cs.updateMutex.Lock()
	// defer cs.updateMutex.Unlock()
	logrus.Warningf("ServeMsgBus.Stop: Stopping message bus server")
	mbs.updateMutex.Lock()
	defer mbs.updateMutex.Unlock()

	for _, c := range mbs.connections {
		c.Close()
	}
	mbs.connections = make([]*websocket.Conn, 0)
	mbs.subscriptions = make(map[string]*Channel)
	if mbs.httpServer != nil {
		mbs.httpServer.Shutdown(nil)
	}
}

// NewServeMsgBus creates an instance of the simple internal message bus server
func NewServeMsgBus() *ServeSmbus {
	p := &ServeSmbus{
		connections:   make([]*websocket.Conn, 0),
		subscriptions: make(map[string]*Channel),
		upgrader:      websocket.Upgrader{},
		updateMutex:   &sync.Mutex{},
	}
	return p
}
