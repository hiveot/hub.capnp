package hub_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/pkg/config"
	"github.com/wostzone/hub/pkg/hub"
)

func TestStartPlugin(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	pluginName := "ls"
	cmd := hub.StartPlugin(home, pluginName, []string{})
	assert.NotNil(t, cmd)
	// output, err := cmd.Output()

	// logrus.Infof("Output: %s", output)
}

func TestStartPluginsFromConfig(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	hubConfig := config.CreateDefaultHubConfig(home)
	err := config.LoadConfig(path.Join(hubConfig.ConfigFolder, "hub.yaml"), hubConfig)
	assert.NoError(t, err)
	hub.StartPlugins("", hubConfig.Plugins, []string{})

}

func TestStopPlugin(t *testing.T) {
	pluginName := "sleep"
	cmd := hub.StartPlugin("", pluginName, []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	err := hub.StopPlugin(pluginName)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
}

func TestStopEndedPlugin(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	pluginName := "ls"
	cmd := hub.StartPlugin(home, pluginName, []string{})
	assert.NotNil(t, cmd)
	// 'ls' returns within 1 sec so this attempts to stop a process that has already ended
	time.Sleep(3 * time.Second)
	err := hub.StopPlugin(pluginName)
	// expect plugin not running error
	assert.Error(t, err)
}

func TestStopAllPlugins(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	cmd := hub.StartPlugin(home, "sleep", []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	hub.StopAllPlugins()

}
