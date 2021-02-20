package main

import (
	"flag"

	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/pkg/logging"
)

func defineCommandline() {

}

// parse commandline, load configuration and start the plugin
func main() {
	level := "info"
	filename := ""
	config := lib.GatewayConfig{}
	lib.SetupGatewayArgs(&config)
	flag.Parse()
	// lib.ParseCommandline(&config)

	err := config.LoadConfig(&config)
	logging.SetLogging(level, filename)
	StartSimpleMessageBus(&config)
	WaitForSignal()
}
