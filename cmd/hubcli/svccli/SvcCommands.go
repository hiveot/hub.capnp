package svccli

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// GetSvcCommands returns the service management commands
func GetSvcCommands(homeFolder string) *cli.Command {

	cmd := &cli.Command{
		//hub prov add|list  <deviceID> <secret>
		Name:  "svc",
		Usage: "Services management",
		Subcommands: []*cli.Command{
			GetSvcListCommand(),
			GetSvcStartCommand(),
			GetSvcStopCommand(),
		},
	}
	return cmd
}

// GetSvcListCommand
// svc list
func GetSvcListCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List Hub services and their status",
		ArgsUsage: ".",
		Action: func(cCtx *cli.Context) error {
			fmt.Println("Listing services")
			return nil
		},
	}
}

// GetSvcStartCommand
// svc start
func GetSvcStartCommand() *cli.Command {
	return &cli.Command{
		Name:      "start",
		Usage:     "Start a Hub service",
		ArgsUsage: "<serviceName>",
		Action: func(cCtx *cli.Context) error {
			fmt.Println("Starting service: ", cCtx.Args().First())
			return nil
		},
	}
}

// GetSvcStopCommand
// svc stop
func GetSvcStopCommand() *cli.Command {
	return &cli.Command{
		Name:      "stop",
		Usage:     "Stop a Hub service",
		ArgsUsage: "<serviceName>",
		Action: func(cCtx *cli.Context) error {
			fmt.Println("Stopping service: ", cCtx.Args().First())
			return nil
		},
	}
}
