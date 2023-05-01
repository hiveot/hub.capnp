package main

import (
	"context"
	"fmt"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/lib/utils"
	"github.com/sirupsen/logrus"
	"os"
	"path"

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
)

const Version = `0.5-alpha`

var binFolder string
var homeFolder string
var runFolder string
var certsFolder string
var configFolder string
var nowrap bool

// CLI Main entry
func main() {
	//logging.SetLogging("info", "")
	binFolder = path.Dir(os.Args[0])
	homeFolder = path.Dir(binFolder)
	nowrap = false
	ctx := context.Background()
	f := svcconfig.GetFolders(homeFolder, false)
	certsFolder = f.Certs
	configFolder = f.Config

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "hubcli",
		Usage:                "Hub Commandline Interface",
		Version:              Version,

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "home",
				Usage:       "Path to home `folder`",
				Value:       homeFolder,
				Destination: &homeFolder,
			},
			&cli.BoolFlag{
				Name:        "nowrap",
				Usage:       "Disable konsole wrapping",
				Value:       nowrap,
				Destination: &nowrap,
			},
		},
		Before: func(c *cli.Context) error {
			f = svcconfig.GetFolders(homeFolder, false)
			certsFolder = f.Certs
			runFolder = f.Run
			configFolder = f.Config
			if nowrap {
				fmt.Printf(utils.WrapOff)
			}
			return nil
		},
		Commands: []*cli.Command{
			launchercli.LauncherListCommand(ctx, &runFolder),
			launchercli.LauncherStartCommand(ctx, &runFolder),
			launchercli.LauncherStopCommand(ctx, &runFolder),

			authncli.AuthnListUsersCommand(ctx, &runFolder),
			authncli.AuthnAddUserCommand(ctx, &runFolder),
			authncli.AuthnRemoveUserCommand(ctx, &runFolder),
			authncli.AuthnPasswordCommand(ctx, &runFolder),

			authzcli.AuthzListGroupsCommand(ctx, &runFolder),
			//authzcli.AuthzSetClientRoleCommand(ctx, &certsFolder),
			//authzcli.AuthzRemoveClientCommand(ctx, &certsFolder),

			certscli.CreateCACommand(ctx, &certsFolder),
			certscli.ViewCACommand(ctx, &certsFolder),
			certscli.CertCreateDeviceCommands(ctx, &runFolder),
			certscli.CertsCreateServiceCommand(ctx, &runFolder),
			certscli.CertsCreateUserCommand(ctx, &runFolder),
			certscli.CertsShowInfoCommand(ctx, &runFolder),

			pubsubcli.SubTDCommand(ctx, &runFolder),
			pubsubcli.SubEventsCommand(ctx, &runFolder),
			pubsubcli.PubActionCommand(ctx, &runFolder),

			directorycli.DirectoryListCommand(ctx, &runFolder),

			//historycli.HistoryCommands(ctx, &runFolder),
			historycli.HistoryInfoCommand(ctx, &runFolder),
			historycli.HistoryListCommand(ctx, &runFolder),
			historycli.HistoryLatestCommand(ctx, &runFolder),
			historycli.HistoryRetainCommand(ctx, &runFolder),

			provcli.ProvisionAddOOBSecretsCommand(ctx, &runFolder),
			provcli.ProvisionApproveRequestCommand(ctx, &runFolder),
			provcli.ProvisionGetPendingRequestsCommand(ctx, &runFolder),
			provcli.ProvisionGetApprovedRequestsCommand(ctx, &runFolder),

			gatewaycli.GatewayListCommand(ctx, &certsFolder, &configFolder),
		},
	}

	// Show the arguments in the command line
	cli.AppHelpTemplate = `NAME:
  {{.Name}} - {{.Usage}}
USAGE:
  {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
  {{if len .Authors}}
AUTHOR:
  {{range .Authors}}{{ . }}{{end}}
  {{end}}{{if .Commands}}
COMMANDS: {{range .VisibleCategories}}{{if .Name}}
   {{.Name }}:{{"\t"}}{{range .VisibleCommands}}
      {{join .Names ", "}} {{.ArgsUsage}} {{"\t"}}{{.Usage}}{{end}}{{else}}{{template "visibleCommandTemplate" .}}{{end}}{{end}}

GLOBAL OPTIONS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
{{end}}
`
	app.Suggest = true
	app.HideHelpCommand = true
	if err := app.Run(os.Args); err != nil {
		logrus.Error("ERROR: ", err)
		helpArgs := append(os.Args, "-h")
		_ = app.Run(helpArgs)
	}
}
