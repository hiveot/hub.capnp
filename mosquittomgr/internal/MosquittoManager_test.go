package internal_test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/wostzone/hub/authz/pkg/aclstore"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/logging"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/hub/authn/pkg/jwtissuer"
	"github.com/wostzone/hub/mosquittomgr/internal"
	"github.com/wostzone/wost-go/pkg/mqttclient"
	"github.com/wostzone/wost-go/pkg/testenv"
)

//var hubConfig *config.HubConfig

var tempFolder string
var templateFolder string // mosquitto.conf template file
var mosqauthPlugin string
var testCerts testenv.TestCerts

//var mosqTemplateFile string // location of the mosquitto template file

var aclFilePath string

func createMMConfig() internal.MMConfig {
	mmConfig := internal.MMConfig{}
	mmConfig.AclFile = aclFilePath
	mmConfig.Address = testenv.ServerAddress
	mmConfig.CaCertFile = path.Join(tempFolder, config.DefaultCaCertFile)
	mmConfig.ClientID = internal.PluginID
	mmConfig.LogFolder = tempFolder
	mmConfig.MosquittoConfFile = path.Join(tempFolder, internal.DefaultConfFile)
	mmConfig.MosquittoTemplateFile = path.Join(templateFolder, internal.DefaultTemplateFile)
	mmConfig.MosqAuthPlugin = mosqauthPlugin
	mmConfig.ServerCertFile = path.Join(tempFolder, config.DefaultServerCertFile)
	mmConfig.ServerKeyFile = path.Join(tempFolder, config.DefaultServerKeyFile)
	mmConfig.MqttPortUnpw = testenv.MqttPortUnpw
	mmConfig.MqttPortCert = testenv.MqttPortCert
	mmConfig.MqttPortWS = testenv.MqttPortWS

	return mmConfig
}

// NOTE: GENERATE MOSQAUTH.SO BEFORE RUNNING THESE TESTS (use make all in the project root)
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	tempFolder = path.Join(os.TempDir(), "wost-mosquittomgr-test")
	testCerts = testenv.CreateCertBundle()
	testenv.SaveCerts(&testCerts, tempFolder)

	// load the plugin config with client certs from the temp folder
	cwd, _ := os.Getwd()
	templateFolder = path.Join(cwd, "..", "test", "config")
	mosqauthPlugin = path.Join(cwd, "..", "dist", "bin", "mosqauth.so")

	// clean acls
	aclFilePath = path.Join(tempFolder, aclstore.DefaultAclFile)
	fp, _ := os.Create(aclFilePath)
	_ = fp.Close()
	result := m.Run()
	os.Exit(result)
}

func TestStartStopMosqManager(t *testing.T) {
	logrus.Infof("---TestStartStopMosqManager---")
	mmConfig := createMMConfig()
	svc := internal.NewMosquittoManager(mmConfig)

	err := svc.Start()
	assert.NoError(t, err)

	// main.AuthPluginInit(nil, nil, 0)

	svc.Stop()
}

func TestPluginConnect(t *testing.T) {
	logrus.Infof("---TestPluginConnect---")
	const pluginID = "mosquittomgr-test"
	const thing1ID = "urn:test:thing1"

	mmConfig := createMMConfig()
	svc := internal.NewMosquittoManager(mmConfig)
	err := svc.Start()
	//mcmd, err := testenv.StartMosquitto(&testCerts, tempFolder)
	assert.NoError(t, err)

	// a plugin must be able to connect using a client certificate
	client := mqttclient.NewMqttClient(pluginID, testCerts.CaCert, 0)
	hostPort := fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortCert)
	err = client.ConnectWithClientCert(hostPort, testCerts.PluginCert)
	if assert.NoError(t, err) {
		assert.NoError(t, err)
		time.Sleep(time.Second)
		client.Disconnect()
	}
	// capture mosquitto printfs?
	time.Sleep(time.Second * 3)
	svc.Stop()
	//testenv.StopMosquitto(mcmd, tempFolder)
}

func TestJWTWithMosqManager(t *testing.T) {
	logrus.Infof("--- TestJWTWithMosqManager ---")
	var err error
	username := "user2"

	mmConfig := createMMConfig()
	svc := internal.NewMosquittoManager(mmConfig)

	err = svc.Start()
	assert.NoError(t, err)
	// for logging timestamps
	time.Sleep(time.Millisecond * 100)

	hostPort := fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortUnpw)
	client := mqttclient.NewMqttClient("clientID", testCerts.CaCert, 0)

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

	mmConfig := createMMConfig()
	svc := internal.NewMosquittoManager(mmConfig)
	err := svc.Start()
	assert.NoError(t, err)

	// a consumer must not be able to subscribe using a invalid password
	hostPort := fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortUnpw)
	// caCertFile := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	client := mqttclient.NewMqttClient("clientID", testCerts.CaCert, 0)
	err = client.ConnectWithAccessToken(hostPort, username, password1)
	require.Error(t, err)
	client.Disconnect() // should not panic

	svc.Stop()
}

func TestTemplateNotFound(t *testing.T) {
	logrus.Infof("---TestTemplateNotFound---")

	mmConfig := createMMConfig()
	mmConfig.MosquittoTemplateFile = "./notatemplatefile"
	svc := internal.NewMosquittoManager(mmConfig)
	err := svc.Start()
	assert.Error(t, err)

	svc.Stop()
}

func TestBadConfigTemplate(t *testing.T) {
	logrus.Infof("---TestBadConfigTemplate---")

	mmConfig := createMMConfig()
	mmConfig.MosquittoTemplateFile = path.Join(templateFolder, "mosquitto.conf.bad-template")
	svc := internal.NewMosquittoManager(mmConfig)
	err := svc.Start()
	assert.Error(t, err)
	time.Sleep(time.Second)
	svc.Stop()
}
