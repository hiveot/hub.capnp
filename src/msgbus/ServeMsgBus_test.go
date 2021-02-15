package msgbus_test

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
	"github.com/wostzone/gateway/src/certs"
	"github.com/wostzone/gateway/src/msgbus"
	"golang.org/x/net/http2"
)

const channel1ID = "channel1"
const channel2ID = "channel2"
const channel3ID = "channel3"
const defaultBufferSize = 1

// const hostPort = "localhost:9678"

const client1ID = "plugin1"

// Test create, store and remove channels by the server
func TestCreateChannel(t *testing.T) {
	logrus.Info("Testing create channels")
	srv := msgbus.NewServeMsgBus()
	c1 := &websocket.Conn{}
	c2 := &websocket.Conn{}
	c3 := &websocket.Conn{}
	// channel1 := srv.NewChannel(channel1ID, defaultBufferSize)
	srv.AddSubscriber(channel1ID, c1)
	srv.AddSubscriber(channel1ID, c2)
	srv.AddSubscriber(channel2ID, c3)

	clist1 := srv.GetSubscribers(channel1ID)
	clist2 := srv.GetSubscribers(channel2ID)
	clist3 := srv.GetSubscribers(channel3ID)
	assert.Equal(t, 2, len(clist1), "Expected 2 subscribers in channel 1")
	assert.Equal(t, 1, len(clist2), "Expected 1 subscriber in channel 2")
	assert.Equal(t, 0, len(clist3), "Expected 0 subscriber in channel 3")

	srv.RemoveConnection(c1)
	cList := srv.GetSubscribers(channel1ID)
	assert.Equal(t, 1, len(cList), "Connection c1 should have been removed")
	srv.RemoveConnection(c2)
	cList = srv.GetSubscribers(channel1ID)
	assert.Equal(t, 0, len(cList), "Connection c2 should have been removed")
	cList = srv.GetSubscribers(channel2ID)
	assert.Equal(t, 1, len(cList), "Connection c3 should not have been removed")

	// removing twice should not fail
	srv.RemoveConnection(c2)
}

// Test client-server connection without TLS
func TestConnectNoTLS(t *testing.T) {
	const channel1 = "Chan1"
	const hostPort = "localhost:9678"

	// logrus.Infof("Testing authentication on channel %s", channel1)
	cs, err := msgbus.Start(hostPort)
	require.NoError(t, err)
	time.Sleep(time.Second)

	conn, err := msgbus.NewWebsocketConnection(hostPort, client1ID, nil)
	require.NoError(t, err, "Error creating publisher: %s", err)
	require.NotNil(t, conn)

	cs.Stop()
}

// test connecting by a non websocket client
func TestConnectNoWSClient(t *testing.T) {
	const channel1 = "Chan1"
	const hostPort = "localhost:9678"
	const client1ID = "cid1"
	const certFolder = "../../test"
	var err error

	srv, err := msgbus.Start(hostPort)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// conn, err := msgbus.NewWebsocketConnection(hostPort, client1ID, nil)
	// require.NoError(t, err, "Error creating publisher: %s", err)
	// require.NotNil(t, conn)

	// url := fmt.Sprintf(msgbus.MsgbusURL, hostPort)
	url := "http://localhost:9678/wost"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set(msgbus.ClientHeader, client1ID)
	resp, err := client.Do(req)

	require.NotNil(t, resp)
	assert.True(t, resp.StatusCode >= 400)
	time.Sleep(time.Second)

	srv.Stop()
}

func TestConnectInvalidClientID(t *testing.T) {
	const channel1 = "Chan1"
	const hostPort = "localhost:9678"

	cs, err := msgbus.Start(hostPort)
	require.NoError(t, err)
	time.Sleep(time.Second)
	_, err = msgbus.NewWebsocketConnection(hostPort, "", nil)
	require.Error(t, err, "Expected error creating subscriber with invalid ID")
	cs.Stop()

}

func TestStartTwice(t *testing.T) {
	const channel1 = "Chan1"
	const hostPort = "localhost:9678"

	cs1, err := msgbus.Start(hostPort)
	require.NoError(t, err)
	cs2, err := msgbus.Start(hostPort)
	time.Sleep(time.Second)

	cs1.Stop()
	cs2.Stop()

}

