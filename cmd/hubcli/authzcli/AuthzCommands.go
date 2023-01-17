package authzcli

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capnpclient"
)

// AuthzListGroupsCommand lists the groups a client is a member off
// hubcli authz groups [clientID]
func AuthzListGroupsCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	clientID := ""
	return &cli.Command{
		Name:      "listgroups [clientID]",
		Aliases:   []string{"lig"},
		Usage:     "List groups",
		UsageText: "List groups the given client is a member of, or all groups.",
		Category:  "authorization",
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

	conn, err := hubclient.CreateLocalClientConnection(authz.ServiceName, f.Run)
	if err == nil {
		authzClient = capnpclient.NewAuthzCapnpClient(ctx, conn)
	}
	if err == nil {
		manageAuthz, _ = authzClient.CapManageAuthz(ctx, "hubcli")
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
