package listener_test

import (
	"crypto/tls"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/test"
)

// CA, server and plugin test certificate
var certs test.TestCerts

// TestMain runs a http server
// Used for all test cases in this package
func TestMain(m *testing.M) {
	certs = test.CreateCertBundle()
	res := m.Run()
	os.Exit(res)
}

func TestConnectWriteRead(t *testing.T) {
	readBuf := make([]byte, 100)
	var message = []byte("hello world")
	var n int
	address := "127.0.0.1:9999"
	rwmux := sync.RWMutex{}

	// create the server listener
	tlsLis, err := listener.CreateListener(address, false, certs.ServerCert, certs.CaCert)
	require.NoError(t, err)
	go func() {
		srvConn, err := tlsLis.Accept()
		require.NoError(t, err)
		err = srvConn.(*tls.Conn).Handshake()
		assert.NoError(t, err)
		err = srvConn.SetReadDeadline(time.Now().Add(time.Second))
		assert.NoError(t, err)
		time.Sleep(time.Millisecond)

		scs := srvConn.(*tls.Conn).ConnectionState()
		if assert.Equal(t, 1, len(scs.PeerCertificates)) {
			// the test cert has a CN of "Plugin"
			pcert := scs.PeerCertificates[0]
			clientID := pcert.Subject.CommonName
			assert.Equal(t, "Plugin", clientID)
		}
		rwmux.Lock()
		n, _ = srvConn.Read(readBuf)
		readBuf = readBuf[0:n]
		remoteClient := srvConn.RemoteAddr().String()
		logrus.Infof("read %d bytes from '%s'", n, remoteClient)
		rwmux.Unlock()
	}()
	time.Sleep(time.Millisecond)
	// create the TLS client and connect
	fullURL := address
	conn, err := hubclient.CreateClientConnection(fullURL, certs.PluginCert, certs.CaCert)
	require.NoError(t, err)

	tlsConn, valid := conn.(*tls.Conn)
	if valid {
		state := tlsConn.ConnectionState()
		t.Logf("SSL ServerName: %s", state.ServerName)
		t.Logf("SSL Handshake: %v", state.HandshakeComplete)
		t.Logf("SSL Mutual: %s", state.NegotiatedProtocol)
	}
	m, err := conn.Write(message)
	assert.NoError(t, err)
	assert.Equal(t, 11, m)

	time.Sleep(time.Millisecond * 10) // give read some time
	rwmux.RLock()
	assert.Equal(t, 11, n)
	assert.Equal(t, message, readBuf)
	rwmux.RUnlock()

	err = tlsConn.Close()
	assert.NoError(t, err)
}
