package launchercli

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnpclient"
)

func LauncherCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	cmd := &cli.Command{
		Name:  "launcher",
		Usage: "Connect stop Hub services",
		Subcommands: []*cli.Command{
			LauncherListCommand(ctx, f),
			LauncherStartCommand(ctx, f),
			LauncherStopCommand(ctx, f),
		},
	}
	return cmd
}

func LauncherListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {

	return &cli.Command{
		Name:      "list",
		Usage:     "List services",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListServices(ctx, f)
			return err
		},
	}
}

func LauncherStartCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {

	return &cli.Command{
		Name:      "start",
		Usage:     "Connect a service or all services",
		ArgsUsage: "start <serviceName> | all",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 1 {
				return fmt.Errorf("expected service name")
			}
			err := HandleStartService(ctx, f, cCtx.Args().First())
			return err
		},
	}
}

func LauncherStopCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {

	return &cli.Command{
		Name:      "stop",
		Usage:     "Stop a running service or all services",
		ArgsUsage: "stop <serviceName> | all",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 1 {
				return fmt.Errorf("expected service name")
			}
			err := HandleStopService(ctx, f, cCtx.Args().First())
			return err
		},
	}
}

// HandleListServices prints a list of available services
func HandleListServices(ctx context.Context, f svcconfig.AppFolders) error {
	var ls launcher.ILauncher

	conn, err := hubclient.CreateLocalClientConnection(launcher.ServiceName, f.Run)
	if err == nil {
		ls, err = capnpclient.NewLauncherCapnpClient(ctx, conn)
	}
	if err != nil {
		return err
	}

	fmt.Println("Service                      Size   Starts       PID    CPU   Memory   Status    Last Error")
	fmt.Println("-------                      ----   ------   -------   ----   ------   -------   -----------")
	entries, _ := ls.List(ctx, false)
	for _, entry := range entries {
		status := "stopped"
		cpu := ""
		memory := ""
		pid := fmt.Sprintf("%d", entry.PID)
		cpu = fmt.Sprintf("%d%%", entry.CPU)
		memory = fmt.Sprintf("%d MB", entry.RSS/1024/1024)
		if entry.Running {
			status = "running"
		}
		fmt.Printf("%-25s %4d MB   %6d   %7s   %4s   %6s   %6s   %s\n",
			entry.Name,
			entry.Size/1024/1024,
			entry.StartCount,
			pid,
			cpu,
			memory,
			status,
			entry.Status,
		)
	}
	return nil
}

// HandleStartService starts a service
func HandleStartService(ctx context.Context, f svcconfig.AppFolders, serviceName string) error {
	var ls launcher.ILauncher

	conn, err := hubclient.CreateLocalClientConnection(launcher.ServiceName, f.Run)

	if err == nil {
		ls, err = capnpclient.NewLauncherCapnpClient(ctx, conn)
	}
	if err != nil {
		return err
	}

	if serviceName == "all" {
		err = ls.StartAll(ctx)

		if err != nil {
			//fmt.Println("Connect all failed with: ", err)
			return err
		}
		fmt.Printf("All services started\n")
	} else {
		info, err2 := ls.StartService(ctx, serviceName)

		if err2 != nil {
			//fmt.Println("Connect failed:", err2)
			return err2
		}
		fmt.Printf("Service '%s' started\n", info.Name)
	}
	// last, show a list of running services
	err = HandleListServices(ctx, f)
	return err
}

// HandleStopService stops a service
func HandleStopService(ctx context.Context, f svcconfig.AppFolders, serviceName string) error {
	var ls launcher.ILauncher

	conn, err := hubclient.CreateLocalClientConnection(launcher.ServiceName, f.Run)
	if err == nil {
		ls, err = capnpclient.NewLauncherCapnpClient(ctx, conn)
	}
	if err != nil {
		return err
	}

	if serviceName == "all" {
		err := ls.StopAll(ctx)

		if err != nil {
			fmt.Println("Stop all failed:", err)
			return err
		}
		fmt.Printf("All services stopped\n")

	} else {
		info, err := ls.StopService(ctx, serviceName)
		if err != nil {
			fmt.Printf("Stop %s failed: %s\n", serviceName, err)
			return err
		}
		fmt.Printf("Service '%s' stopped\n", info.Name)
	}
	// last, show a list of running services
	err = HandleListServices(ctx, f)
	return err
}
