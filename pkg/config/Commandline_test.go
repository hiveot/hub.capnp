package config_test

import (
	"flag"
	"os"
	"path"
	"strings"
	"testing"
	_ "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/pkg/config"
)

// CustomConfig as example of how to extend the hub configuration
type CustomConfig struct {
	ExtraVariable string
}

var homeFolder string
var customConfig *CustomConfig
var hubConfig *config.HubConfig

// Use the project app folder during testing
func setup() {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	customConfig = &CustomConfig{}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	// os.Args = append(os.Args[0:1], strings.Split("", " ")...)
	// hubConfig, _ = config.SetupConfig(homeFolder, pluginID, customConfig)
}
func teardown() {
}
func TestSetupHubCommandline(t *testing.T) {
	setup()
	// vscode debug and test runs use different binary folder.
	// Use current dir instead to determine where home is.
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")

	myArgs := strings.Split("--hostname bob --logFile logfile.log --logLevel debug", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig := config.CreateDefaultHubConfig(home)
	config.SetHubCommandlineArgs(hubConfig)
	// hubConfig, err := config.SetupConfig("", nil)

	flag.Parse()
	// assert.NoError(t, err)
	assert.Equal(t, "bob", hubConfig.Messenger.HostPort)
	assert.Equal(t, "logfile.log", hubConfig.Logging.LogFile)
	assert.Equal(t, "debug", hubConfig.Logging.Loglevel)
	// assert.Equal(t, "/etc/cert", hubConfig.Messenger.CertFolder)
}

func TestCommandlineWithError(t *testing.T) {
	setup()
	myArgs := strings.Split("--hostname bob --badarg=bad", " ")
	// myArgs := strings.Split("--hostname bob", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig, err := config.SetupConfig(homeFolder, "", nil)

	assert.Error(t, err, "Parse flag -badarg should fail")
	assert.Equal(t, "bob", hubConfig.Messenger.HostPort)
	teardown()
}

// Test setup with extra commandline flag '--extra'
func TestSetupHubCommandlineWithExtendedConfig(t *testing.T) {
	setup()

	myArgs := strings.Split("-c ./config/hub.yaml --home ../../test --hostname bob --extra value1", " ")
	// Remove testing package commandline arguments so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	// hubConfig := config.CreateDefaultHubConfig("")
	pluginConfig := CustomConfig{}
	// config.HubConfig = *hubConfig

	// config.SetHubCommandlineArgs(&config.HubConfig)
	flag.StringVar(&pluginConfig.ExtraVariable, "extra", "", "Extended extra configuration")

	// err := config.ParseCommandline(myArgs, &config)
	hubConfig, err := config.SetupConfig("", "", pluginConfig)

	assert.NoError(t, err)
	assert.Equal(t, "bob", hubConfig.Messenger.HostPort)
	assert.Equal(t, "value1", pluginConfig.ExtraVariable)
}

// Test with a custom and bad config file
func TestSetupConfigBadConfigfile(t *testing.T) {
	setup()
	// The default directory is the project folder
	myArgs := strings.Split("-c ./config/hub-bad.yaml", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig, err := config.SetupConfig(homeFolder, "", nil)
	assert.Error(t, err)
	assert.Equal(t, "yaml: line 10", err.Error()[0:13], "Expected yaml parse error")
	assert.NotNil(t, hubConfig)
}

// Test with an invalid config file
func TestSetupConfigInvalidConfigfile(t *testing.T) {
	setup()
	myArgs := strings.Split("-c ./config/hub-invalid.yaml", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig, err := config.SetupConfig(homeFolder, "", nil)
	assert.Equal(t, "debug", hubConfig.Logging.Loglevel, "config file wasn't loaded")
	assert.Error(t, err, "Expected validation of config to fail")
	assert.NotNil(t, hubConfig)
}

// TestSetupConfigNoConfig checks that setup still works if the plugin config doesn't exist
func TestSetupConfigNoConfig(t *testing.T) {
	setup()
	myArgs := strings.Split("", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	pluginConfig := CustomConfig{}
	hubConfig, err := config.SetupConfig(homeFolder, "notaconfigfile", pluginConfig)
	assert.NoError(t, err)
	assert.NotNil(t, hubConfig)
}

func TestSetupLogging(t *testing.T) {
	setup()
	myArgs := strings.Split("--logLevel debug", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig, err := config.SetupConfig(homeFolder, "myplugin", nil)
	assert.NoError(t, err)
	require.NotNil(t, hubConfig)
	assert.Equal(t, "debug", hubConfig.Logging.Loglevel)
}
