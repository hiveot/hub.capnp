package config_test

import (
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/pkg/config"
)

type ConfigType1 struct {
	C1 string
	c2 string
}

func TestDefaultConfigNoHome(t *testing.T) {
	// This result is unpredictable as it depends on where the binary lives.
	// This changes depends on whether to run as debug, coverage or F5 run
	gc := config.CreateDefaultHubConfig("")
	require.NotNil(t, gc)
	err := config.ValidateConfig(gc)
	_ = err // unpredictable outcome
	// assert.NoError(t, err)
	gc = config.CreateDefaultHubConfig("./")
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
	gc := config.CreateDefaultHubConfig(home)
	require.NotNil(t, gc)
	err := config.ValidateConfig(gc)
	assert.NoError(t, err)
}

func TestLoadHubConfig(t *testing.T) {
	wd, _ := os.Getwd()
	gc := config.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	require.NotNil(t, gc)

	configFile := path.Join(gc.ConfigFolder, "hub.yaml")
	err := config.LoadConfig(configFile, gc)
	assert.NoError(t, err)
	err = config.ValidateConfig(gc)
	assert.NoError(t, err)
	assert.Equal(t, "info", gc.Logging.Loglevel)
}

func TestLoadHubConfigNotFound(t *testing.T) {
	wd, _ := os.Getwd()
	gc := config.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	require.NotNil(t, gc)
	configFile := path.Join(gc.ConfigFolder, "hub-notfound.yaml")
	err := config.LoadConfig(configFile, gc)
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadHubConfigYamlError(t *testing.T) {
	wd, _ := os.Getwd()
	gc := config.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	require.NotNil(t, gc)

	configFile := path.Join(gc.ConfigFolder, "hub-bad.yaml")
	err := config.LoadConfig(configFile, gc)
	// Error should contain info on bad file
	errTxt := err.Error()
	assert.Equal(t, "yaml: line 10", errTxt[:13], "Expected line 10 to be bad")
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadHubConfigBadFolders(t *testing.T) {

	wd, _ := os.Getwd()
	gc := config.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	err := config.ValidateConfig(gc)
	assert.NoError(t, err, "Default config should be okay")

	gc2 := *gc
	gc2.Home = "/not/a/home/folder"
	err = config.ValidateConfig(&gc2)
	assert.Error(t, err)
	gc2 = *gc
	gc2.ConfigFolder = "./doesntexist"
	err = config.ValidateConfig(&gc2)
	assert.Error(t, err)
	gc2 = *gc
	gc2.Logging.LogFile = "/this/path/doesntexist"
	err = config.ValidateConfig(&gc2)
	assert.Error(t, err)
	gc2 = *gc
	gc2.Messenger.CertFolder = "./doesntexist"
	err = config.ValidateConfig(&gc2)
	assert.Error(t, err)
	gc2 = *gc
	gc2.PluginFolder = "./doesntexist"
	err = config.ValidateConfig(&gc2)
	assert.Error(t, err)
}

func TestLogging(t *testing.T) {
	wd, _ := os.Getwd()
	logFile := path.Join(wd, "../../test/logs/TestLogging.log")

	os.Remove(logFile)
	config.SetLogging("info", logFile, "")
	logrus.Info("Hello world")
	assert.FileExists(t, logFile)
	os.Remove(logFile)
}
