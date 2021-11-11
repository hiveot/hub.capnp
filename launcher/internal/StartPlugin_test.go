package internal_test

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/launcher/internal"
	"github.com/wostzone/hub/lib/client/pkg/config"
)

//--- THIS USES TestMain from StartHub_test.go ---

func TestStartPlugin(t *testing.T) {
	// the binary 'ls' exists on Linux
	pluginName := "/bin/ls"
	cmd := internal.StartPlugin(homeFolder, pluginName, []string{})
	assert.NotNil(t, cmd)
	// output, err := cmd.Output()

	// logrus.Infof("Output: %s", output)
}

func TestStartPluginTwice(t *testing.T) {

	// the binary 'ls' exists on Linux and Windows
	pluginName := "/bin/sleep"
	cmd := internal.StartPlugin(homeFolder, pluginName, []string{"1"})
	require.NotNil(t, cmd)
	time.Sleep(time.Millisecond)
	// second time should fail as only single instances are allowed
	cmd = internal.StartPlugin(homeFolder, pluginName, []string{})
	assert.Nil(t, cmd)
	// wait until the first sleep ends
	time.Sleep(time.Second)
}

func TestStartPluginsFromConfig(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	hc := config.CreateDefaultHubConfig(homeFolder)
	pluginConfig := internal.PluginConfig{}
	err := config.LoadYamlConfig(path.Join(hc.ConfigFolder, "launcher.yaml"), &pluginConfig, nil)
	assert.NoError(t, err)
	internal.StartPlugins("", pluginConfig.Plugins, []string{})

}

func TestStopPlugin(t *testing.T) {
	pluginName := "/bin/sleep"
	cmd := internal.StartPlugin("", pluginName, []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	err := internal.StopPlugin(pluginName)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
}

func TestStopEndedPlugin(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	pluginName := "/bin/ls"
	cmd := internal.StartPlugin(homeFolder, pluginName, []string{})
	assert.NotNil(t, cmd)
	// 'ls' returns within 1 sec so this attempts to stop a process that has already ended
	time.Sleep(3 * time.Second)
	err := internal.StopPlugin(pluginName)
	// expect plugin not running error
	assert.Error(t, err)
}

func TestStopAllPlugins(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	cmd := internal.StartPlugin(homeFolder, "/bin/sleep", []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	internal.StopAllPlugins()

}
