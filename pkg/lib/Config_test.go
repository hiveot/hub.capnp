package lib_test

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/pkg/lib"
)

const logFile = "../../test/logs/setlogging.log"

type ConfigType1 struct {
	C1 string
	c2 string
}

func TestDefaultConfig(t *testing.T) {
	ci := lib.CreateGatewayConfig("")
	require.NotNil(t, ci)
	err := lib.ValidateConfig(ci)
	assert.Error(t, err, "Expected error as this test app does not reside in bin folder")
}

func TestLoadGatewayConfig(t *testing.T) {
	testFile := "../../test/config/gateway.yaml"

	ci := lib.CreateGatewayConfig("../../test")
	require.NotNil(t, ci)
	err := lib.LoadConfig(testFile, ci)
	assert.NoError(t, err)
	err = lib.ValidateConfig(ci)
	assert.NoError(t, err)
}

func TestLoadGatewayConfigNotFound(t *testing.T) {
	testFile := "../../test/config/gateway-notfound.yaml"

	ci := lib.CreateGatewayConfig("../../test")
	require.NotNil(t, ci)
	err := lib.LoadConfig(testFile, ci)
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadGatewayConfigYamlError(t *testing.T) {
	testFile := "../../test/config/gateway-bad.yaml"

	ci := lib.CreateGatewayConfig("../../test")
	require.NotNil(t, ci)
	err := lib.LoadConfig(testFile, ci)
	// Error should contain info on bad file
	errTxt := err.Error()
	assert.Equal(t, "yaml: line 11", errTxt[:13], "Expected line 11 to be bad")
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadGatewayConfigBadFolders(t *testing.T) {

	ci := lib.CreateGatewayConfig("../../test")
	err := lib.ValidateConfig(ci)
	assert.NoError(t, err, "Default config should be okay")
	ci2 := *ci
	ci2.ConfigFolder = "/doesntexist"
	err = lib.ValidateConfig(&ci2)
	assert.Error(t, err)
	ci2 = *ci
	ci2.Logging.LogsFolder = "/doesntexist"
	err = lib.ValidateConfig(&ci2)
	assert.Error(t, err)
	ci2 = *ci
	ci2.Messenger.CertsFolder = "/doesntexist"
	err = lib.ValidateConfig(&ci2)
	assert.Error(t, err)
	ci2 = *ci
	ci2.PluginFolder = "/doesntexist"
	err = lib.ValidateConfig(&ci2)
	assert.Error(t, err)
}

func TestBaseConfig(t *testing.T) {
	ci := lib.CreateGatewayConfig("../../test")
	require.NotNil(t, ci)
	err := lib.ValidateConfig(ci)
	assert.NoError(t, err)
}

func TestLogging(t *testing.T) {
	os.Remove(logFile)
	lib.SetLogging("info", logFile)
	logrus.Info("Hello world")
	assert.FileExists(t, logFile)
}
