package mosquittomgr_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/core/mosquittomgr"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/vocab"
)

var hubConfig *hubconfig.HubConfig
var homeFolder string
var configFolder string

// NOTE: GENERATE MOSQAUTH.SO BEFORE RUNNING THESE TESTS
// eg, cd mosquitto-pb/mosqauth/main && make

// easy cleanup for existing  certificate
// func removeCerts(folder string) {
// 	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(folder, "*.pem")).Output()
// }

// TestMain uses the project test folder as the home folder and generates test certificates
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	hubConfig, _ = hubconfig.LoadHubConfig(homeFolder)
	configFolder = hubConfig.ConfigFolder
	hubconfig.SetLogging(hubConfig.Loglevel, "", "")

	// TODO: should cert be based on name or address, or both?
	ip := hubconfig.GetOutboundIP(hubConfig.MqttAddress).String()
	names := []string{string(ip), hubConfig.MqttAddress}
	// for testing the certs must exist
	certsFolder := path.Join(homeFolder, "certs")
	// removeCerts(certsFolder)
	certsetup.CreateCertificateBundle(names, certsFolder)

	result := m.Run()
	os.Exit(result)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("---TestStartStop---")
	const pluginID = "mosquitto-pb-test"

	svc := mosquittomgr.NewMosquittoManager()
	err := hubconfig.LoadPluginConfig(configFolder, mosquittomgr.PluginID, &svc.Config)
	assert.NoError(t, err)

	err = svc.Start(hubConfig)
	assert.NoError(t, err)

	// main.AuthPluginInit(nil, nil, 0)

	svc.Stop()
}

func TestPluginConnect(t *testing.T) {
	logrus.Infof("---TestPluginConnect---")
	const pluginID = "mosquitto-pb-test"
	const plugin2ID = "mosquitto-pb-test2"
	const thing1ID = "urn:test:thing1"

	svc := mosquittomgr.NewMosquittoManager()
	err := hubconfig.LoadPluginConfig(configFolder, mosquittomgr.PluginID, &svc.Config)
	assert.NoError(t, err)

	err = svc.Start(hubConfig)
	assert.NoError(t, err)

	// a plugin must be able to connect using a client certificate
	client := hubclient.NewMqttHubPluginClient(pluginID, hubConfig)
	err = client.Start()
	require.NoError(t, err)

	// publish should succeed
	td := td.CreateTD(thing1ID, vocab.DeviceTypeService)
	err = client.PublishTD(thing1ID, td)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	svc.Stop()
}
