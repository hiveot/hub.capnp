package directorycli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/araddon/dateparse"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpclient"
)

func DirectoryCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
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
func DirectoryListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
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
func HandleListDirectory(ctx context.Context, f svcconfig.AppFolders, limit int, offset int) error {
	var dir directory.IDirectory
	var rd directory.IReadDirectory
	var tdDoc thing.ThingDescription

	conn, err := listener.CreateLocalClientConnection(directory.ServiceName, f.Run)
	if err == nil {
		dir, err = capnpclient.NewDirectoryCapnpClient(ctx, conn)
	}
	if err == nil {
		rd = dir.CapReadDirectory(ctx)
	}
	if err != nil {
		return err
	}

	cursor := rd.Cursor(ctx)
	fmt.Println("Thing ID                            Modified                       type       props  events  actions")
	fmt.Println("--------                            -------                        ----       -----  ------  -------")
	for tv, valid := cursor.First(); valid; tv, valid = cursor.Next() {
		err = json.Unmarshal(tv.ValueJSON, &tdDoc)

		utime, err := dateparse.ParseAny(tdDoc.Modified)
		if err != nil {
			logrus.Infof("Parsing time failed '%s': %s", tdDoc.Modified, err)
		}

		fmt.Printf("%-35s %-30s %-10s %5d\n",
			tdDoc.ID,
			//tdDoc.Modified,
			utime.Format("02 Jan 2006 15:04:05 -0700"),
			tdDoc.AtType,
			len(tdDoc.Properties),
		)
	}
	return nil
}
