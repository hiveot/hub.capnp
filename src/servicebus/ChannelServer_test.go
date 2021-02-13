package servicebus_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/src/lib"
	"github.com/wostzone/gateway/src/servicebus"
	"golang.org/x/net/http2"
)

const channel1ID = "channel1"
const channel2ID = "channel2"
const defaultBufferSize = 1

// const hostPort = "localhost:9678"

const client1ID = "plugin1"

// Test create, store and remove channels by the server
func TestCreateChannel(t *testing.T) {
	logrus.Info("Testing create channels")
	srv := servicebus.NewChannelServer()
	c1 := &websocket.Conn{}
	c2 := &websocket.Conn{}
	c3 := &websocket.Conn{}
	c4 := &websocket.Conn{}
	channel1 := srv.NewChannel(channel1ID, defaultBufferSize)
	srv.AddSubscriber(channel1, c1)
	srv.AddSubscriber(channel1, c2)
	srv.AddPublisher(channel1, c3)
	srv.AddPublisher(channel1, c4)

	clist1 := srv.GetSubscribers(channel1ID)
	clist2 := srv.GetSubscribers(channel2ID)
	clist3 := srv.GetPublishers(channel1ID)
	clist4 := srv.GetPublishers("not-a-channel")
	assert.Equal(t, 2, len(clist1), "Expected 2 subscriber in channel 1")
	assert.Equal(t, 0, len(clist2), "Expected 0 subscribers in channel 2")
	assert.Equal(t, 2, len(clist3), "Expected 2 publisher in channel 1")
	assert.Equal(t, 0, len(clist4), "Expected 0 publishers in not-a-channel ")

	removeSuccessful := srv.RemoveConnection(c1)
	assert.True(t, removeSuccessful, "Connection c1 should have been found")
	removeSuccessful = srv.RemoveConnection(c2)
	assert.True(t, removeSuccessful, "Connection c2 should have been found")
	removeSuccessful = srv.RemoveConnection(c3)
	assert.True(t, removeSuccessful, "Connection c3 should have been found")
	removeSuccessful = srv.RemoveConnection(c4)
	assert.True(t, removeSuccessful, "Connection c4 should have been found")

	clist1 = srv.GetSubscribers(channel1ID)
	assert.Equal(t, 0, len(clist1), "Expected 0 remaining connections in channel 1")

	// removing twice should not fail
	srv.RemoveConnection(c1)
	srv.RemoveConnection(c4)
}

// Test client-server connection without TLS
func TestNewPubSub(t *testing.T) {
	const channel1 = "Chan1"
	const hostPort = "localhost:9678"

	logrus.Infof("Testing authentication on channel %s", channel1)
	cs, err := servicebus.StartServiceBus(hostPort)
	require.NoError(t, err)
	time.Sleep(time.Second)

	conn, err := lib.NewPublisher(hostPort, client1ID, channel1)
	require.NoError(t, err, "Error creating publisher: %s", err)
	require.NotNil(t, conn)

	conn, err = lib.NewSubscriber(hostPort, client1ID, channel1, func(channel string, msg []byte) {})
	require.NoError(t, err, "Error creating subscriber")

	_, err = lib.NewPublisher(hostPort, "", channel1)
	require.Error(t, err, "Expected error creating subscriber with invalid ID")

	cs.Stop()
}

