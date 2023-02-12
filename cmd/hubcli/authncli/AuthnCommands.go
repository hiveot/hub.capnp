package authncli

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capnpclient"
)

// AuthnAddUserCommand adds a user
func AuthnAddUserCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "adduser <loginID>", // loginID is ignored in the command
		Usage:     "Add a user",
		Aliases:   []string{"addu"},
		UsageText: "Add a Hub user and generate a temporary password.",
		Category:  "authentication",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 1 {
				err := fmt.Errorf("expected 1 argument")
				return err
			}
			loginID := cCtx.Args().Get(0)
			err := HandleAddUser(ctx, f, loginID)
			return err
		},
	}
}

// AuthnListUsersCommand lists user profiles
func AuthnListUsersCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "listusers",
		Aliases:   []string{"liu"},
		Usage:     "List users",
		UsageText: "List all registered Hub users",
		Category:  "authentication",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() > 0 {
				err := fmt.Errorf("too many arguments")
				return err
			}
			err := HandleListUsers(ctx, f)
			return err
		},
	}
}

// AuthnRemoveUserCommand removes a user
func AuthnRemoveUserCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "removeuser <loginID>",
		Aliases:   []string{"remu"},
		Usage:     "Remove a user",
		UsageText: "Remove a user from the Hub. Use with care. This does not ask for confirmation",
		Category:  "authentication",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 1 {
				err := fmt.Errorf("expected 1 arguments")
				return err
			}
			loginID := cCtx.Args().Get(0)
			err := HandleRemoveUser(ctx, f, loginID)
			return err
		},
	}
}

// HandleAddUser adds a user
func HandleAddUser(ctx context.Context, f svcconfig.AppFolders, loginID string) error {
	var err error
	var authnClient authn.IAuthnService
	var manageAuthn authn.IManageAuthn

	conn, err := hubclient.ConnectToService(authn.ServiceName, f.Run)
	if err == nil {
		authnClient = capnpclient.NewAuthnClientFromCapnpConnection(ctx, conn)
	}
	if err == nil {
		manageAuthn, _ = authnClient.CapManageAuthn(ctx, "hubcli")
	}
	if err != nil {
		return err
	}
	newPassword, err := manageAuthn.AddUser(ctx, loginID)

	if err != nil {
		fmt.Println("Error: " + err.Error())
	} else {
		fmt.Println("User " + loginID + " added successfully. Temp password: " + newPassword)

	}
	return err
}

// HandleListUsers shows a list of user profiles
func HandleListUsers(ctx context.Context, f svcconfig.AppFolders) error {
	var err error
	var authnClient authn.IAuthnService
	var manageAuthn authn.IManageAuthn

	conn, err := hubclient.ConnectToService(authn.ServiceName, f.Run)
	if err == nil {
		authnClient = capnpclient.NewAuthnClientFromCapnpConnection(ctx, conn)
	}
	if err == nil {
		manageAuthn, _ = authnClient.CapManageAuthn(ctx, "hubcli")
	}
	if err != nil {
		return err
	}
	profileList, err := manageAuthn.ListUsers(ctx)

	fmt.Println("Login ID                             User Name")
	fmt.Println("--------                             ---------")
	for _, profile := range profileList {

		fmt.Printf("%-35s  %-10s\n",
			profile.LoginID,
			profile.Name,
		)
	}
	return err
}

// HandleRemoveUser removes a user
func HandleRemoveUser(ctx context.Context, f svcconfig.AppFolders, loginID string) error {
	var err error
	var authnClient authn.IAuthnService
	var manageAuthn authn.IManageAuthn

	conn, err := hubclient.ConnectToService(authn.ServiceName, f.Run)
	if err == nil {
		authnClient = capnpclient.NewAuthnClientFromCapnpConnection(ctx, conn)
	}
	if err == nil {
		manageAuthn, _ = authnClient.CapManageAuthn(ctx, "hubcli")
	}
	if err != nil {
		return err
	}
	// TODO: that the user's data should also be removed
	err = manageAuthn.RemoveUser(ctx, loginID)

	if err != nil {
		fmt.Println("Error: " + err.Error())
	} else {
		fmt.Println("User " + loginID + " removed")

	}
	return err
}
