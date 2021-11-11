package config_test

import (
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/config"
)

type TestConfig struct {
	mystring string
}

func TestDefaultConfigNoHome(t *testing.T) {
	// This result is unpredictable as it depends on where the binary lives.
	// This changes depends on whether to run as debug, coverage or F5 run
	hc := config.CreateDefaultHubConfig("")
	require.NotNil(t, hc)
	err := config.ValidateHubConfig(hc)
	_ = err // unpredictable outcome
	// assert.NoError(t, err)
	hc = config.CreateDefaultHubConfig("../../test")
	require.NotNil(t, hc)
	_ = err // unpredictable outcome
	// assert.NoError(t, err)

}
func TestDefaultConfigWithHome(t *testing.T) {
	// vscode debug and test runs use different binary folder.
	// Use current dir instead to determine where home is.
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	logrus.Infof("TestDefaultConfig: Current folder is %s", wd)
	hc := config.CreateDefaultHubConfig(home)
	require.NotNil(t, hc)
	err := config.ValidateHubConfig(hc)
	assert.NoError(t, err)
}

func TestLoadHubConfig(t *testing.T) {
	wd, _ := os.Getwd()
	appFolder := path.Join(wd, "../../test")
	hc := config.CreateDefaultHubConfig(appFolder)
	err := config.LoadHubConfig("", "plugin1", hc)
	assert.NoError(t, err)
	err = config.ValidateHubConfig(hc)
	assert.NoError(t, err)
	assert.Equal(t, "info", hc.Loglevel)
}

func TestSubstitute(t *testing.T) {
	substMap := make(map[string]string)
	substMap["{clientID}"] = "plugin1"
	hc := config.HubConfig{}
	wd, _ := os.Getwd()
	templateFile := path.Join(wd, "../../test/config/hub-template.yaml")
	err := config.LoadYamlConfig(templateFile, &hc, substMap)
	assert.NoError(t, err)
	// from the template file
	assert.Equal(t, "/var/log/plugin1.log", hc.LogFile)
}

func TestLoadHubConfigNotFound(t *testing.T) {
	wd, _ := os.Getwd()
	hc := config.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	require.NotNil(t, hc)
	configFile := path.Join(hc.ConfigFolder, "hub-notfound.yaml")
	err := config.LoadYamlConfig(configFile, hc, nil)
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadHubConfigYamlError(t *testing.T) {
	wd, _ := os.Getwd()
	hc := config.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	require.NotNil(t, hc)

	configFile := path.Join(hc.ConfigFolder, "hub-bad.yaml")
	err := config.LoadYamlConfig(configFile, hc, nil)
	// Error should contain info on bad file
	errTxt := err.Error()
	assert.Equal(t, "yaml: line 12", errTxt[:13], "Expected line 12 to be bad")
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadHubConfigBadFolders(t *testing.T) {

	wd, _ := os.Getwd()
	hc := config.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	err := config.ValidateHubConfig(hc)
	assert.NoError(t, err, "Default config should be okay")

	gc2 := *hc
	gc2.AppFolder = "/not/an/app/folder"
	err = config.ValidateHubConfig(&gc2)
	assert.Error(t, err)
	gc2 = *hc
	gc2.ConfigFolder = "./doesntexist"
	err = config.ValidateHubConfig(&gc2)
	assert.Error(t, err)
	gc2 = *hc
	gc2.LogsFolder = "/this/path/doesntexist"
	err = config.ValidateHubConfig(&gc2)
	assert.Error(t, err)
	gc2 = *hc
	gc2.CertsFolder = "./doesntexist"
	err = config.ValidateHubConfig(&gc2)
	assert.Error(t, err)
	gc2 = *hc
	// gc2.PluginFolder = "./doesntexist"
	// err = config.ValidateConfig(&gc2)
	// assert.Error(t, err)
}

func TestLoadAllConfig(t *testing.T) {
	clientID := "plugin1"
	cwd, _ := os.Getwd()
	appFolder := path.Join(path.Dir(cwd), "../test")
	args := os.Args
	appConfig := TestConfig{}
	hubConfig, err := config.LoadAllConfig(args, appFolder, clientID, &appConfig)
	assert.NotNil(t, hubConfig)
	assert.NoError(t, err)
}
