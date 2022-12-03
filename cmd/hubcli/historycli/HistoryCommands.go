package historycli

import (
	"context"
	"fmt"

	"github.com/araddon/dateparse"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpclient"
)

func HistoryCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	cmd := &cli.Command{
		Name:  "hist",
		Usage: "List and query Thing events",
		Subcommands: []*cli.Command{
			HistoryListCommand(ctx, f),
		},
	}
	return cmd
}

// HistoryListCommand
func HistoryListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List recent events",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListEvents(ctx, f, "", 100)
			return err
		},
	}
}

// HandleListEvents lists the history content
func HandleListEvents(ctx context.Context, f svcconfig.AppFolders, thingAddr string, limit int) error {
	var hist history.IHistoryService
	var rd history.IReadHistory

	conn, err := listener.CreateClientConnection(f.Run, history.ServiceName)
	if err == nil {
		hist, err = capnpclient.NewHistoryCapnpClient(ctx, conn)
	}
	if err == nil {
		rd = hist.CapReadHistory(ctx, thingAddr)
	}
	if err != nil {
		return err
	}
	eventName := ""
	cursor := rd.GetEventHistory(ctx, eventName)
	fmt.Println("Thing ID                            Timestamp                      Event      Value (truncated)")
	fmt.Println("--------                            -------                        ----       ---------------- ")
	count := 0
	for tv, valid := cursor.Last(); valid && count < limit; tv, valid = cursor.Prev() {
		count++
		utime, err := dateparse.ParseAny(tv.Created)

		if err != nil {
			logrus.Infof("Parsing time failed '%s': %s", tv.Created, err)
		}

		fmt.Printf("%-35s %-30s %-10s %-30s\n",
			tv.ThingAddr,
			utime.Format("02 Jan 2006 15:04:05 -0700"),
			tv.Name,
			tv.ValueJSON,
		)
	}
	return nil
}
