package pubsubcli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub.capnp/go/vocab"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/capnpclient"
)

// PubSubCommands returns the publish/subscribe service commands
// This requires the pubsub service to run.
func PubSubCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {

	cmd := &cli.Command{
		//hub prov add|list  <deviceID> <secret>

		Name:  "pubsub",
		Usage: "Publish/Subscribe",
		Subcommands: cli.Commands{
			SubTDCommand(ctx, f),
			SubEventsCommand(ctx, f),
			//PubActionCommand(ctx, f),
		},
	}
	return cmd
}

func SubTDCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "subtd",
		Usage:     "Subscribe to TD publications. Use Ctrl-C to stop watching.",
		ArgsUsage: "(no arguments)",
		Action: func(cCtx *cli.Context) error {
			err := HandleSubTD(ctx, f)
			return err
		},
	}
}

func SubEventsCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "subevents",
		Usage:     "Subscribe to Thing events. Use Ctrl-C to stop watching.",
		ArgsUsage: "(no arguments)",
		Action: func(cCtx *cli.Context) error {
			err := HandleSubEvents(ctx, f)
			return err
		},
	}
}

// HandleSubTD subscribes and prints TD publications
func HandleSubTD(ctx context.Context, f svcconfig.AppFolders) error {
	var pubSubSvc pubsub.IPubSubService

	conn, err := hubclient.CreateLocalClientConnection(pubsub.ServiceName, f.Run)
	if err == nil {
		pubSubSvc = capnpclient.NewPubSubCapnpClient(ctx, conn)
	}
	if err != nil {
		return err
	}
	pubSubUser, _ := pubSubSvc.CapUserPubSub(ctx, "hubcli")
	err = pubSubUser.SubTDs(ctx, func(event *thing.ThingValue) {
		var td thing.TD
		err = json.Unmarshal(event.ValueJSON, &td)
		valid := err == nil
		//createdTime, _ := dateparse.ParseAny(event.Created)
		createdTime, _ := dateparse.ParseAny(td.Modified)
		timeStr := createdTime.Format("15:04:05.000")
		fmt.Printf("%-16s %-20s %-25s %-15s %v \n",
			timeStr, event.PublisherID, event.ThingID, td.DeviceType, valid)
	})
	fmt.Printf("Created          Publisher            ThingID                   Type            Valid TD  \n")
	fmt.Printf("---------------  -------------------  ------------------------  --------------  ----\n")

	if err != nil {
		return err
	}
	time.Sleep(time.Hour * 24)
	return nil
}

// HandleSubEvents subscribes and prints value and property events
func HandleSubEvents(ctx context.Context, f svcconfig.AppFolders) error {
	var pubSubSvc pubsub.IPubSubService

	conn, err := hubclient.CreateLocalClientConnection(pubsub.ServiceName, f.Run)
	if err == nil {
		pubSubSvc = capnpclient.NewPubSubCapnpClient(ctx, conn)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Time             Publisher            ThingID                   Name                 Value\n")
	fmt.Printf("---------------  -------------------  ------------------------  -------------------  ---------\n")

	pubSubUser, _ := pubSubSvc.CapUserPubSub(ctx, "hubcli")
	err = pubSubUser.SubEvent(ctx, "", "", "", func(event *thing.ThingValue) {
		createdTime, _ := dateparse.ParseAny(event.Created)
		timeStr := createdTime.Format("15:04:05.000")
		value := fmt.Sprintf("%-.30s", event.ValueJSON)
		if event.Name == vocab.WoTProperties {
			var props map[string][]byte
			json.Unmarshal(event.ValueJSON, &props)
			value = fmt.Sprintf("(%d): %s", len(props), props)
		}

		fmt.Printf("%-16s %-20s %-25s %-20s %-30s\n",
			timeStr, event.PublisherID, event.ThingID, event.Name, value)
	})
	if err != nil {
		return err
	}
	time.Sleep(time.Hour * 24)
	return nil
}
