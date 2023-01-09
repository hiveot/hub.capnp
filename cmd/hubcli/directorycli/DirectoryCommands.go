package directorycli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/araddon/dateparse"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/lib/thing"

	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
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

// HandleListDirectory lists the directory content
func HandleListDirectory(ctx context.Context, f svcconfig.AppFolders, limit int, offset int) error {
	var dir directory.IDirectory
	var rd directory.IReadDirectory
	var tdDoc thing.TD

	conn, err := listener.CreateLocalClientConnection(directory.ServiceName, f.Run)
	if err == nil {
		dir = capnpclient.NewDirectoryCapnpClient(ctx, conn)
		rd, err = dir.CapReadDirectory(ctx, "hubcli")
	}
	if err != nil {
		return err
	}

	cursor := rd.Cursor(ctx)
	fmt.Println("PublisherID    Thing ID              Modified                       type       props  events  actions")
	fmt.Println("-----------    ---------------       -------                        ----       -----  ------  -------")
	i := 0
	tv, valid := cursor.First()
	if offset > 0 {
		// TODO, skip
		//tv, valid = cursor.Skip(offset)
	}
	for ; valid && i < limit; tv, valid = cursor.Next() {
		err = json.Unmarshal(tv.ValueJSON, &tdDoc)

		utime, err := dateparse.ParseAny(tdDoc.Modified)
		if err != nil {
			logrus.Infof("Parsing time failed '%s': %s", tdDoc.Modified, err)
		}

		fmt.Printf("%-15s %-20s %-30s %-10s %5d\n",
			tv.PublisherID,
			tdDoc.ID,
			//tdDoc.Modified,
			utime.Format("02 Jan 2006 15:04:05 -0700"),
			tdDoc.DeviceType,
			len(tdDoc.Properties),
		)
	}
	return nil
}
