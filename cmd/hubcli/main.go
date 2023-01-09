package main

import (
	"context"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/cmd/hubcli/authn"
	"github.com/hiveot/hub/cmd/hubcli/authz"
	"github.com/hiveot/hub/cmd/hubcli/certscli"
	"github.com/hiveot/hub/cmd/hubcli/directorycli"
	"github.com/hiveot/hub/cmd/hubcli/gatewaycli"
	"github.com/hiveot/hub/cmd/hubcli/historycli"
	"github.com/hiveot/hub/cmd/hubcli/launchercli"
	"github.com/hiveot/hub/cmd/hubcli/provcli"
	"github.com/hiveot/hub/lib/svcconfig"
)

const Version = `0.5-alpha`

var binFolder string
var homeFolder string

// CLI Main entry
func main() {
	//logging.SetLogging("info", "")
	binFolder = path.Dir(os.Args[0])
	homeFolder = path.Dir(binFolder)
	ctx := context.Background()
	f, _, _ := svcconfig.LoadServiceConfig("hubcli", false, nil)

	//logrus.Infof("folders is %v", f)
	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "hubcli",
		Usage:                "Hub Commandline Interface",
		Version:              Version,

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "home",
				Usage:       "Path to home `folder`.",
				Value:       homeFolder,
				Destination: &homeFolder,
			},
		},
		Commands: []*cli.Command{
			certscli.CACommands(ctx, f),
			launchercli.LauncherCommands(ctx, f),
			authn.AuthnCommands(ctx, f),
			authz.AuthzCommands(ctx, f),
			certscli.CertCommands(ctx, f),
			directorycli.DirectoryCommands(ctx, f),
			historycli.HistoryCommands(ctx, f),
			provcli.ProvisioningCommands(ctx, f),
			gatewaycli.GatewayCommands(ctx, f),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Error("ERROR: ", err)
	}
}