func TestPubSub(t *testing.T) {
	const channel1 = "Chan1"
	const channel2 = "Chan2"
	const hostPort = "localhost:9678"
	const msg1 = "Hello world 1"
	const msg2 = "Hello world 2"
	var rx string

	mb, err := msgbus.Start(hostPort)
	require.NoError(t, err)
	time.Sleep(time.Second)

	rawHandler1 := func(command string, channel string, msg []byte) {
		logrus.Infof("TestPubSub: received command '%s' for channel '%s'", command, channel)
		if command == msgbus.MsgBusCommandReceive {
			rx = string(msg)
		}
	}
	c, _ := msgbus.NewWebsocketConnection(hostPort, client1ID, rawHandler1)
	require.NotNil(t, c)
	// must receive a message to the subscribed channel
	err = msgbus.Subscribe(c, channel1)
	require.NoError(t, err)

	// publish to channel with subscribers
	err = msgbus.Publish(c, channel1, []byte(msg1))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, msg1, rx)

	// publish to  channel without subscribers
	err = msgbus.Publish(c, channel2, []byte(msg1))
	require.NoError(t, err)

	// publish to unsubscribed channel
	err = msgbus.Unsubscribe(c, channel1)
	msgbus.Publish(c, channel1, []byte(msg2))
	time.Sleep(time.Millisecond)
	assert.NotEqual(t, msg2, rx)

	time.Sleep(time.Second)

	mb.Stop()
}

// Test a TLS connection using a self generated certificates
// This also serves as an example on how to setup a server and client using a CA and client certificate
func TestTLSCerts(t *testing.T) {
	// hostname := "10.3.3.30"
	hostname := "localhost"
	hostPort := hostname + ":9678"
	message := "success!"

	// const channel1 = "CH1"
	router := mux.NewRouter()
	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, message)
	}))

	caCertPEM, caKeyPEM := certs.CreateWoSTCA()
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)

	serverCertPEM, serverKeyPEM, err := certs.CreateGatewayCert(caCertPEM, caKeyPEM, hostname)
	clientCertPEM, clientKeyPEM, err := certs.CreateClientCert(caCertPEM, caKeyPEM, hostname)
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
func TestTLSPubSub(t *testing.T) {
	const channel1 = "Chan1"
	const pubMsg1 = "Message 1"
	var subMsg1 = ""
	const hostPort = "localhost:9678"
	const certFolder = "../../test/"
	mutex1 := sync.Mutex{}

	logrus.Infof("Testing channel %s", channel1)
	// create new certificates in the test folder
	os.Remove(certFolder + msgbus.CaCertFile)
	os.Remove(certFolder + msgbus.CaKeyFile)
	os.Remove(certFolder + msgbus.ServerCertFile)
	os.Remove(certFolder + msgbus.ServerKeyFile)
	os.Remove(certFolder + msgbus.ClientCertFile)
	os.Remove(certFolder + msgbus.ClientKeyFile)

	// This re-generates the certificates
	cs, err := msgbus.StartTLS(hostPort, certFolder)
	require.NoError(t, err)

	time.Sleep(time.Second)
	clientCertPEM, _ := ioutil.ReadFile(certFolder + msgbus.ClientCertFile)
	clientKeyPEM, _ := ioutil.ReadFile(certFolder + msgbus.ClientKeyFile)
	caCertPEM, _ := ioutil.ReadFile(certFolder + msgbus.CaCertFile)

	// clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)

	handler1 := func(command string, channel string, msg []byte) {
		logrus.Infof("TestTLSPubSubChannel: handler1 received command '%s' on channel '%s'", command, channel)
		mutex1.Lock()
		subMsg1 = string(msg)
		mutex1.Unlock()
	}

	// send published channel messages to subscribers
	conn1, err := msgbus.NewTLSWebsocketConnection(hostPort, client1ID, handler1,
		clientCertPEM, clientKeyPEM, caCertPEM)
	require.NoError(t, err)
	require.NotNil(t, conn1)

	err = msgbus.Subscribe(conn1, channel1)
	require.NoError(t, err)

	logrus.Infof("TestTLSPubSubChannel: publishing message on channel '%s'", channel1)
	time.Sleep(10 * time.Millisecond)
	err = msgbus.Publish(conn1, channel1, []byte(pubMsg1))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	mutex1.Lock()
	assert.Equal(t, pubMsg1, subMsg1)
	mutex1.Unlock()

	time.Sleep(time.Second * 1)

	// publisher.Close()
	conn1.Close()
	// time.Sleep(time.Second)
	cs.Stop()
	cs.Stop()
}

