package directorycli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpclient"
)

func DirectoryCommands(ctx context.Context, f folders.AppFolders) *cli.Command {
	cmd := &cli.Command{
		Name:  "dir",
		Usage: "List and query directory content",
		Subcommands: []*cli.Command{
			DirectoryListCommand(ctx, f),
		},
	}
	return cmd
}

// DirectoryListCommand
func DirectoryListCommand(ctx context.Context, f folders.AppFolders) *cli.Command {
	var limit = 100
	var offset = 0
	return &cli.Command{
		Name:      "list",
		Usage:     "List services",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListDirectory(ctx, f, limit, offset)
			return err
		},
	}
}

// HandleListDirectory lists the directoryc content
func HandleListDirectory(ctx context.Context, f folders.AppFolders, limit int, offset int) error {
	var dir directory.IDirectory
	var rd directory.IReadDirectory
	var tdDoc thing.ThingDescription

	conn, err := listener.CreateClientConnection(f.Run, directory.ServiceName)
	if err == nil {
		dir, err = capnpclient.NewDirectoryCapnpClient(ctx, conn)
	}
	if err == nil {
		rd = dir.CapReadDirectory()
	}
	if err != nil {
		return err
	}

	jsonEntries, _ := rd.ListTDs(ctx, limit, offset)
	fmt.Println("Thing ID              Updated         type      props  events  actions")
	fmt.Println("--------              -------------   ----      -----  ------  -------")
	for _, entry := range jsonEntries {
		err = json.Unmarshal([]byte(entry), &tdDoc)
		fmt.Printf("%-25s %20s   %10s   \n",
			tdDoc.ID,
			tdDoc.Modified,
		)
	}
	return nil
}
