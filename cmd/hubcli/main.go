package main

import (
	"context"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/cmd/hubcli/authncli"
	"github.com/hiveot/hub/cmd/hubcli/authzcli"
	"github.com/hiveot/hub/cmd/hubcli/certscli"
	"github.com/hiveot/hub/cmd/hubcli/directorycli"
	"github.com/hiveot/hub/cmd/hubcli/gatewaycli"
	"github.com/hiveot/hub/cmd/hubcli/historycli"
	"github.com/hiveot/hub/cmd/hubcli/launchercli"
	"github.com/hiveot/hub/cmd/hubcli/provcli"
	"github.com/hiveot/hub/cmd/hubcli/pubsubcli"
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
	f, _, _ := svcconfig.SetupFolderConfig("hubcli")

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
			launchercli.LauncherListCommand(ctx, f),
			launchercli.LauncherStartCommand(ctx, f),
			launchercli.LauncherStopCommand(ctx, f),

			authncli.AuthnListUsersCommand(ctx, f),
			authncli.AuthnAddUserCommand(ctx, f),
			authncli.AuthnRemoveUserCommand(ctx, f),

			authzcli.AuthzListGroupsCommand(ctx, f),
			//authzcli.AuthzSetClientRoleCommand(ctx, f),
			//authzcli.AuthzRemoveClientCommand(ctx, f),

			certscli.CreateCACommand(ctx, f.Certs),
			certscli.ViewCACommand(ctx, f.Certs),
			certscli.CertCreateDeviceCommands(ctx, f),
			certscli.CertsCreateServiceCommand(ctx, f),
			certscli.CertsCreateUserCommand(ctx, f),
			certscli.CertsShowInfoCommand(ctx, f),

			pubsubcli.SubTDCommand(ctx, f),
			pubsubcli.SubEventsCommand(ctx, f),
			//pubsubcli.PubActionCommand(ctx, f),

			directorycli.DirectoryListCommand(ctx, f),

			//historycli.HistoryCommands(ctx, f),
			historycli.HistoryInfoCommand(ctx, f),
			historycli.HistoryListCommand(ctx, f),
			historycli.HistoryLatestCommand(ctx, f),
			historycli.HistoryRetainCommand(ctx, f),

			provcli.ProvisionAddOOBSecretsCommand(ctx, f),
			provcli.ProvisionApproveRequestCommand(ctx, f),
			provcli.ProvisionGetPendingRequestsCommand(ctx, f),
			provcli.ProvisionGetApprovedRequestsCommand(ctx, f),

			gatewaycli.GatewayListCommand(ctx, f),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Error("ERROR: ", err)
		helpArgs := append(os.Args, "-h")
		app.Run(helpArgs)
	}
}
