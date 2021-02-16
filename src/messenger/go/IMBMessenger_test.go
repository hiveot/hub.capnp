package messenger_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	messenger "github.com/wostzone/gateway/src/messenger/go"
	"github.com/wostzone/gateway/src/msgbus"
)

const isbCertFolder = "../../../test"
const serverHostPort = msgbus.DefaultMsgBusHost

// Test create and close the internal service bus channel
func TestISBConnection(t *testing.T) {
	srv, _ := msgbus.Start(serverHostPort)
	client := messenger.NewISBMessenger("")
	TMessengerConnect(t, client, serverHostPort)
	srv.Stop()
}

func TestISBNoConnect(t *testing.T) {
	srv, _ := msgbus.Start(serverHostPort)
	client := messenger.NewISBMessenger("")
	TMessengerNoConnect(t, client)
	srv.Stop()
}

func TestISBPubSubNoTLS(t *testing.T) {
	srv, _ := msgbus.Start(serverHostPort)
	client := messenger.NewISBMessenger("")
	TMessengerPubSubNoTLS(t, client, serverHostPort)
	srv.Stop()
}
func TestISBPubSubWithTLS(t *testing.T) {
	srv, err := msgbus.StartTLS(serverHostPort, isbCertFolder)
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	client := messenger.NewISBMessenger(isbCertFolder)
	TMessengerPubSubWithTLS(t, client, serverHostPort)
	srv.Stop()
}

func TestISBMultipleSubscriptions(t *testing.T) {
	srv, _ := msgbus.Start(serverHostPort)
	client := messenger.NewISBMessenger("")
	TMessengerMultipleSubscriptions(t, client, serverHostPort)
	srv.Stop()
}

func TestISBBadUnsubscribe(t *testing.T) {
	srv, _ := msgbus.Start(serverHostPort)
	client := messenger.NewISBMessenger("")
	TMessengerBadUnsubscribe(t, client, serverHostPort)
	srv.Stop()
}

func TestISBPubNoConnect(t *testing.T) {
	srv, _ := msgbus.Start(serverHostPort)
	client := messenger.NewISBMessenger("")
	TMessengerPubNoConnect(t, client)
	srv.Stop()
}

func TestISBSubBeforeConnect(t *testing.T) {
	srv, _ := msgbus.Start(serverHostPort)
	client := messenger.NewISBMessenger("")
	TMessengerSubBeforeConnect(t, client, serverHostPort)
	srv.Stop()
}
