package main

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
	smbserver "github.com/wostzone/gateway/plugins/smbserver/internal"
)

func defineCommandline() {

}

// parse commandline, load configuration and start the plugin
func main() {
	level := "info"
	filename := ""

	config := &smbserver.SmbusConfig{}
	gwConfig := lib.CreateDefaultGatewayConfig("")
	config.GatewayConfig = *gwConfig
	filename = path.Join(config.ConfigFolder, "gateway.yaml")
	err := lib.LoadConfig(filename, config)

	lib.SetGatewayCommandlineArgs(&config.GatewayConfig)
	// Add optional arguments here
	// flag.StringVar(&config.ExtraVariable, "extra", "", "Extended extra configuration")
	flag.Parse()

	if err != nil {
		logrus.Errorf("Smbserver: Error loading config: %s", err)
		os.Exit(1)
	}
	lib.SetLogging(level, filename)
	server, err := smbserver.StartSmbusServer(config)
	if err == nil {
		lib.WaitForSignal()
		server.Stop()
	}
}
