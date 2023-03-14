package authzcli

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capnpclient"
)

// AuthzListGroupsCommand lists the groups a client is a member off
// hubcli authz groups [clientID]
func AuthzListGroupsCommand(ctx context.Context, runFolder *string) *cli.Command {
	clientID := ""
	return &cli.Command{
		Name:      "lgr",
		Usage:     "List groups the user is a member of",
		ArgsUsage: "[<loginID>]",
		Category:  "authorization",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() > 0 {
				clientID = cCtx.Args().Get(0)
			} else if cCtx.NArg() > 1 {
				err := fmt.Errorf("multiple arguments, expected only a single clientID")
				return err
			}
			err := HandleListGroups(ctx, *runFolder, clientID)
			return err
		},
	}
}

// HandleListGroups shows a list of groups the client is a member of
func HandleListGroups(ctx context.Context, runFolder, clientID string) error {
	var err error
	var authzClient authz.IAuthz
	var manageAuthz authz.IManageAuthz

	conn, err := hubclient.ConnectToService(authz.ServiceName, runFolder)
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

	fmt.Println("Group ID                          role")
	fmt.Println("----------                          role")
	for groupName, role := range groupRoles {

		fmt.Printf("%-35s %10s\n",
			groupName,
			role,
		)
	}
	return err
}