// Test a TLS connection using a self generated certificates
// This also serves as an example on how to setup a server and client using a CA and client certificate
func TestTLSConnection(t *testing.T) {
	hostname := "10.3.3.30"
	// hostname := "localhost"
	hostPort := hostname + ":9678"
	message := "success!"

	const channel1 = "CH1"
	router := mux.NewRouter()
	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, message)
	}))

	caCertPEM, caKeyPEM := lib.CreateWoSTCA()
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)

	serverCertPEM, serverKeyPEM, err := lib.CreateGatewayCert(caCertPEM, caKeyPEM, hostname)
	clientCertPEM, clientKeyPEM, err := lib.CreateClientCert(caCertPEM, caKeyPEM, hostname)
	clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	require.NoErrorf(t, err, "Creating certificates failed:")

	serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	require.NoErrorf(t, err, "Creating server cert failed:")

	// The server TLS Config contains the client's CA  for Client certificate validation
	serverTLSConf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
	}
	server := &http.Server{
		Addr:      hostPort,
		Handler:   router,
		TLSConfig: serverTLSConf,
	}
	go server.ListenAndServeTLS("", "")
	time.Sleep(time.Second)
	defer server.Close()

	//-----
	// communicate with the server using an http.Client with the CA to verify the server
	// clientCertPool := x509.NewCertPool()
	// clientCertPool.AppendCertsFromPEM(clientCertPEM)
	clientTLSConf := &tls.Config{
		RootCAs:      caCertPool, // the server certificate must be signed by this CA
		Certificates: []tls.Certificate{clientCert},
	}
	transport := &http2.Transport{
		TLSClientConfig: clientTLSConf,
	}
	http := http.Client{Transport: transport}
	resp, err := http.Get("https://" + hostPort)
	require.NoError(t, err, "Failed reading from server")

	// verify the response
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoErrorf(t, err, "Failed reading response")
	assert.Equal(t, message, string(respBodyBytes))
}

// Test publish and subscribe client over TLS
func TestTLSPubSubChannel(t *testing.T) {
	const channel1 = "Chan1"
	const pubMsg1 = "Message 1"
	var subMsg1 = ""
	const hostPort = "localhost:9678"
	const certFolder = "../../test/"
	mutex1 := sync.Mutex{}

	logrus.Infof("Testing channel %s", channel1)
	// create new certificates in the test folder
	os.Remove(certFolder + servicebus.CaCertFile)
	os.Remove(certFolder + servicebus.CaKeyFile)
	os.Remove(certFolder + servicebus.ServerCertFile)
	os.Remove(certFolder + servicebus.ServerKeyFile)
	os.Remove(certFolder + servicebus.ClientCertFile)
	os.Remove(certFolder + servicebus.ClientKeyFile)

	// This re-generates the certificates
	cs, err := servicebus.StartTLSServiceBus(hostPort, certFolder)
	require.NoError(t, err)

	time.Sleep(time.Second)
	clientCertPEM, _ := ioutil.ReadFile(certFolder + servicebus.ClientCertFile)
	clientKeyPEM, _ := ioutil.ReadFile(certFolder + servicebus.ClientKeyFile)
	caCertPEM, _ := ioutil.ReadFile(certFolder + servicebus.CaCertFile)

	// clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)

	// send published channel messages to subscribers
	publisher, err := lib.NewTLSPublisher(hostPort, client1ID, channel1,
		clientCertPEM, clientKeyPEM, caCertPEM)
	require.NoError(t, err)

	subscriber, err := lib.NewTLSSubscriber(hostPort, client1ID, channel1,
		clientCertPEM, clientKeyPEM, caCertPEM,
		func(channel string, msg []byte) {
			logrus.Infof("TestChannel: Received published message on channel %s", channel)
			mutex1.Lock()
			subMsg1 = string(msg)
			mutex1.Unlock()
		})
	require.NoError(t, err)

	lib.SendMessage(publisher, []byte(pubMsg1))
	time.Sleep(1 * time.Second)
	mutex1.Lock()
	assert.Equal(t, pubMsg1, subMsg1)
	mutex1.Unlock()

	time.Sleep(time.Second * 1)

	// publisher.Close()
	subscriber.Close()
	// time.Sleep(time.Second)
	cs.Stop()
	cs.Stop()
}

