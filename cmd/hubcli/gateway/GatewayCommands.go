package gatewaycli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
)

func GatewayCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	cmd := &cli.Command{
		Name:  "gw",
		Usage: "List gateway capabilities",
		Subcommands: []*cli.Command{
			GatewayListCommand(ctx, f),
		},
	}
	return cmd
}

func GatewayListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List gateway capabilities",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListGateway(ctx, f)
			return err
		},
	}
}

func HandleListGateway(ctx context.Context, f svcconfig.AppFolders) error {
	var gw *capnpclient.GatewayServiceCapnpClient //gateway.IGatewayService
	//logrus.Infof("Contacting gateway on '%s'", f.Run)
	ctx2, _ := context.WithTimeout(ctx, time.Second*10)
	conn, err := listener.CreateClientConnection(f.Run, gateway.ServiceName)
	if err == nil {
		//logrus.Infof("Connection established")
		gw, err = capnpclient.NewGatewayServiceCapnpClient(ctx2, conn)
		defer gw.Release()
	}
	if err != nil {
		return err
	}
	//logrus.Infof("Sending request")

	// ask as a service. we might want to make this a parameter
	capList, err := gw.ListCapabilities(ctx2, hubapi.ClientTypeService)
	fmt.Println("Capability                          Service                        ClientTypes")
	fmt.Println("--------                            -------                        ----       ")
	for _, capInfo := range capList {
		clientTypeAsText := strings.Join(capInfo.ClientTypes, ",")
		fmt.Printf("%-35s %-30s %-30s\n",
			capInfo.CapabilityName,
			capInfo.ServiceName,
			clientTypeAsText,
		)
	}
	return nil
}
