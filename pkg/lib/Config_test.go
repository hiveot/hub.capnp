package lib_test

import (
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/pkg/lib"
)

type ConfigType1 struct {
	C1 string
	c2 string
}

func TestDefaultConfigNoHome(t *testing.T) {
	// This result is unpredictable as it depends on where the binary lives.
	// This changes depends on whether to run as debug, coverage or F5 run
	gc := lib.CreateDefaultGatewayConfig("")
	require.NotNil(t, gc)
	err := lib.ValidateConfig(gc)
	_ = err // unpredictable outcome
	// assert.NoError(t, err)
	gc = lib.CreateDefaultGatewayConfig("./")
	require.NotNil(t, gc)
	_ = err // unpredictable outcome
	// assert.NoError(t, err)

}
func TestDefaultConfigWithHome(t *testing.T) {
	// vscode debug and test runs use different binary folder.
	// Use current dir instead to determine where home is.
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	logrus.Infof("TestDefaultConfig: Current folder is %s", wd)
	gc := lib.CreateDefaultGatewayConfig(home)
	require.NotNil(t, gc)
	err := lib.ValidateConfig(gc)
	assert.NoError(t, err)
}

func TestLoadGatewayConfig(t *testing.T) {
	wd, _ := os.Getwd()
	gc := lib.CreateDefaultGatewayConfig(path.Join(wd, "../../test"))
	require.NotNil(t, gc)

	configFile := path.Join(gc.ConfigFolder, "gateway.yaml")
	err := lib.LoadConfig(configFile, gc)
	assert.NoError(t, err)
	err = lib.ValidateConfig(gc)
	assert.NoError(t, err)
	assert.Equal(t, "info", gc.Logging.Loglevel)
}

func TestLoadGatewayConfigNotFound(t *testing.T) {
	wd, _ := os.Getwd()
	gc := lib.CreateDefaultGatewayConfig(path.Join(wd, "../../test"))
	require.NotNil(t, gc)
	configFile := path.Join(gc.ConfigFolder, "gateway-notfound.yaml")
	err := lib.LoadConfig(configFile, gc)
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadGatewayConfigYamlError(t *testing.T) {
	wd, _ := os.Getwd()
	gc := lib.CreateDefaultGatewayConfig(path.Join(wd, "../../test"))
	require.NotNil(t, gc)

	configFile := path.Join(gc.ConfigFolder, "gateway-bad.yaml")
	err := lib.LoadConfig(configFile, gc)
	// Error should contain info on bad file
	errTxt := err.Error()
	assert.Equal(t, "yaml: line 11", errTxt[:13], "Expected line 11 to be bad")
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadGatewayConfigBadFolders(t *testing.T) {

	wd, _ := os.Getwd()
	gc := lib.CreateDefaultGatewayConfig(path.Join(wd, "../../test"))
	err := lib.ValidateConfig(gc)
	assert.NoError(t, err, "Default config should be okay")

	gc2 := *gc
	gc2.Home = "/not/a/home/folder"
	err = lib.ValidateConfig(&gc2)
	assert.Error(t, err)
	gc2 = *gc
	gc2.ConfigFolder = "./doesntexist"
	err = lib.ValidateConfig(&gc2)
	assert.Error(t, err)
	gc2 = *gc
	gc2.Logging.LogFile = "/this/path/doesntexist"
	err = lib.ValidateConfig(&gc2)
	assert.Error(t, err)
	gc2 = *gc
	gc2.Messenger.CertsFolder = "./doesntexist"
	err = lib.ValidateConfig(&gc2)
	assert.Error(t, err)
	gc2 = *gc
	gc2.PluginFolder = "./doesntexist"
	err = lib.ValidateConfig(&gc2)
	assert.Error(t, err)
}

func TestLogging(t *testing.T) {
	wd, _ := os.Getwd()
	logFile := path.Join(wd, "../../test/logs/TestLogging.log")

	os.Remove(logFile)
	lib.SetLogging("info", logFile)
	logrus.Info("Hello world")
	assert.FileExists(t, logFile)
	os.Remove(logFile)
}
