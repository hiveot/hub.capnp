package mosquittomgr_test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/core/mosquittomgr"
	"github.com/wostzone/hub/pkg/auth"
	"github.com/wostzone/hub/pkg/unpwstore"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
	"github.com/wostzone/wostlib-go/pkg/hubnet"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/vocab"
)

var hubConfig *hubconfig.HubConfig
var homeFolder string
var configFolder string

// NOTE: GENERATE MOSQAUTH.SO BEFORE RUNNING THESE TESTS
// eg, cd mosquitto-pb/mosqauth/main && make

// TestMain uses the project test folder as the home folder and generates test certificates

// these names must match the auth_opt_* filenames in mosquitto.conf.template
const aclFileName = "test.acl" // auth_opt_aclFile
const unpwFileName = "test.passwd"

var aclFilePath string
var unpwFilePath string

func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	hubConfig, _ = hubconfig.LoadHubConfig(homeFolder, mosquittomgr.PluginID)
	configFolder = hubConfig.ConfigFolder
	hubconfig.SetLogging(hubConfig.Loglevel, "")

	ip := hubnet.GetOutboundIP(hubConfig.MqttAddress).String()
	names := []string{string(ip), hubConfig.MqttAddress}
	// for testing the certs must exist
	certsFolder := path.Join(homeFolder, "certs")
	certsetup.CreateCertificateBundle(names, certsFolder)

	// clean acls and passwd file
	aclFilePath = path.Join(configFolder, aclFileName)
	unpwFilePath = path.Join(configFolder, unpwFileName)
	fp, _ := os.Create(aclFilePath)
	fp.Close()
	fp, _ = os.Create(unpwFilePath)
	fp.Close()
	result := m.Run()
	os.Exit(result)
}

func TestStartStopMosqManager(t *testing.T) {
	logrus.Infof("---TestStartStopMosqManager---")

	// FIXME: configuration password and acl store location
	svc := mosquittomgr.NewMosquittoManager()
	err := hubconfig.LoadPluginConfig(configFolder, mosquittomgr.PluginID, &svc.Config, nil)
	assert.NoError(t, err)

	err = svc.Start(hubConfig)
	assert.NoError(t, err)

	// main.AuthPluginInit(nil, nil, 0)

	svc.Stop()
}

func TestPluginConnect(t *testing.T) {
	logrus.Infof("---TestPluginConnect---")
	const pluginID = "mosquitto-pb-test"
	// const plugin2ID = "mosquitto-pb-test2"
	const thing1ID = "urn:test:thing1"

	svc := mosquittomgr.NewMosquittoManager()
	err := hubconfig.LoadPluginConfig(configFolder, mosquittomgr.PluginID, &svc.Config, nil)
	assert.NoError(t, err)

	err = svc.Start(hubConfig)
	assert.NoError(t, err)

	// a plugin must be able to connect using a client certificate
	client := hubclient.NewMqttHubPluginClient(pluginID, hubConfig)
	err = client.Connect()
	require.NoError(t, err)

	// publish should succeed
	td := td.CreateTD(thing1ID, vocab.DeviceTypeService)
	err = client.PublishTD(thing1ID, td)
	assert.NoError(t, err)
	time.Sleep(time.Second)
	client.Close()
	svc.Stop()
}

func TestPasswdWithMosqManager(t *testing.T) {
	logrus.Infof("--- TestPasswdWithMosqManager ---")
	var err error
	username := "user2"
	password1 := "user2"

	pfs := unpwstore.NewPasswordFileStore(unpwFilePath, "TestPasswdWithMosqManager")
	err = pfs.Open()
	assert.NoError(t, err)
	// for logging timestamps
	time.Sleep(time.Millisecond * 100)
	pwhash, err := auth.CreatePasswordHash(password1, auth.PWHASH_ARGON2id, 0)
	assert.NoError(t, err)
	logrus.Infof("--- TestPasswdWithMosqManager: Setting password for %s", username)
	pfs.SetPasswordHash(username, pwhash)
	assert.NoError(t, err)

	// for logging timestamps - dont mix setting passwd with starting mosq mgr
	time.Sleep(time.Millisecond * 1000)

	logrus.Infof("--- TestPasswdWithMosqManager: Creating MosquittoManager")
	svc := mosquittomgr.NewMosquittoManager()
	err = hubconfig.LoadPluginConfig(configFolder, mosquittomgr.PluginID, &svc.Config, nil)
	assert.NoError(t, err)

	err = svc.Start(hubConfig)
	assert.NoError(t, err)
	// for logging timestamps
	time.Sleep(time.Millisecond * 100)

	// a consumer must be able to subscribe using a valid password
	hostPort := fmt.Sprintf("%s:%d", hubConfig.MqttAddress, hubConfig.MqttUnpwPortWS)
	caCertFile := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	client := hubclient.NewMqttHubClient(hostPort, caCertFile, username, password1)

	err = client.Connect()
	require.NoError(t, err)
	client.Close()

	time.Sleep(time.Second)
	// close twice should not fail
	client.Close()
	svc.Stop()
}

func TestBadPasswd(t *testing.T) {
	logrus.Infof("---TestBadPasswd---")
	username := "user1"
	password1 := "badpass"

	svc := mosquittomgr.NewMosquittoManager()
	err := hubconfig.LoadPluginConfig(configFolder, mosquittomgr.PluginID, &svc.Config, nil)
	assert.NoError(t, err)
	err = svc.Start(hubConfig)
	assert.NoError(t, err)

	// a consumer must not be able to subscribe using a invalid password
	hostPort := fmt.Sprintf("%s:%d", hubConfig.MqttAddress, hubConfig.MqttUnpwPortWS)
	caCertFile := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	client := hubclient.NewMqttHubClient(hostPort, caCertFile, username, password1)
	err = client.Connect()
	require.Error(t, err)
	client.Close() // should not panic

	svc.Stop()
}

func TestTemplateNotFound(t *testing.T) {
	logrus.Infof("---TestTemplateNotFound---")

	svc := mosquittomgr.NewMosquittoManager()
	err := hubconfig.LoadPluginConfig(configFolder, mosquittomgr.PluginID, &svc.Config, nil)
	svc.Config.MosquittoTemplate = "./notatemplatefile"
	assert.NoError(t, err)
	err = svc.Start(hubConfig)
	assert.Error(t, err)

	svc.Stop()
}

func TestBadConfigTemplate(t *testing.T) {
	logrus.Infof("---TestBadConfigTemplate---")

	svc := mosquittomgr.NewMosquittoManager()
	err := hubconfig.LoadPluginConfig(configFolder, mosquittomgr.PluginID, &svc.Config, nil)
	assert.NoError(t, err)
	svc.Config.MosquittoTemplate = "mosquitto.conf.bad-template"
	err = svc.Start(hubConfig)
	assert.Error(t, err)
	time.Sleep(time.Second)
	svc.Stop()
}
