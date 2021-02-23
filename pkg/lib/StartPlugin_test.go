package lib_test

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/pkg/lib"
)

func TestStartPlugin(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	pluginName := "ls"
	cmd := lib.StartPlugin("", pluginName, []string{})
	assert.NotNil(t, cmd)
	// output, err := cmd.Output()

	// logrus.Infof("Output: %s", output)
}

func TestStartPluginsFromConfig(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	config := lib.CreateDefaultGatewayConfig("../../test")
	err := lib.LoadConfig(path.Join(config.ConfigFolder, "gateway.yaml"), config)
	assert.NoError(t, err)
	lib.StartPlugins("", config.Plugins, []string{})

}

func TestStopPlugin(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	pluginName := "sleep"
	cmd := lib.StartPlugin("", pluginName, []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	err := lib.StopPlugin(pluginName)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
}

func TestStopEndedPlugin(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	pluginName := "ls"
	cmd := lib.StartPlugin("", pluginName, []string{})
	assert.NotNil(t, cmd)
	// 'ls' returns within 1 sec so this attempts to stop a process that has already ended
	time.Sleep(3 * time.Second)
	err := lib.StopPlugin(pluginName)
	// expect plugin not running error
	assert.Error(t, err)
}

func TestStopAllPlugins(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	cmd := lib.StartPlugin("", "sleep", []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	lib.StopAllPlugins()

}
