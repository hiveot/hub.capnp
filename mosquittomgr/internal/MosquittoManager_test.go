package internal_test

import (
	"fmt"
	"github.com/wostzone/wost-go/pkg/logging"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/authn/pkg/jwtissuer"
	"github.com/wostzone/hub/mosquittomgr/internal"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/mqttclient"
	"github.com/wostzone/wost-go/pkg/testenv"
)

var hubConfig *config.HubConfig
var configFolder string // mosquitto config files
var tempFolder string
var testCerts testenv.TestCerts
var mosqTemplateFile string // location of the mosquitto template file

// NOTE: GENERATE MOSQAUTH.SO BEFORE RUNNING THESE TESTS
// eg, cd mosquitto-pb/mosqauth/main && make

// TestMain uses the project test folder as the home folder and generates test certificates

// these names must match the auth_opt_* filenames in mosquitto.conf.template
const aclFileName = "test.acl" // auth_opt_aclFile
const unpwFileName = "test.passwd"

var aclFilePath string
var unpwFilePath string

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	testCerts = testenv.CreateCertBundle()
	tempFolder = path.Join(os.TempDir(), "wost-mosquittomgr-test")
	testenv.SaveCerts(&testCerts, tempFolder)

	// load the plugin config with client certs from the temp folder
	cwd, _ := os.Getwd()
	configFolder = path.Join(cwd, "..", "test", "config")
	configFile := path.Join(configFolder, config.DefaultHubConfigName)
	mosqTemplateFile = path.Join(configFolder, internal.DefaultTemplateFile)

	hubConfig = config.CreateHubConfig(tempFolder)
	// mosqauth.so plugin lives in bin folder
	hubConfig.BinFolder = path.Join(cwd, "..", "..", "dist", "bin")
	hubConfig.Load(configFile, internal.PluginID)

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
	svc.Config.MosquittoTemplate = mosqTemplateFile

	err := svc.Start(hubConfig)
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
	svc.Config.MosquittoTemplate = mosqTemplateFile
	err := svc.Start(hubConfig)
	assert.NoError(t, err)

	// a plugin must be able to connect using a client certificate
	client := mqttclient.NewMqttClient(pluginID, hubConfig.CaCert, 0)
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortCert)
	err = client.ConnectWithClientCert(hostPort, hubConfig.PluginCert)
	if assert.NoError(t, err) {
		//tdoc := thing.CreateTD(thing1ID, "test thing", vocab.DeviceTypeService)
		//eThing := mqttbinding.CreateExposedThing(thing1ID, tdoc, client)
		// publish should succeed
		//err = eThing.Expose()
		assert.NoError(t, err)
		time.Sleep(time.Second)

		client.Disconnect()
	}
	// capture mosquitto printfs?
	time.Sleep(time.Second * 3)
	svc.Stop()
}

func TestJWTWithMosqManager(t *testing.T) {
	logrus.Infof("--- TestJWTWithMosqManager ---")
	var err error
	username := "user2"

	svc := internal.NewMosquittoManager()
	svc.Config.MosquittoTemplate = mosqTemplateFile
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
	client.Disconnect()

	time.Sleep(time.Second)
	// close twice should not fail
	client.Disconnect()
	svc.Stop()
}

func TestBadPasswd(t *testing.T) {
	logrus.Infof("---TestBadPasswd---")
	username := "user1"
	password1 := "badpass"

	svc := internal.NewMosquittoManager()
	svc.Config.MosquittoTemplate = mosqTemplateFile
	err := svc.Start(hubConfig)
	assert.NoError(t, err)

	// a consumer must not be able to subscribe using a invalid password
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortUnpw)
	// caCertFile := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	client := mqttclient.NewMqttClient("clientID", hubConfig.CaCert, 0)
	err = client.ConnectWithAccessToken(hostPort, username, password1)
	require.Error(t, err)
	client.Disconnect() // should not panic

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
	svc.Config.MosquittoTemplate = path.Join(configFolder, "mosquitto.conf.bad-template")
	err := svc.Start(hubConfig)
	assert.Error(t, err)
	time.Sleep(time.Second)
	svc.Stop()
}
