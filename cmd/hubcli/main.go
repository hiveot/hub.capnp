package main

import (
	"context"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/cmd/hubcli/certcli"
	"github.com/hiveot/hub/cmd/hubcli/launchercli"
	"github.com/hiveot/hub/cmd/hubcli/provcli"
	"github.com/hiveot/hub/internal/folders"
)

const Version = `0.5-alpha`

var binFolder string
var homeFolder string

// CLI Main entry
func main() {
	logrus.SetLevel(logrus.InfoLevel)
	binFolder = path.Dir(os.Args[0])
	homeFolder = path.Dir(binFolder)
	f := folders.GetFolders(homeFolder, false)
	ctx := context.Background()

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "hubcli",
		Usage:                "Hub Commandline Interface",
		Version:              Version,

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "home",
				Usage:       "Path to home `folder`.",
				Value:       f.Home,
				Destination: &f.Home,
			},
			&cli.StringFlag{
				Name:        "services",
				Usage:       "Path to services directory",
				Value:       f.Services,
				Destination: &f.Services,
			},
			&cli.StringFlag{
				Name:        "certs",
				Usage:       "Path to certificate `folder`.",
				Value:       f.Certs,
				Destination: &f.Certs,
			},
		},
		Commands: []*cli.Command{
			certcli.CACommands(ctx, f),
			certcli.CertCommands(ctx, f),
			launchercli.LauncherCommands(ctx, f),
			provcli.ProvisioningCommands(ctx, f),
			//svccli.GetSvcCommands(homeFolder),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Error("ERROR: ", err)
	}
}
