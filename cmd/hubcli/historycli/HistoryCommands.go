package historycli

import (
	"context"
	"fmt"

	"github.com/araddon/dateparse"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpclient"
)

func HistoryCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	cmd := &cli.Command{
		Name:  "history",
		Usage: "List and query Thing events",
		Subcommands: []*cli.Command{
			HistoryInfoCommand(ctx, f),
			HistoryListCommand(ctx, f),
			HistoryLatestCommand(ctx, f),
			HistoryRetainCommand(ctx, f),
		},
	}
	return cmd
}

func HistoryInfoCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "info",
		Usage:     "Show history store info",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleHistoryInfo(ctx, f)
			return err
		},
	}
}

func HistoryListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List recent events",
		ArgsUsage: "publisherID thingID",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 2 {
				return fmt.Errorf("publisherID and thingID expected")
			}
			err := HandleListEvents(ctx, f, cCtx.Args().First(), cCtx.Args().Get(1), 100)
			return err
		},
	}
}

func HistoryLatestCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "latest",
		Usage:     "List latest event/property values",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListLatestEvents(ctx, f)
			return err
		},
	}
}
func HistoryRetainCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "retain",
		Usage:     "List retained events",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListRetainedEvents(ctx, f)
			return err
		},
	}
}

func HandleHistoryInfo(ctx context.Context, f svcconfig.AppFolders) error {
	var hist history.IHistoryService
	var rd history.IReadHistory

	conn, err := hubclient.CreateLocalClientConnection(history.ServiceName, f.Run)
	if err == nil {
		hist = capnpclient.NewHistoryCapnpClient(ctx, conn)
		rd, err = hist.CapReadHistory(ctx, "hubcli", "", "")
	}
	if err != nil {
		return err
	}
	info := rd.Info(ctx)

	fmt.Println(fmt.Sprintf("ID:          %s", info.Id))
	fmt.Println(fmt.Sprintf("Size:        %d", info.DataSize))
	fmt.Println(fmt.Sprintf("Nr Records   %d", info.NrRecords))
	fmt.Println(fmt.Sprintf("Engine       %s", info.Engine))

	rd.Release()
	return conn.Close()
}

// HandleListEvents lists the history content
func HandleListEvents(ctx context.Context, f svcconfig.AppFolders, publisherID, thingID string, limit int) error {
	var hist history.IHistoryService
	var rd history.IReadHistory

	conn, err := hubclient.CreateLocalClientConnection(history.ServiceName, f.Run)
	if err == nil {
		hist = capnpclient.NewHistoryCapnpClient(ctx, conn)
		rd, err = hist.CapReadHistory(ctx, "hubcli", publisherID, thingID)
	}
	if err != nil {
		return err
	}
	eventName := ""
	cursor := rd.GetEventHistory(ctx, eventName)
	fmt.Println("PublisherID    thingID           Timestamp                    Event           Value (truncated)")
	fmt.Println("-----------    -------           ---------                    -----           ---------------- ")
	count := 0
	for tv, valid := cursor.Last(); valid && count < limit; tv, valid = cursor.Prev() {
		count++
		utime, err := dateparse.ParseAny(tv.Created)

		if err != nil {
			logrus.Infof("Parsing time failed '%s': %s", tv.Created, err)
		}

		fmt.Printf("%-13s %-18s %-28s %-15s %-30s\n",
			tv.PublisherID,
			tv.ThingID,
			utime.Format("02 Jan 2006 15:04:05 -0700"),
			tv.Name,
			tv.ValueJSON,
		)
	}
	rd.Release()
	conn.Close()
	return nil
}

// HandleListRetainedEvents lists the events that are retained
func HandleListRetainedEvents(ctx context.Context, f svcconfig.AppFolders) error {
	var hist history.IHistoryService
	var mngRet history.IManageRetention

	conn, err := hubclient.CreateLocalClientConnection(history.ServiceName, f.Run)
	if err == nil {
		hist = capnpclient.NewHistoryCapnpClient(ctx, conn)
		mngRet, err = hist.CapManageRetention(ctx, "hubcli")
	}
	if err != nil {
		return err
	}
	evList, _ := mngRet.GetEvents(ctx)

	fmt.Println("Event Name      days     publishers          Things                         Excluded")
	fmt.Println("----------      ----     ----------          ------                         -------- ")
	for _, evRet := range evList {

		fmt.Printf("%-15s %-8d %-30s %-30s %-30s,\n",
			evRet.Name,
			evRet.RetentionDays,
			evRet.Publishers,
			evRet.Things,
			evRet.Exclude,
		)
	}
	mngRet.Release()
	conn.Close()
	return nil
}

func HandleListLatestEvents(ctx context.Context, f svcconfig.AppFolders) error {
	var hist history.IHistoryService
	var readHist history.IReadHistory

	conn, err := hubclient.CreateLocalClientConnection(history.ServiceName, f.Run)
	if err == nil {
		hist = capnpclient.NewHistoryCapnpClient(ctx, conn)
		readHist, err = hist.CapReadHistory(ctx, "hubcli", "", "")
	}
	if err != nil {
		return err
	}
	props := readHist.GetProperties(ctx, nil)

	fmt.Println("Event Name      Publisher       Thing           Created         Value")
	fmt.Println("----------      ---------       -----           -------         -----")
	for _, prop := range props {

		fmt.Printf("%-15s %-15s %-15s %-15s %s\n",
			prop.Name,
			prop.PublisherID,
			prop.ThingID,
			prop.Created,
			prop.ValueJSON,
		)
	}
	readHist.Release()
	conn.Close()
	return nil
}
