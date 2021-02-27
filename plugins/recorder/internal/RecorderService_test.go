package internal_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/pkg/messaging"
	"github.com/wostzone/gateway/pkg/messaging/smbserver"
	"github.com/wostzone/gateway/plugins/recorder/internal"
)

var homeFolder string

const pluginID = "recorder-test"

// Use the project app folder during testing
func init() {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../../test")
}

func TestStartStopRecorder(t *testing.T) {

	recConfig := &internal.RecorderConfig{} // use defaults
	gwConfig, err := lib.SetupConfig(homeFolder, pluginID, recConfig)
	assert.NoError(t, err)
	server, err := smbserver.StartSmbServer(gwConfig)
	assert.NoError(t, err)

	svc := internal.NewRecorderService()
	err = svc.Start(gwConfig, recConfig)
	assert.NoError(t, err)
	svc.Stop()
	server.Stop()
}

func TestRecordMessage(t *testing.T) {

	recConfig := &internal.RecorderConfig{} // use defaults
	gwConfig, err := lib.SetupConfig(homeFolder, pluginID, recConfig)
	assert.NoError(t, err)
	server, err := smbserver.StartSmbServer(gwConfig)
	assert.NoError(t, err)
	svc := internal.NewRecorderService()
	err = svc.Start(gwConfig, recConfig)
	client, err := messaging.StartGatewayMessenger("test1", gwConfig)
	assert.NoError(t, err)

	client.Publish(messaging.EventsChannelID, []byte("Hello world"))
	time.Sleep(1 * time.Second)
	client.Disconnect()

	assert.NoError(t, err)
	svc.Stop()
	server.Stop()
}
