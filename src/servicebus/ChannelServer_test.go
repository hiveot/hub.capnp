package servicebus_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	client "github.com/wostzone/gateway/client/go"
	"github.com/wostzone/gateway/src/servicebus"
)

const channel1ID = "channel1"
const channel2ID = "channel2"
const defaultBufferSize = 1

const host = "localhost:9678"

const client1ID = "plugin1"
const authToken1 = "token1"

var authTokens = map[string]string{
	client1ID: authToken1,
}

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

func TestInvalidAuthentication(t *testing.T) {
	const channel1 = "Chan1"
	const invalidAuthToken = "invalid-token"
	const certFolder = "../test"

	logrus.Infof("Testing authentication on channel %s", channel1)
	cs := servicebus.StartServiceBus(host, "", authTokens)
	time.Sleep(time.Second)

	_, err1 := client.NewPublisher(host, client1ID, invalidAuthToken, channel1)
	_, err2 := client.NewSubscriber(host, client1ID, invalidAuthToken, channel1, func(msg []byte) {})
	_, err3 := client.NewPublisher(host, "", "", channel1)

	require.Error(t, err1, "Expected error creating publisher with invalid auth")
	require.Error(t, err2, "Expected error creating subscriber with invalid auth")
	require.Error(t, err3, "Expected error creating subscriber with invalid auth")

	cs.Stop()
}

func TestTLS(t *testing.T) {
	// 	const channel1 = "Chan1"
	// 	const pubMsg1 = "Message 1"
	// host := "localhost:9678"
	// host := "127.0.0.1:9678"
	// hostname := "127.0.0.1"
	hostname := "localhost"
	hostPort := hostname + ":9678"

	// srv, clientTLSConf := servicebus.StartServiceBus(host, authTokens)
	// _ = srv
	// get our ca and server certificate
	caCertPEM, caKeyPEM := servicebus.CreateWoSTCA()
	serverCertPEM, serverKeyPEM, err := servicebus.CreateGatewayCert(caCertPEM, caKeyPEM, hostname)

	require.NoErrorf(t, err, "Failed creating server certificate")
	require.NotNilf(t, serverCertPEM, "Failed creating server certificate")
	require.NotNilf(t, serverKeyPEM, "Failed creating server private key")

	serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	serverTLSConf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	router := mux.NewRouter()
	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "success!")
	}))

	server := &http.Server{
		Addr:      hostPort,
		Handler:   router,
		TLSConfig: serverTLSConf,
		// TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	go server.ListenAndServeTLS("", "")
	time.Sleep(time.Second)
	defer server.Close()

	//-----
	// communicate with the server using an http.Client configured to trust our CA
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caCertPEM)
	clientTLSConf := &tls.Config{
		RootCAs: certpool,
	}
	transport := &http.Transport{
		TLSClientConfig: clientTLSConf,
		// EnableHTTP2: true,
	}
	// clientTLSConf.InsecureSkipVerify = true
	http := http.Client{
		Transport: transport,
	}
	resp, err := http.Get("https://" + hostPort)
	require.NoError(t, err, "Failed reading from server")

	// verify the response
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	body := strings.TrimSpace(string(respBodyBytes[:]))
	if body == "success!" {
		fmt.Println(body)
	} else {
		panic("not successful!")
	}
}

// Test publish and subscribe client
func TestPubSubChannel(t *testing.T) {
	const channel1 = "Chan1"
	const pubMsg1 = "Message 1"
	var subMsg1 = ""

	logrus.Infof("Testing channel %s", channel1)
	cs := servicebus.StartServiceBus(host, "", authTokens)
	time.Sleep(time.Second)

	// send published channel messages to subscribers
	publisher, err := client.NewPublisher(host, client1ID, authToken1, channel1)
	require.NoError(t, err)

	subscriber, err := client.NewSubscriber(host, client1ID, authToken1, channel1,
		func(msg []byte) {
			logrus.Info("TestChannel: Received published message")
			subMsg1 = string(msg)
		})
	require.NoError(t, err)

	client.SendMessage(publisher, []byte(pubMsg1))
	time.Sleep(1 * time.Second)
	assert.Equal(t, pubMsg1, subMsg1)

	time.Sleep(time.Second * 1)

	// publisher.Close()
	subscriber.Close()
	// time.Sleep(time.Second)
	cs.Stop()
	cs.Stop()
}

// test sending messages to multiple subscribers
func TestLoad(t *testing.T) {
	var err error
	var pCon *websocket.Conn
	var t3 time.Time
	var t4 time.Time
	var rxCount int32 = 0
	var txCount int = 0
	var lastclient *websocket.Conn

	cs := servicebus.StartServiceBus(host, "", authTokens)
	time.Sleep(time.Second * 1)
	t0 := time.Now()
	// test creating 1000 publishers and subscribers
	var sCount int = 0
	for sCount = 0; sCount < 200; sCount++ {
		c, err := client.NewSubscriber(host, client1ID, authToken1, channel1ID, func(msg []byte) {
			atomic.AddInt32(&rxCount, 1)
			t4 = time.Now() // latest received time
			// logrus.Infof("Received message on receiver %d", sCount)
		})
		assert.NoErrorf(t, err, "Unexpected error creating subscriber %d", sCount)
		lastclient = c
	}

	t1 := time.Now()
	var pCount = 0
	// var pCon *websocket.Conn
	for pCount = 0; pCount < 200; pCount++ {
		pCon, err = client.NewPublisher(host, client1ID, authToken1, channel1ID)
		assert.NoErrorf(t, err, "Unexpected error creating publisher %d", pCount)
	}
	t2 := time.Now()

	// pretend a subscriber connection dropped while sending
	lastclient.Close()
	sCount--

	for i := 0; i < 500; i++ {
		client.SendMessage(pCon, []byte("Hello world"))
		txCount++
	}
	t3 = time.Now()

	// take time to receive them all
	time.Sleep(time.Second * 3)

	assert.Equal(t, txCount*sCount, int(rxCount), "not all subscribers received a message")
	chan1 := cs.GetChannel(channel1ID)
	assert.Equal(t, txCount, chan1.MessageCount, "Server received messages mismatch")

	cs.Stop()
	// time.Sleep(time.Millisecond * 1)
	logrus.Printf("Time to create %d subscribers: %d msec", sCount, t1.Sub(t0)/time.Millisecond)
	logrus.Printf("Time to create %d publishers: %d msec", pCount, t2.Sub(t1)/time.Millisecond)
	logrus.Printf("Time to send %d messages %d usec", txCount, t3.Sub(t2)/time.Microsecond)
	logrus.Printf("Time to receive %d messages by subscribers: %d msec", rxCount, t4.Sub(t2)/time.Millisecond)
}
