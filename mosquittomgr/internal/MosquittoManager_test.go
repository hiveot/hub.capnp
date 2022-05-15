package internal_test

import (
	"fmt"
	"github.com/wostzone/hub/authn/pkg/jwtissuer"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"github.com/wostzone/hub/mosquittomgr/internal"
)

var hubConfig *config.HubConfig
var homeFolder string

var testCerts testenv.TestCerts

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
	homeFolder = path.Join(cwd, "../test")
	certsFolder := path.Join(homeFolder, config.DefaultCertsFolder)
	testCerts = testenv.CreateCertBundle()
	testenv.SaveCerts(&testCerts, certsFolder)

	// load the plugin config with client cert
	hubConfig = config.CreateDefaultHubConfig(homeFolder)
	_ = config.LoadHubConfig("", internal.PluginID, hubConfig)

	// clean acls and passwd file
	aclFilePath = path.Join(hubConfig.ConfigFolder, aclFileName)
	unpwFilePath = path.Join(hubConfig.ConfigFolder, unpwFileName)
	fp, _ := os.Create(aclFilePath)
	_ = fp.Close()
	fp, _ = os.Create(unpwFilePath)
	_ = fp.Close()
	result := m.Run()
	os.Exit(result)
}

func TestStartStopMosqManager(t *testing.T) {
	logrus.Infof("---TestStartStopMosqManager---")

	// FIXME: configuration password and acl store location
	svc := internal.NewMosquittoManager()
	configFile := path.Join(hubConfig.ConfigFolder, internal.PluginID+".yaml")
	err := config.LoadYamlConfig(configFile, &svc.Config, nil)
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

	svc := internal.NewMosquittoManager()
	err := svc.Start(hubConfig)
	assert.NoError(t, err)

	// a plugin must be able to connect using a client certificate
	client := mqttclient.NewMqttClient(pluginID, hubConfig.CaCert, 0)
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortCert)
	err = client.ConnectWithClientCert(hostPort, hubConfig.PluginCert)
	if assert.NoError(t, err) {

		tdoc := thing.CreateTD(thing1ID, "test thing", vocab.DeviceTypeService)
		eThing := mqttbinding.CreateExposedThing(thing1ID, tdoc, client)
		// publish should succeed
		err = eThing.Expose()
		assert.NoError(t, err)
		time.Sleep(time.Second)
		client.Close()
	}
	// capture mosquitto printfs?
	time.Sleep(time.Second * 3)
	svc.Stop()
}

func TestJWTWithMosqManager(t *testing.T) {
	logrus.Infof("--- TestJWTWithMosqManager ---")
	var err error
	username := "user2"

	logrus.Infof("--- TestPasswdWithMosqManager: Creating MosquittoManager")
	svc := internal.NewMosquittoManager()
	err = svc.Start(hubConfig)
	assert.NoError(t, err)
	// for logging timestamps
	time.Sleep(time.Millisecond * 100)

	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortUnpw)
	client := mqttclient.NewMqttClient("clientID", hubConfig.CaCert, 0)

	issuer := jwtissuer.NewJWTIssuer("test", testCerts.ServerKey, 10, 10, nil)
	accessToken, _, _ := issuer.CreateJWTTokens(username)
	err = client.ConnectWithAccessToken(hostPort, username, accessToken)
	assert.NoError(t, err)
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

	svc := internal.NewMosquittoManager()
	err := svc.Start(hubConfig)
	assert.NoError(t, err)

	// a consumer must not be able to subscribe using a invalid password
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortUnpw)
	// caCertFile := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	client := mqttclient.NewMqttClient("clientID", hubConfig.CaCert, 0)
	err = client.ConnectWithAccessToken(hostPort, username, password1)
	require.Error(t, err)
	client.Close() // should not panic

	svc.Stop()
}

func TestTemplateNotFound(t *testing.T) {
	logrus.Infof("---TestTemplateNotFound---")

	svc := internal.NewMosquittoManager()
	svc.Config.MosquittoTemplate = "./notatemplatefile"
	err := svc.Start(hubConfig)
	assert.Error(t, err)

	svc.Stop()
}

func TestBadConfigTemplate(t *testing.T) {
	logrus.Infof("---TestBadConfigTemplate---")

	svc := internal.NewMosquittoManager()
	svc.Config.MosquittoTemplate = "mosquitto.conf.bad-template"
	err := svc.Start(hubConfig)
	assert.Error(t, err)
	time.Sleep(time.Second)
	svc.Stop()
}
