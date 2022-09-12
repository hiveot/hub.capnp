package provcli

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// FIXME: centralize service name, until protobuf supports constants... (hint hint)
const provisioningAppID = "oobprov"

// GetProvCommands returns the provisioning handling commands
// This requires the provisioning service to run.
func GetProvCommands(homeFolder string) *cli.Command {

	cmd := &cli.Command{
		//hub prov add|list  <deviceID> <secret>

		Name:  "prov",
		Usage: "IoT device provisioning",
		Subcommands: []*cli.Command{
			GetProvAddCommand(),
		},
	}

	return cmd
}

// GetProvAddCommand
// prov add [--secrets=folder] <deviceID> <oobsecret>
func GetProvAddCommand() *cli.Command {
	provServiceAddress := "localhost:8881"
	return &cli.Command{
		Name:      "add",
		Usage:     "Add an out-of-band device provisioning secret for automatic provisioning",
		ArgsUsage: "<deviceID> <oobSecret>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "address",
				Usage:       "Provisioning service address",
				Value:       provServiceAddress,
				Destination: &provServiceAddress,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 2 {
				return fmt.Errorf("expected 2 arguments. Got %d instead", cCtx.NArg())
			}
			err := HandleAddOobSecret(
				provServiceAddress,
				cCtx.Args().Get(0),
				cCtx.Args().Get(1))
			fmt.Println("Adding secret for device: ", cCtx.Args().First())
			return err
		},
	}
}

// HandleAddOobSecret invokes the out-of-band provisioning service to add a provisioning secret
//  address is the destination service's address
//  deviceID is the ID of the device whose secret to set
//  secret to set
func HandleAddOobSecret(address string, deviceID string, secret string) error {
	//// Set up a connection to the server.
	//daprClient, err := dapr.NewClient()
	//if err != nil {
	//	err2 := fmt.Errorf("Error initialize dapr client. Make sure thsi runs with a sidecart:%s", err)
	//	log.Println(err2)
	//	return err
	//}
	//defer daprClient.Close()
	cred := insecure.NewCredentials()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	conn, err := Dial(ctx, address,
		grpc.WithTransportCredentials(cred),
		grpc.WithBlock(),
	)
	if err != nil {
		logrus.Errorf("failed to connect: %v", err)
		return err
	}
	defer conn.Close()
	cl := svc.NewProvisioningClient(conn)

	// The service name tells the sidecar what service to connect to
	ctx = metadata.AppendToOutgoingContext(ctx, "dapr-app-id", appID)
	defer cancel()

	args := &svc.AddOobSecrets_Args{
		Secrets: make([]*svc.AddOobSecrets_Args_OobSecret, 0),
	}
	_, err = cl.AddOobSecrets(ctx, args)
	return err
}
