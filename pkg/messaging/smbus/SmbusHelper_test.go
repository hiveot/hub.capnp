package smbus_test

import (
	"io/ioutil"
	"net/http"
	"path"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/pkg/certs"
	"github.com/wostzone/gateway/pkg/messaging/smbus"
)

// helper function to test WsConnect methods
func startLittleServer(t *testing.T, hostPort string) *http.Server {
	var upgrader websocket.Upgrader = websocket.Upgrader{}

	router := mux.NewRouter()
	router.HandleFunc("/wost", func(resp http.ResponseWriter, req *http.Request) {
		pubOrSub := mux.Vars(req)["stage"]
		logrus.Infof("TestNewPubSub: calling socket upgrade to websocket: %s", pubOrSub)
		upgrader.Upgrade(resp, req, nil)
	})

	httpServer := &http.Server{
		Addr:    hostPort,
		Handler: router,
	}
	go func() {
		// cs.updateMutex.Unlock()
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(t, err, "ListenAndServe failed: %s", err)
		}
	}()
	return httpServer
}

func TestConnect(t *testing.T) {
	const channel1 = "Chan1"
	const hostPort = "localhost:9666"
	const client1ID = "cid1"
	const certFolder = "../../test"
	var err error
	httpServer := startLittleServer(t, hostPort)
	time.Sleep(100 * time.Millisecond)

	conn, err := smbus.NewWebsocketConnection(hostPort, client1ID, nil)
	require.NoError(t, err)

	// subConn, err := NewSubscriber(hostPort, client1ID, channel1, func(channel string, msg []byte) {
	// })
	// assert.NoError(t, err)

	err = smbus.Publish(conn, channel1, []byte("Hello world"))
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	httpServer.Close()
	conn.Close()
	time.Sleep(100 * time.Millisecond)
}

func TestNewPubSubErrors(t *testing.T) {
	const channel1 = "Chan1"
	const hostPort = "localhost:9666"
	const client1ID = "cid1"
	const certFolder = "../../test"

	serverCertPath := path.Join(certFolder, certs.ServerCertFile)
	clientCertPath := path.Join(certFolder, certs.ClientCertFile)
	clientKeyPath := path.Join(certFolder, certs.ClientKeyFile)

	serverCertPEM, _ := ioutil.ReadFile(serverCertPath)
	clientCertPEM, _ := ioutil.ReadFile(clientCertPath)
	clientKeyPEM, _ := ioutil.ReadFile(clientKeyPath)

	logrus.Infof("Testing authentication on channel %s", channel1)

	// These should fail as no server is listening
	_, err := smbus.NewWebsocketConnection(hostPort, client1ID, nil)
	require.Error(t, err)

	_, err = smbus.NewTLSWebsocketConnection(hostPort, client1ID, nil, clientCertPEM, clientKeyPEM, serverCertPEM)
	require.Error(t, err)

	// c := &websocket.Conn{}
	// err = SendMessage(c, []byte("no c error"))
	// require.Error(t, err, "Error creating subscriber")

	// cs.Stop()
}

func TestBadPublish(t *testing.T) {
	const channel1 = "Chan1"
	const hostPort = "localhost:1111"
	const client1ID = "cid1"
	const msg1 = "tada"

	c, _ := smbus.NewWebsocketConnection(hostPort, client1ID, nil)
	require.Nil(t, c)

	// publish to channel with subscribers
	err := smbus.Publish(c, channel1, []byte(msg1))
	require.Error(t, err)

}
