package directorycli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpclient"
)

func DirectoryListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	var limit = 100
	var offset = 0
	return &cli.Command{
		Name:      "listdir",
		Aliases:   []string{"lid"},
		Category:  "directory",
		Usage:     "List directory",
		UsageText: "List all Things in the directory",
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

	conn, err := hubclient.ConnectToService(directory.ServiceName, f.Run)
	if err == nil {
		dir = capnpclient.NewDirectoryCapnpClient(ctx, conn)
		rd, err = dir.CapReadDirectory(ctx, "hubcli")
	}
	if err != nil {
		return err
	}

	cursor := rd.Cursor(ctx)
	fmt.Println("PublisherID    Thing ID              Modified                       type            nr props   events  actions")
	fmt.Println("-----------    ---------------       --------                       ----            --------   ------  -------")
	i := 0
	tv, valid := cursor.First()
	if offset > 0 {
		// TODO, skip
		//tv, valid = cursor.Skip(offset)
	}
	for ; valid && i < limit; tv, valid = cursor.Next() {
		err = json.Unmarshal(tv.ValueJSON, &tdDoc)
		var utime time.Time
		if tdDoc.Modified != "" {
			utime, err = dateparse.ParseAny(tdDoc.Modified)
		} else if tdDoc.Created != "" {
			utime, err = dateparse.ParseAny(tdDoc.Created)
		}
		timeStr := utime.In(time.Local).Format("02 Jan 2006 15:04:05 -0700")
		//if err != nil {
		//	logrus.Infof("Parsing time failed '%s': %s", tdDoc.Modified, err)
		//}

		fmt.Printf("%-15s %-20s %-30s %-15s %8d %8d %8d\n",
			tv.PublisherID,
			tdDoc.ID,
			timeStr,
			tdDoc.DeviceType,
			len(tdDoc.Properties),
			len(tdDoc.Events),
			len(tdDoc.Actions),
		)
	}
	return nil
}
