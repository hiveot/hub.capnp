package lib_test

import (
	"flag"
	"os"
	"path"
	"strings"
	"testing"
	_ "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/pkg/lib"
)

// CustomConfig as example of how to extend the gateway configuration
type CustomConfig struct {
	lib.GatewayConfig // embedded
	ExtraVariable     string
}

func TestSetupGatewayCommandline(t *testing.T) {
	// vscode debug and test runs use different binary folder.
	// Use current dir instead to determine where home is.
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")

	myArgs := strings.Split("--hostname bob --logFile logfile.log --useTLS=False --logLevel debug", " ")
	// Remove testing package created commandline and flags so we can test ours
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	gwConfig := lib.CreateDefaultGatewayConfig(home)
	lib.SetGatewayCommandlineArgs(gwConfig)
	// gwConfig, err := lib.SetupConfig("", nil)

	flag.Parse()
	// assert.NoError(t, err)
	assert.Equal(t, "bob", gwConfig.Messenger.HostPort)
	assert.Equal(t, "logfile.log", gwConfig.Logging.LogFile)
	assert.Equal(t, "debug", gwConfig.Logging.Loglevel)
	assert.Equal(t, false, gwConfig.Messenger.UseTLS)
}

func TestCommandlineWithError(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	myArgs := strings.Split("--hostname bob --badarg=bad", " ")
	// Remove testing package created commandline and flags so we can test ours
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	gwConfig, err := lib.SetupConfig(home, "", nil)

	assert.Error(t, err, "Parse flag -badarg should fail")
	assert.Equal(t, "bob", gwConfig.Messenger.HostPort)
}

// Test setup with extra commandline flag '--extra'
func TestSetupGatewayCommandlineWithExtendedConfig(t *testing.T) {
	// wd, _ := os.Getwd()
	// home := path.Join(wd, "../../test")
	myArgs := strings.Split("-c ./config/gateway.yaml --home ../../test --hostname bob --extra value1", " ")
	// Remove testing package commandline arguments so we can test ours
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	// gwConfig := lib.CreateDefaultGatewayConfig("")
	pluginConfig := CustomConfig{}
	// config.GatewayConfig = *gwConfig

	// lib.SetGatewayCommandlineArgs(&config.GatewayConfig)
	flag.StringVar(&pluginConfig.ExtraVariable, "extra", "", "Extended extra configuration")

	// err := lib.ParseCommandline(myArgs, &config)
	gwConfig, err := lib.SetupConfig("", "", pluginConfig)

	assert.NoError(t, err)
	assert.Equal(t, "bob", gwConfig.Messenger.HostPort)
	assert.Equal(t, "value1", pluginConfig.ExtraVariable)
}

// Test with a custom and bad config file
func TestSetupConfigBadConfigfile(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	// The default directory is the project folder
	myArgs := strings.Split("-c ./config/gateway-bad.yaml", " ")
	// Remove testing package created commandline and flags so we can test ours
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	gwConfig, err := lib.SetupConfig(home, "", nil)
	assert.Error(t, err)
	assert.Equal(t, "yaml: line 11", err.Error()[0:13], "Expected yaml parse error")
	assert.NotNil(t, gwConfig)
}

// Test with an invalid config file
func TestSetupConfigInvalidConfigfile(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")

	myArgs := strings.Split("-c ./config/gateway-invalid.yaml", " ")
	// Remove testing package created commandline and flags so we can test ours
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	gwConfig, err := lib.SetupConfig(home, "", nil)
	assert.Equal(t, "debug", gwConfig.Logging.Loglevel, "config file wasn't loaded")
	assert.Error(t, err, "Expected validation of config to fail")
	assert.NotNil(t, gwConfig)
}

// TestSetupConfigNoConfig checks that setup still works if the plugin config doesn't exist
func TestSetupConfigNoConfig(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	myArgs := strings.Split("", " ")
	// Remove testing package created commandline and flags so we can test ours
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	pluginConfig := CustomConfig{}
	gwConfig, err := lib.SetupConfig(home, "notaconfigfile", pluginConfig)
	assert.NoError(t, err)
	assert.NotNil(t, gwConfig)
}

func TestSetupLogging(t *testing.T) {
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	myArgs := strings.Split("--logLevel debug", " ")
	// Remove testing package created commandline and flags so we can test ours
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	gwConfig, err := lib.SetupConfig(home, "myplugin", nil)
	assert.NoError(t, err)
	require.NotNil(t, gwConfig)
	assert.Equal(t, "debug", gwConfig.Logging.Loglevel)
}
