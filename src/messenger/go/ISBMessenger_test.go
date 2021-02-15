package messenger_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messenger "github.com/wostzone/gateway/src/messenger/go"
)

const isbCertFolder = "../../test"
const serverHostPort = "localhost:8883"

// Test create and close the internal service bus channel
func TestISBConnection(t *testing.T) {
	serverHostPort := "localhost"
	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerConnect(t, client, serverHostPort)
}

func TestISBNoConnect(t *testing.T) {
	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerNoConnect(t, client)
}

func TestISBNoClientID(t *testing.T) {
	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerNoClientID(t, client, serverHostPort)
}

func TestISBPubSubNoTLS(t *testing.T) {
	serverHostPort := "localhost:9678" // default

	isbServer, err := StartSimbu(serverHostPort)
	_ = isbServer
	require.NoError(t, err, "Failed starting the ISB server")
	time.Sleep(10 * time.Millisecond)

	client := messenger.NewISBMessenger("")
	TMessengerPubSub(t, client, serverHostPort)
}
func TestISBPubSubWithTLS(t *testing.T) {
	serverHostPort := "localhost:9678" // default

	isbServer, err := StartSimbuTLS(serverHostPort, isbCertFolder)
	_ = isbServer
	require.NoError(t, err, "Failed starting the ISB server")
	time.Sleep(10 * time.Millisecond)

	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerPubSub(t, client, serverHostPort)
}

func TestISBMultipleSubscriptions(t *testing.T) {
	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerMultipleSubscriptions(t, client, serverHostPort)
}

func TestISBBadUnsubscribe(t *testing.T) {
	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerBadUnsubscribe(t, client, serverHostPort)
}

func TestISBPubNoConnect(t *testing.T) {
	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerPubNoConnect(t, client)
}

func TestISBSubBeforeConnect(t *testing.T) {
	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerSubBeforeConnect(t, client, serverHostPort)
}