// Test that after closing a channel no message is received
func TestCloseSubscriberChannel(t *testing.T) {
	const channel1 = "Chan1"
	// const channel2 = "Chan2"
	const pubMsg1 = "Message 1"
	hostname := "localhost"
	hostPort := hostname + ":9678"
	var msgCount = 0
	const certFolder = "../../test/"

	// setup
	cs, err := msgbus.StartTLS(hostPort, certFolder)
	require.NoError(t, err)
	time.Sleep(time.Second * 1)
	clientCertPEM, _ := ioutil.ReadFile(certFolder + msgbus.ClientCertFile)
	clientKeyPEM, _ := ioutil.ReadFile(certFolder + msgbus.ClientKeyFile)
	caCertPEM, _ := ioutil.ReadFile(certFolder + msgbus.CaCertFile)

	handler := func(command string, channel string, msg []byte) {
		if command == msgbus.MsgBusCommandReceive {
			msgCount = msgCount + 1
			logrus.Infof("Received a message. This should show only once. Msgcount=%d", msgCount)
		}
	}

	c1, err := msgbus.NewTLSWebsocketConnection(hostPort, client1ID, handler,
		clientCertPEM, clientKeyPEM, caCertPEM)
	c2, err := msgbus.NewTLSWebsocketConnection(hostPort, client1ID, handler,
		clientCertPEM, clientKeyPEM, caCertPEM)
	msgbus.Subscribe(c1, channel1)
	msgbus.Subscribe(c2, channel1)
	time.Sleep(10 * time.Millisecond)
	// first message is received twice
	msgbus.Publish(c1, channel1, []byte(pubMsg1))
	time.Sleep(1 * time.Second)

	// second message is received only once
	c2.Close()
	// time.Sleep(10 * time.Millisecond)
	msgbus.Publish(c1, channel1, []byte(pubMsg1))
	// msgbus.Publish(c1, channel2, []byte(pubMsg1))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 3, msgCount)

	cs.Stop()
}

// test sending messages to multiple subscribers
func TestLoad(t *testing.T) {
	const hostPort = "localhost:9678"
	var err error
	var conn *websocket.Conn
	var t3 time.Time
	var t4 time.Time
	var rxCount int32 = 0
	var txCount int32 = 0
	const certFolder = "../../test/"
	mutex1 := sync.Mutex{}

	cs, err := msgbus.StartTLS(hostPort, certFolder)
	require.NoError(t, err)
	time.Sleep(time.Second * 1)
	clientCertPEM, _ := ioutil.ReadFile(certFolder + msgbus.ClientCertFile)
	clientKeyPEM, _ := ioutil.ReadFile(certFolder + msgbus.ClientKeyFile)
	caCertPEM, _ := ioutil.ReadFile(certFolder + msgbus.CaCertFile)

	t1 := time.Now()
	// test creating 100 connections
	handler := func(command string, channel string, msg []byte) {
		mutex1.Lock()
		defer mutex1.Unlock()
		// msg is nil if connection closes
		if command == msgbus.MsgBusCommandReceive && msg != nil {
			atomic.AddInt32(&rxCount, 1)
			t4 = time.Now() // latest received time
		}
		// logrus.Infof("Received message on receiver %d", sCount)
	}
	var cCount int
	for cCount = 0; cCount < 100; cCount++ {
		conn, err = msgbus.NewTLSWebsocketConnection(hostPort, client1ID, handler,
			clientCertPEM, clientKeyPEM, caCertPEM)
		assert.NoErrorf(t, err, "Unexpected error creating subscriber %d", cCount)
		msgbus.Subscribe(conn, channel1ID)
	}

	t2 := time.Now()

	// time.Sleep(1 * time.Millisecond)
	for i := 0; i < 1000; i++ {
		msgbus.Publish(conn, channel1ID, []byte("Hello world"))
		txCount++
	}
	t3 = time.Now()

	// take time to receive them all
	time.Sleep(time.Second * 5)

	mutex1.Lock()
	assert.Equal(t, int(txCount)*cCount, int(rxCount), "not all subscribers received a message")
	chan1 := cs.GetChannel(channel1ID)
	require.NotNil(t, chan1)
	assert.Equal(t, txCount, atomic.LoadInt32(&chan1.MessageCount), "Server received messages mismatch")
	mutex1.Unlock()

	cs.Stop()
	// time.Sleep(time.Millisecond * 1)
	mutex1.Lock()
	logrus.Printf("Time to create %d TLS connections: %d msec", cCount, t2.Sub(t1)/time.Millisecond)
	logrus.Printf("Time to send %d TLS messages %d msec", txCount, t3.Sub(t2)/time.Millisecond)
	logrus.Printf("Time to receive %d TLS messages by subscribers: %d msec", rxCount, t4.Sub(t2)/time.Millisecond)
	mutex1.Unlock()
}

func TestServeHome(t *testing.T) {
	hostname := "localhost"
	hostPort := hostname + ":9678"

	// setup
	mb, err := msgbus.Start(hostPort)
	time.Sleep(time.Millisecond)
	res, err := http.Get("http://" + hostPort + "/")
	require.NoError(t, err)
	logrus.Infof("TestServeHome: result: %s", res.Status)

	res, err = http.Get("http://" + hostPort + "/non-page")
	require.NoError(t, err)

	res, err = http.Post("http://"+hostPort+"/", "text", nil)
	require.NoError(t, err)

	// time.Sleep(100 * time.Second)
	mb.Stop()
}
