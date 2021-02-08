package servicebus_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
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
	client "github.com/wostzone/gateway/client/go"
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

	conn, err := client.NewPublisher(hostPort, client1ID, channel1)
	require.NoError(t, err, "Error creating publisher: %s", err)
	require.NotNil(t, conn)

	conn, err = client.NewSubscriber(hostPort, client1ID, channel1, func(msg []byte) {})
	require.NoError(t, err, "Error creating subscriber")

	_, err = client.NewPublisher(hostPort, "", channel1)
	require.Error(t, err, "Expected error creating subscriber with invalid ID")

	cs.Stop()
}

func TestTLSCertificateGeneration(t *testing.T) {
	// host := "localhost:9678"
	// hostname := "127.0.0.1:9678"
	// hostname := "10.3.3.30"
	hostname := "localhost"

	// test creating ca and server certificates
	caCertPEM, caKeyPEM := servicebus.CreateWoSTCA()
	require.NotNilf(t, caCertPEM, "Failed creating CA certificate")

	caCert, err := tls.X509KeyPair(caCertPEM, caKeyPEM)
	_ = caCert
	require.NoErrorf(t, err, "Failed parsing CA certificate")

	serverCertPEM, serverKeyPEM, err := servicebus.CreateGatewayCert(caCertPEM, caKeyPEM, hostname)
	require.NoErrorf(t, err, "Failed creating server certificate")
	// serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	require.NoErrorf(t, err, "Failed creating server certificate")
	require.NotNilf(t, serverCertPEM, "Failed creating server certificate")
	require.NotNilf(t, serverKeyPEM, "Failed creating server private key")

	// verify the certificate
	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM(caCertPEM)
	require.True(t, ok, "Failed parsing CA certificate")

	serverBlock, _ := pem.Decode(serverCertPEM)
	require.NotNil(t, serverBlock, "Failed decoding server certificate PEM")

	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	require.NoError(t, err, "ParseCertificate for server failed")

	opts := x509.VerifyOptions{
		Roots:   certpool,
		DNSName: hostname,
		// DNSName:       "127.0.0.1",
		Intermediates: x509.NewCertPool(),
	}
	_, err = serverCert.Verify(opts)
	require.NoError(t, err, "Verify for server certificate failed")
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

	caCertPEM, caKeyPEM := servicebus.CreateWoSTCA()
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)

	serverCertPEM, serverKeyPEM, err := servicebus.CreateGatewayCert(caCertPEM, caKeyPEM, hostname)
	clientCertPEM, clientKeyPEM, err := servicebus.CreateClientCert(caCertPEM, caKeyPEM, hostname)
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
	publisher, err := client.NewTLSPublisher(hostPort, client1ID, channel1,
		clientCertPEM, clientKeyPEM, caCertPEM)
	require.NoError(t, err)

	subscriber, err := client.NewTLSSubscriber(hostPort, client1ID, channel1,
		clientCertPEM, clientKeyPEM, caCertPEM,
		func(msg []byte) {
			logrus.Info("TestChannel: Received published message")
			mutex1.Lock()
			subMsg1 = string(msg)
			mutex1.Unlock()
		})
	require.NoError(t, err)

	client.SendMessage(publisher, []byte(pubMsg1))
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

// test sending messages to multiple subscribers
func TestLoad(t *testing.T) {
	const hostPort = "localhost:9678"
	var err error
	var pCon *websocket.Conn
	var t3 time.Time
	var t4 time.Time
	var rxCount int32 = 0
	var txCount int32 = 0
	var lastclient *websocket.Conn
	const certFolder = "../../test/"
	mutex1 := sync.Mutex{}

	cs, err := servicebus.StartTLSServiceBus(hostPort, certFolder)
	require.NoError(t, err)
	time.Sleep(time.Second * 1)
	clientCertPEM, _ := ioutil.ReadFile(certFolder + servicebus.ClientCertFile)
	clientKeyPEM, _ := ioutil.ReadFile(certFolder + servicebus.ClientKeyFile)
	caCertPEM, _ := ioutil.ReadFile(certFolder + servicebus.CaCertFile)

	t0 := time.Now()
	// test creating 1000 publishers and subscribers
	var sCount int = 0
	for sCount = 0; sCount < 201; sCount++ {
		c, err := client.NewTLSSubscriber(hostPort, client1ID, channel1ID,
			clientCertPEM, clientKeyPEM, caCertPEM, func(msg []byte) {
				mutex1.Lock()
				atomic.AddInt32(&rxCount, 1)
				t4 = time.Now() // latest received time
				mutex1.Unlock()
				// logrus.Infof("Received message on receiver %d", sCount)
			})
		assert.NoErrorf(t, err, "Unexpected error creating subscriber %d", sCount)
		lastclient = c
	}

	t1 := time.Now()
	var pCount = 0
	// var pCon *websocket.Conn
	for pCount = 0; pCount < 200; pCount++ {
		pCon, err = client.NewTLSPublisher(hostPort, client1ID, channel1ID,
			clientCertPEM, clientKeyPEM, caCertPEM)
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

	mutex1.Lock()
	assert.Equal(t, int(txCount)*sCount, int(rxCount), "not all subscribers received a message")
	chan1 := cs.GetChannel(channel1ID)
	assert.Equal(t, txCount, atomic.LoadInt32(&chan1.MessageCount), "Server received messages mismatch")
	mutex1.Unlock()

	cs.Stop()
	// time.Sleep(time.Millisecond * 1)
	mutex1.Lock()
	logrus.Printf("Time to create %d subscribers: %d msec", sCount, t1.Sub(t0)/time.Millisecond)
	logrus.Printf("Time to create %d publishers: %d msec", pCount, t2.Sub(t1)/time.Millisecond)
	logrus.Printf("Time to send %d messages %d usec", txCount, t3.Sub(t2)/time.Microsecond)
	logrus.Printf("Time to receive %d messages by subscribers: %d msec", rxCount, t4.Sub(t2)/time.Millisecond)
	mutex1.Unlock()
}
