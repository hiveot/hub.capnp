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
	var limit = 100
	var offset = 0
	return &cli.Command{
		Name:      "list",
		Usage:     "List events",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListEvents(ctx, f, limit, offset)
			return err
		},
	}
}

// HandleListEvents lists the history content
func HandleListEvents(ctx context.Context, f svcconfig.AppFolders, limit int, offset int) error {
	var hist history.IHistory
	var rd history.IReadHistory

	conn, err := listener.CreateClientConnection(f.Run, history.ServiceName)
	if err == nil {
		hist, err = capnpclient.NewHistoryCapnpClient(ctx, conn)
	}
	if err == nil {
		rd = hist.CapReadHistory()
	}
	if err != nil {
		return err
	}
	thingID := ""
	eventName := ""
	after := ""
	before := ""
	events, _ := rd.GetEventHistory(ctx, thingID, eventName, after, before, limit)
	fmt.Println("Thing ID                            Timestamp                      Event      Value (truncated)")
	fmt.Println("--------                            -------                        ----       ---------------- ")
	for _, event := range events {
		utime, err := dateparse.ParseAny(event.Created)

		if err != nil {
			logrus.Infof("Parsing time failed '%s': %s", event.Created, err)
		}

		fmt.Printf("%-35s %-30s %-10s %-30s\n",
			event.ThingID,
			utime.Format("02 Jan 2006 15:04:05 -0700"),
			event.Name,
			event.ValueJSON,
		)
	}
	return nil
}
