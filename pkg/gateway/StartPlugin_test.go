package gateway_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/pkg/config"
	"github.com/wostzone/gateway/pkg/gateway"
)

func TestStartPlugin(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	pluginName := "ls"
	cmd := gateway.StartPlugin(home, pluginName, []string{})
	assert.NotNil(t, cmd)
	// output, err := cmd.Output()

	// logrus.Infof("Output: %s", output)
}

func TestStartPluginsFromConfig(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	gwConfig := config.CreateDefaultGatewayConfig(home)
	err := config.LoadConfig(path.Join(gwConfig.ConfigFolder, "gateway.yaml"), gwConfig)
	assert.NoError(t, err)
	gateway.StartPlugins("", gwConfig.Plugins, []string{})

}

func TestStopPlugin(t *testing.T) {
	pluginName := "sleep"
	cmd := gateway.StartPlugin("", pluginName, []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	err := gateway.StopPlugin(pluginName)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
}

func TestStopEndedPlugin(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	pluginName := "ls"
	cmd := gateway.StartPlugin(home, pluginName, []string{})
	assert.NotNil(t, cmd)
	// 'ls' returns within 1 sec so this attempts to stop a process that has already ended
	time.Sleep(3 * time.Second)
	err := gateway.StopPlugin(pluginName)
	// expect plugin not running error
	assert.Error(t, err)
}

func TestStopAllPlugins(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// the binary 'ls' exists on Linux and Windows
	cmd := gateway.StartPlugin(home, "sleep", []string{"10"})
	assert.NotNil(t, cmd)
	time.Sleep(1 * time.Second)
	gateway.StopAllPlugins()

}
