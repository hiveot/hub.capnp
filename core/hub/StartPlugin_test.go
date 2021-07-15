package hub_test

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/core/hub"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

//--- THIS USES TestMain from StartHub_test.go ---

func TestStartPlugin(t *testing.T) {
	// the binary 'ls' exists on Linux
	pluginName := "/bin/ls"
	cmd := hub.StartPlugin(homeFolder, pluginName, []string{})
	assert.NotNil(t, cmd)
	// output, err := cmd.Output()

	// logrus.Infof("Output: %s", output)
}

func TestStartPluginTwice(t *testing.T) {

	// the binary 'ls' exists on Linux and Windows
	pluginName := "/bin/sleep"
	cmd := hub.StartPlugin(homeFolder, pluginName, []string{"1"})
	require.NotNil(t, cmd)
	time.Sleep(time.Millisecond)
	// second time should fail as only single instances are allowed
	cmd = hub.StartPlugin(homeFolder, pluginName, []string{})
	assert.Nil(t, cmd)
	// wait until the first sleep ends
	time.Sleep(time.Second)
}

func TestStartPluginsFromConfig(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	hc := hubconfig.CreateDefaultHubConfig(homeFolder)
	err := hubconfig.LoadConfig(path.Join(hc.ConfigFolder, "hub.yaml"), hc, nil)
	assert.NoError(t, err)
	hub.StartPlugins("", hc.Plugins, []string{})

}

func TestStopPlugin(t *testing.T) {
	pluginName := "/bin/sleep"
	cmd := hub.StartPlugin("", pluginName, []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	err := hub.StopPlugin(pluginName)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
}

func TestStopEndedPlugin(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	pluginName := "/bin/ls"
	cmd := hub.StartPlugin(homeFolder, pluginName, []string{})
	assert.NotNil(t, cmd)
	// 'ls' returns within 1 sec so this attempts to stop a process that has already ended
	time.Sleep(3 * time.Second)
	err := hub.StopPlugin(pluginName)
	// expect plugin not running error
	assert.Error(t, err)
}

func TestStopAllPlugins(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	cmd := hub.StartPlugin(homeFolder, "/bin/sleep", []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	hub.StopAllPlugins()

}
