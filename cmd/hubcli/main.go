package main

import (
	"log"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/wostzone/hub/cmd/hubcli/certcli"
	"github.com/wostzone/hub/cmd/hubcli/provcli"
	"github.com/wostzone/hub/cmd/hubcli/svccli"
)

const Version = `0.4-alpha`

var binFolder string
var homeFolder string

// CLI Main entry
func main() {
	logrus.SetLevel(logrus.InfoLevel)
	binFolder = path.Dir(os.Args[0])
	homeFolder = path.Dir(binFolder)

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "hubcli",
		Usage:                "Hub Commandline Interface",
		Version:              Version,
		Commands: []*cli.Command{
			certcli.GetCertCommands(homeFolder),
			provcli.GetProvCommands(homeFolder),
			svccli.GetSvcCommands(homeFolder),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
