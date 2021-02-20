package lib_test

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/pkg/lib"
)

// CustomConfig as example of how to extend the gateway configuration
type CustomConfig struct {
	lib.GatewayConfig // embedded
	ExtraVariable     string
}

func TestSetupGatewayCommandline(t *testing.T) {
	myArgs := "--hostname bob --logsFolder logs --loglevel debug --useTLS=False"
	config := lib.CreateGatewayConfig("")
	// erase existing flags to avoid flag redefined error
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	lib.SetupGatewayArgs(config)
	err := lib.ParseCommandline(myArgs, &config)
	assert.NoError(t, err)
	assert.Equal(t, "bob", config.Messenger.HostPort)
	assert.Equal(t, "logs", config.Logging.LogsFolder)
	assert.Equal(t, "debug", config.Logging.Loglevel)
	assert.Equal(t, false, config.Messenger.UseTLS)
}

// func TestCommandlineWithError(t *testing.T) {
// 	myArgs := "--hostname bob --badarg=bad"
// 	config := lib.CreateGatewayConfig("")
// 	lib.SetupGatewayArgs(config)
// 	err := lib.ParseCommandline(myArgs, &config)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "bob", config.HostPort)
// }

func TestSetupGatewayCommandlineWithExtendedConfig(t *testing.T) {
	myArgs := "--hostname bob --extra value1"

	gwConfig := lib.CreateGatewayConfig("")
	config := CustomConfig{}
	config.GatewayConfig = *gwConfig

	// erase existing flags to avoid flag redefined error
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	lib.SetupGatewayArgs(&config.GatewayConfig)
	flag.StringVar(&config.ExtraVariable, "extra", "", "Extended extra configuration")

	err := lib.ParseCommandline(myArgs, &config)
	assert.NoError(t, err)
	assert.Equal(t, "bob", config.Messenger.HostPort)
	assert.Equal(t, "value1", config.ExtraVariable)
}
