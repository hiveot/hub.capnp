package smbclient_test

import (
	"testing"

	"github.com/wostzone/gateway/pkg/config"
	"github.com/wostzone/gateway/pkg/messaging/smbclient"
	testhelper "github.com/wostzone/gateway/pkg/messaging/testhelper"
	"github.com/wostzone/gateway/pkg/smbserver"
)

const smbCertFolder = "../../../test/certs"

const serverHostPort = "localhost:9999"

var srv *smbserver.ServeSmbus

func init() {
	config.SetLogging("info", "", "")
}

func setup() {
	// these tests need a smbserver running
	// config.SetLogging("info", "", "")
	// cwd, _ := os.Getwd()
	// homeFolder = path.Join(cwd, "../../test")
	srv, _ = smbserver.StartTLS(serverHostPort, smbCertFolder)
}
func teardown() {
	srv.Stop()
}

// Test create and close the simple message bus channel
func TestSmbClientConnection(t *testing.T) {
	setup()
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerConnect(t, client)
	teardown()
}

func TestSmbusNoConnect(t *testing.T) {
	setup()
	client := smbclient.NewSmbClient(smbCertFolder, "localhost:0")
	testhelper.TMessengerNoConnect(t, client)
	teardown()
}

// func TestSmbusPubSubNoTLS(t *testing.T) {
// 	// srv, _ := smbserver.Start(serverHostPort)
// 	client := smbclient.NewSmbusMessenger("", serverHostPort)
// 	testhelper.TMessengerPubSub(t, client)
// 	// srv.Stop()
// }
func TestSmbusPubSubWithTLS(t *testing.T) {
	setup()
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerPubSub(t, client)
	teardown()
}

func TestSmbusMultipleSubscriptions(t *testing.T) {
	setup()
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerMultipleSubscriptions(t, client)
	teardown()
}

func TestSmbusBadUnsubscribe(t *testing.T) {
	setup()
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerBadUnsubscribe(t, client)
	teardown()
}

func TestSmbusPubNoConnect(t *testing.T) {
	setup()
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerPubNoConnect(t, client)
	teardown()
}

func TestSmbusSubBeforeConnect(t *testing.T) {
	setup()
	client := smbclient.NewSmbClient(smbCertFolder, serverHostPort)
	testhelper.TMessengerSubBeforeConnect(t, client)
	teardown()
}
