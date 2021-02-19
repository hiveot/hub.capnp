package smbus_test

import (
	"testing"
	"time"

	"github.com/wostzone/gateway/pkg/messaging/smbus"
	testhelper "github.com/wostzone/gateway/pkg/messaging/test"
)

// const smbusCertFolder = "../../../test/certs"
const smbusCertFolder = "/home/henk/bin/wost/certs"
const serverHostPort = smbus.DefaultSmbusHost

// THIS REQUIRES THE LWWS PLUGIN TO RUN

// Test create and close the simple message bus channel
func TestSmbusConnection(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbus.NewSmbusMessenger(smbusCertFolder, serverHostPort)
	testhelper.TMessengerConnect(t, client)
	// srv.Stop()
}

func TestSmbusNoConnect(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbus.NewSmbusMessenger(smbusCertFolder, "localhost:0")
	testhelper.TMessengerNoConnect(t, client)
	// srv.Stop()
}

// func TestSmbusPubSubNoTLS(t *testing.T) {
// 	// srv, _ := smbserver.Start(serverHostPort)
// 	client := smbus.NewSmbusMessenger("", serverHostPort)
// 	testhelper.TMessengerPubSub(t, client)
// 	// srv.Stop()
// }
func TestSmbusPubSubWithTLS(t *testing.T) {
	// srv, err := smbserver.StartTLS(serverHostPort, certFolder)
	// assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	client := smbus.NewSmbusMessenger(smbusCertFolder, serverHostPort)
	testhelper.TMessengerPubSub(t, client)
	// srv.Stop()
}

func TestSmbusMultipleSubscriptions(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbus.NewSmbusMessenger(smbusCertFolder, serverHostPort)
	testhelper.TMessengerMultipleSubscriptions(t, client)
	// srv.Stop()
}

func TestSmbusBadUnsubscribe(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbus.NewSmbusMessenger(smbusCertFolder, serverHostPort)
	testhelper.TMessengerBadUnsubscribe(t, client)
	// srv.Stop()
}

func TestSmbusPubNoConnect(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbus.NewSmbusMessenger(smbusCertFolder, serverHostPort)
	testhelper.TMessengerPubNoConnect(t, client)
	// srv.Stop()
}

func TestSmbusSubBeforeConnect(t *testing.T) {
	// srv, _ := smbserver.Start(serverHostPort)
	client := smbus.NewSmbusMessenger(smbusCertFolder, serverHostPort)
	testhelper.TMessengerSubBeforeConnect(t, client)
	// srv.Stop()
}
