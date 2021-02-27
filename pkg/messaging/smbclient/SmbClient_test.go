package smbclient_test

import (
	"testing"
	"time"

	"github.com/wostzone/gateway/pkg/messaging/smbclient"
	testhelper "github.com/wostzone/gateway/pkg/messaging/test"
)

const smbCertFolder = "../../../test/certs"

const serverHostPort = "localhost:9999"

// !!! THE TESTS REQUIRE THAT A SMBSERVER IS RUNNING !!!
// from plugins/smbserver run:
//    go run plugins/smbserver/main.go --home ../../test

// Test create and close the simple message bus channel
func TestSmbClientConnection(t *testing.T) {
	// srv, _ := smberver.StartSmbus(serverHostPort)
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerConnect(t, client)
	// srv.Stop()
}

func TestSmbusNoConnect(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbclient.NewSmbClient(smbCertFolder, "localhost:0")
	testhelper.TMessengerNoConnect(t, client)
	// srv.Stop()
}

// func TestSmbusPubSubNoTLS(t *testing.T) {
// 	// srv, _ := smbserver.Start(serverHostPort)
// 	client := smbclient.NewSmbusMessenger("", serverHostPort)
// 	testhelper.TMessengerPubSub(t, client)
// 	// srv.Stop()
// }
func TestSmbusPubSubWithTLS(t *testing.T) {
	// srv, err := smbserver.StartTLS(serverHostPort, certFolder)
	// assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerPubSub(t, client)
	// srv.Stop()
}

func TestSmbusMultipleSubscriptions(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerMultipleSubscriptions(t, client)
	// srv.Stop()
}

func TestSmbusBadUnsubscribe(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerBadUnsubscribe(t, client)
	// srv.Stop()
}

func TestSmbusPubNoConnect(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerPubNoConnect(t, client)
	// srv.Stop()
}

func TestSmbusSubBeforeConnect(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerSubBeforeConnect(t, client)
	// srv.Stop()
}