// Test that after closing a channel no message is received
func TestCloseSubscriberChannel(t *testing.T) {
	const channel1 = "Chan1"
	const pubMsg1 = "Message 1"
	hostname := "localhost"
	hostPort := hostname + ":9678"
	var msgCount = 0
	const certFolder = "../../test/"

	// setup
	cs, err := servicebus.StartTLSServiceBus(hostPort, certFolder)
	require.NoError(t, err)
	time.Sleep(time.Second * 1)
	clientCertPEM, _ := ioutil.ReadFile(certFolder + servicebus.ClientCertFile)
	clientKeyPEM, _ := ioutil.ReadFile(certFolder + servicebus.ClientKeyFile)
	caCertPEM, _ := ioutil.ReadFile(certFolder + servicebus.CaCertFile)

	c1, err := lib.NewTLSSubscriber(hostPort, client1ID, channel1ID,
		clientCertPEM, clientKeyPEM, caCertPEM, func(channel string, msg []byte) {
			msgCount = msgCount + 1
			logrus.Infof("Received a message. This should show only once. Msgcount=%d", msgCount)
		})
	// _ = c1

	p1, err := lib.NewTLSPublisher(hostPort, client1ID, channel1ID,
		clientCertPEM, clientKeyPEM, caCertPEM)

	lib.SendMessage(p1, []byte(pubMsg1))
	c1.Close()
	lib.SendMessage(p1, []byte(pubMsg1))
	time.Sleep(1000 * time.Millisecond)
	assert.Equalf(t, 1, msgCount, "Expected only 1 message")
	cs.Stop()
}

// test sending messages to multiple subscribers
func TestLoad(t *testing.T) {
	const hostPort = "localhost:9678"
	var err error
	var pCon *websocket.Conn
	var t3 time.Time
	var t4 time.Time
	var rxCount int32 = 0
	var txCount int32 = 0
	const certFolder = "../../test/"
	mutex1 := sync.Mutex{}

	cs, err := servicebus.StartTLSServiceBus(hostPort, certFolder)
	require.NoError(t, err)
	time.Sleep(time.Second * 1)
	clientCertPEM, _ := ioutil.ReadFile(certFolder + servicebus.ClientCertFile)
	clientKeyPEM, _ := ioutil.ReadFile(certFolder + servicebus.ClientKeyFile)
	caCertPEM, _ := ioutil.ReadFile(certFolder + servicebus.CaCertFile)

	t0 := time.Now()
	// test creating 100 publishers and subscribers
	var sCount int
	for sCount = 0; sCount < 100; sCount++ {
		_, err := lib.NewTLSSubscriber(hostPort, client1ID, channel1ID,
			clientCertPEM, clientKeyPEM, caCertPEM, func(channel string, msg []byte) {
				mutex1.Lock()
				defer mutex1.Unlock()
				// msg is nil if connection closes
				if msg != nil {
					atomic.AddInt32(&rxCount, 1)
					t4 = time.Now() // latest received time
				}
				// logrus.Infof("Received message on receiver %d", sCount)
			})
		assert.NoErrorf(t, err, "Unexpected error creating subscriber %d", sCount)
	}

	t1 := time.Now()
	var pCount int
	for pCount = 0; pCount < 100; pCount++ {
		pCon, err = lib.NewTLSPublisher(hostPort, client1ID, channel1ID,
			clientCertPEM, clientKeyPEM, caCertPEM)
		assert.NoErrorf(t, err, "Unexpected error creating publisher %d", pCount)
	}
	t2 := time.Now()

	// time.Sleep(1 * time.Millisecond)
	for i := 0; i < 10; i++ {
		lib.SendMessage(pCon, []byte("Hello world"))
		txCount++
	}
	t3 = time.Now()

	// take time to receive them all
	time.Sleep(time.Second * 5)

	mutex1.Lock()
	assert.Equal(t, int(txCount)*sCount, int(rxCount), "not all subscribers received a message")
	chan1 := cs.GetChannel(channel1ID)
	assert.Equal(t, txCount, atomic.LoadInt32(&chan1.MessageCount), "Server received messages mismatch")
	mutex1.Unlock()

	cs.Stop()
	// time.Sleep(time.Millisecond * 1)
	mutex1.Lock()
	logrus.Printf("Time to create %d TLS subscribers: %d msec", sCount, t1.Sub(t0)/time.Millisecond)
	logrus.Printf("Time to create %d TLS publishers: %d msec", pCount, t2.Sub(t1)/time.Millisecond)
	logrus.Printf("Time to send %d TLS messages %d msec", txCount, t3.Sub(t2)/time.Millisecond)
	logrus.Printf("Time to receive %d TLS messages by subscribers: %d msec", rxCount, t4.Sub(t2)/time.Millisecond)
	mutex1.Unlock()
}
