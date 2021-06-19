package internal_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/internal"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

func TestStartPlugin(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../test")
	// the binary 'ls' exists on Linux
	pluginName := "/bin/ls"
	cmd := internal.StartPlugin(home, pluginName, []string{})
	assert.NotNil(t, cmd)
	// output, err := cmd.Output()

	// logrus.Infof("Output: %s", output)
}

func TestStartPluginTwice(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../test")
	// the binary 'ls' exists on Linux and Windows
	pluginName := "/bin/sleep"
	cmd := internal.StartPlugin(home, pluginName, []string{"1"})
	require.NotNil(t, cmd)
	time.Sleep(time.Millisecond)
	// second time should fail as only single instances are allowed
	cmd = internal.StartPlugin(home, pluginName, []string{})
	assert.Nil(t, cmd)
}

func TestStartPluginsFromConfig(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../test")
	// the binary 'ls' exists on Linux and Windows
	hc := hubconfig.CreateDefaultHubConfig(home)
	err := hubconfig.LoadConfig(path.Join(hc.ConfigFolder, "hub.yaml"), hc)
	assert.NoError(t, err)
	internal.StartPlugins("", hc.Plugins, []string{})

}

func TestStopPlugin(t *testing.T) {
	pluginName := "sleep"
	cmd := internal.StartPlugin("", pluginName, []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	err := internal.StopPlugin(pluginName)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
}

func TestStopEndedPlugin(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	pluginName := "ls"
	cmd := internal.StartPlugin(home, pluginName, []string{})
	assert.NotNil(t, cmd)
	// 'ls' returns within 1 sec so this attempts to stop a process that has already ended
	time.Sleep(3 * time.Second)
	err := internal.StopPlugin(pluginName)
	// expect plugin not running error
	assert.Error(t, err)
}

func TestStopAllPlugins(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	cmd := internal.StartPlugin(home, "sleep", []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	internal.StopAllPlugins()

}
