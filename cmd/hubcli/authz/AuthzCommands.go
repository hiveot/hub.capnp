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

// AuthzCommands returns the list of Authorization commands
func AuthzCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	cmd := &cli.Command{
		Name:      "authz",
		Usage:     "Manage authorization",
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

	conn, err := listener.CreateLocalClientConnection(authz.ServiceName, f.Run)
	if err == nil {
		authzClient, err = capnpclient.NewAuthzCapnpClient(ctx, conn)
	}
	if err == nil {
		manageAuthz = authzClient.CapManageAuthz(ctx)
	}
	if err != nil {
		return err
	}
	groupRoles, err := manageAuthz.GetGroupRoles(ctx, clientID)

	fmt.Println("Group Name                          role")
	fmt.Println("----------                          role")
	for groupName, role := range groupRoles {

		fmt.Printf("%-35s %10s\n",
			groupName,
			role,
		)
	}
	return err
}
