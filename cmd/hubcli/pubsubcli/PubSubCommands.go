package pubsubcli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/api/go/vocab"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/capnpclient"
)

// SubTDCommand shows TD publications
func SubTDCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "subtd",
		Aliases:   []string{"std"},
		Usage:     "Show TD publications",
		UsageText: "Live show TD publications from the pubsub message bus. Use Ctrl-C to stop watching.",
		Category:  "pubsub",
		Action: func(cCtx *cli.Context) error {
			err := HandleSubTD(ctx, f)
			return err
		},
	}
}

func SubEventsCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "subevents",
		Aliases:   []string{"sev"},
		Usage:     "Show Thing events",
		UsageText: "Live show Thing event publications from the pubsub message bus. Use Ctrl-C to stop watching.",
		Category:  "pubsub",
		Action: func(cCtx *cli.Context) error {
			err := HandleSubEvents(ctx, f)
			return err
		},
	}
}

// HandleSubTD subscribes and prints TD publications
func HandleSubTD(ctx context.Context, f svcconfig.AppFolders) error {
	var pubSubSvc pubsub.IPubSubService

	conn, err := hubclient.ConnectToService(pubsub.ServiceName, f.Run)
	if err == nil {
		pubSubSvc = capnpclient.NewPubSubCapnpClient(ctx, conn)
	}
	if err != nil {
		return err
	}
	pubSubUser, _ := pubSubSvc.CapUserPubSub(ctx, "hubcli")
	err = pubSubUser.SubTDs(ctx, func(event *thing.ThingValue) {
		var td thing.TD
		//fmt.Printf("%s\n", event.ValueJSON)
		err = json.Unmarshal(event.ValueJSON, &td)

		modifiedTime, _ := dateparse.ParseAny(td.Modified)                  // can be in any TZ
		timeStr := modifiedTime.In(time.Local).Format("15:04:05.000 -0700") // want local time
		fmt.Printf("%-20s %-25s %-20s %-20s %-18s\n",
			event.PublisherID, event.ThingID, td.Title, td.DeviceType, timeStr)
	})
	fmt.Printf("Publisher ID         Thing ID                  Title                Type                 Modified          \n")
	fmt.Printf("-------------------  ------------------------  -------------------  -------------------  ------------------\n")

	if err != nil {
		return err
	}
	time.Sleep(time.Hour * 24)
	return nil
}

// HandleSubEvents subscribes and prints value and property events
func HandleSubEvents(ctx context.Context, f svcconfig.AppFolders) error {
	var pubSubSvc pubsub.IPubSubService

	conn, err := hubclient.ConnectToService(pubsub.ServiceName, f.Run)
	if err == nil {
		pubSubSvc = capnpclient.NewPubSubCapnpClient(ctx, conn)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Time             Publisher            ThingID                   PropID                         Value\n")
	fmt.Printf("---------------  -------------------  ------------------------  -----------------------------  ---------\n")

	pubSubUser, _ := pubSubSvc.CapUserPubSub(ctx, "hubcli")
	err = pubSubUser.SubEvent(ctx, "", "", "", func(event *thing.ThingValue) {
		createdTime, _ := dateparse.ParseAny(event.Created)
		timeStr := createdTime.Format("15:04:05.000")
		value := fmt.Sprintf("%-.30s", event.ValueJSON)
		if event.ID == vocab.WoTProperties {
			var props map[string][]byte
			json.Unmarshal(event.ValueJSON, &props)
			value = fmt.Sprintf("(%d): %s", len(props), props)
		}

		fmt.Printf("%-16s %-20s %-25s %-30s %-30s\n",
			timeStr, event.PublisherID, event.ThingID, event.ID, value)
	})
	if err != nil {
		return err
	}
	time.Sleep(time.Hour * 24)
	return nil
}
