package authz

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capnpclient"
)

// AuthzCommands returns the list of Authentication service commands
func AuthzCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	cmd := &cli.Command{
		Name:      "authz",
		Usage:     "Manage authentication",
		ArgsUsage: ";", // no args
		Subcommands: []*cli.Command{
			AuthzListGroupsCommand(ctx, f),
			//AuthzSetClientRoleCommand(ctx, f),
			//AuthzRemoveClientCommand(ctx, f),
		},
	}
	return cmd
}

// AuthzListGroupsCommand lists the groups a client is a member off
// hubcli authz groups [clientID]
func AuthzListGroupsCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	clientID := ""
	return &cli.Command{
		Name:      "groups",
		Usage:     "List groups",
		ArgsUsage: "[clientID]",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() > 0 {
				clientID = cCtx.Args().Get(0)
			} else if cCtx.NArg() > 1 {
				err := fmt.Errorf("multiple arguments, expected only a single clientID")
				return err
			}
			err := HandleListGroups(ctx, f, clientID)
			return err
		},
	}
}

// HandleListGroups shows a list of groups the client is a member of
func HandleListGroups(ctx context.Context, f svcconfig.AppFolders, clientID string) error {
	var err error
	var authzClient authz.IAuthz
	var manageAuthz authz.IManageAuthz

	conn, err := listener.CreateClientConnection(f.Run, authz.ServiceName)
	if err == nil {
		authzClient, err = capnpclient.NewAutzCapnpClient(ctx, conn)
	}
	if err == nil {
		manageAuthz = authzClient.CapManageAuthorization(ctx)
	}
	if err != nil {
		return err
	}
	groups := manageAuthz.GetGroups(ctx, clientID)

	fmt.Println("Group Name                          role")
	fmt.Println("----------                          role")
	for _, entry := range groups {
		name := entry

		fmt.Printf("%-35s %10s\n",
			name,
			"tbd",
		)
	}
	return nil
}
