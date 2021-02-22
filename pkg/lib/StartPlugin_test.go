package lib_test

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/pkg/lib"
)

func TestStartPlugin(t *testing.T) {
	// the binary 'ls' exists on Linux and Windows
	pluginName := "ls"
	cmd := lib.StartPlugin("", pluginName, []string{})
	err := cmd.Run()
	assert.NoError(t, err)
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
