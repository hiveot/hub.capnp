package authservice_test

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/authn/pkg/authservice"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"
	"os"
	"path"
	"testing"
	"time"
)

var serverAddress = "127.0.0.1"
var serverPort uint = 9881
var testCerts testenv.TestCerts
var passwordFile string

//var serverCertFolder string
//var clientHostPort string

var storeFolder = ""

const user1 = "user1"
const pass1 = "secret1"

// helper to start the authn service for testing
// containing a password for user1
func startAuthService() (*authservice.AuthService, error) {
	config := authservice.AuthServiceConfig{
		Address:                  serverAddress,
		Port:                     serverPort,
		PasswordFile:             passwordFile,
		ConfigStoreFolder:        storeFolder,
		ConfigStoreEnabled:       true,
		AccessTokenValiditySec:   10,
		RefreshTokenValidityDays: 1,
	}
	srv := authservice.NewJwtAuthService(config, nil, testCerts.ServerCert, testCerts.CaCert)
	err := srv.Start()
	if err == nil {
		err = srv.SetPassword(user1, pass1)
	}
	return srv, err
}

// TestMain runs a http server
// Used for all test cases in this package
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of AuthService_test ------")
	//clientHostPort = fmt.Sprintf("%s:%d", serverAddress, serverPort)

	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "..", "..", "test")
	//serverCertFolder = path.Join(homeFolder, "certs")
	storeFolder = path.Join(homeFolder, "configStore")
	passwordFile = path.Join(homeFolder, "config", "test.passwd")
	// empty file
	fp, _ := os.Create(passwordFile)
	_ = fp.Close()

	testCerts = testenv.CreateCertBundle()
	res := m.Run()

	time.Sleep(time.Second)
	os.Exit(res)
}

// Create and verify a JWT token
func TestStartStop(t *testing.T) {
	//user1 := "user1"
	srv, err := startAuthService()
	assert.NoError(t, err)

	// start twice should not break things
	err = srv.Start()
	assert.Error(t, err)

	srv.Stop()
	// stopping twice should not break things
	srv.Stop()
}

// Create and verify a JWT token
func TestStartTwice(t *testing.T) {
	//user1 := "user1"
	srv, err := startAuthService()
	assert.NoError(t, err)

	// run duplicate should fail
	srv2, err := startAuthService()
	assert.Error(t, err)
	srv2.Stop()

	srv.Stop()
}

func TestLogin(t *testing.T) {
	pass2 := "secret2"
	srv, err := startAuthService()
	assert.NoError(t, err)
	//
	hostPort := fmt.Sprintf("%s:%d", serverAddress, serverPort)
	authClient := tlsclient.NewTLSClient(hostPort, testCerts.CaCert)

	accessToken, err := authClient.ConnectWithJWTLogin(user1, pass1, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	accessToken, err = authClient.ConnectWithJWTLogin(user1, pass2, "")
	assert.Error(t, err)
	assert.Empty(t, accessToken)

	srv.Stop()
}

func TestRefresh(t *testing.T) {
	//user1 := "user1"
}

func TestRefreshInvalid(t *testing.T) {
	//user1 := "user1"
}

func TestGetConfig(t *testing.T) {
	srv, err := startAuthService()
	assert.NoError(t, err)
	//
	hostPort := fmt.Sprintf("%s:%d", serverAddress, serverPort)
	authClient := tlsclient.NewTLSClient(hostPort, testCerts.CaCert)

	accessToken, err := authClient.ConnectWithJWTLogin(user1, pass1, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	myConfig := "my configuration object"
	_, err = authClient.Put(tlsclient.DefaultJWTConfigPath+"/app1", myConfig)
	assert.NoError(t, err)

	data, err := authClient.Get(tlsclient.DefaultJWTConfigPath + "/app1")
	assert.NoError(t, err)
	assert.Equal(t, myConfig, string(data))

	data, err = authClient.Get(tlsclient.DefaultJWTConfigPath + "/app2")
	assert.NoError(t, err)
	assert.Empty(t, data)
	srv.Stop()

}

func TestUpdateConfigBadMethod(t *testing.T) {
	srv, err := startAuthService()
	assert.NoError(t, err)

	myConfig := "my configuration object"
	hostPort := fmt.Sprintf("%s:%d", serverAddress, serverPort)
	authClient := tlsclient.NewTLSClient(hostPort, testCerts.CaCert)
	accessToken, err := authClient.ConnectWithJWTLogin(user1, pass1, "")
	_ = accessToken
	assert.NoError(t, err)

	_, err = authClient.Post(tlsclient.DefaultJWTConfigPath+"/app1", myConfig)
	assert.Error(t, err)
	srv.Stop()

}
