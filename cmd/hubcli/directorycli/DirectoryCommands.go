package directorycli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hiveot/hub/lib/utils"
	"time"

	"github.com/araddon/dateparse"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpclient"
)

const Reset = "\033[0m"
const Red = "\033[31m"
const Green = "\033[32m"
const Yellow = "\033[33m"
const Blue = "\033[34m"
const Purple = "\033[35m"
const Cyan = "\033[36m"
const Gray = "\033[37m"
const White = "\033[97m"

func DirectoryListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	var limit = 100
	var offset = 0
	var verbose = false
	return &cli.Command{
		Name:      "listdir [<publisherID> <thingID> [-v]]",
		Aliases:   []string{"ld"},
		Category:  "directory",
		Usage:     "List directory",
		UsageText: "List all Things or a selected Thing in the directory",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "v",
				Usage:       "Verbose, display raw json",
				Value:       false,
				Destination: &verbose,
			},
		}, Action: func(cCtx *cli.Context) error {
			var err error = fmt.Errorf("expected 0 or 2 parameters")
			if cCtx.NArg() == 0 {
				err = HandleListDirectory(ctx, f, limit, offset)
			} else if cCtx.NArg() == 2 {
				if !verbose {
					err = HandleListThing(ctx, f, cCtx.Args().First(), cCtx.Args().Get(1))
				} else {
					err = HandleListThingVerbose(ctx, f, cCtx.Args().First(), cCtx.Args().Get(1))
				}
			}
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
	fmt.Println("Publisher ID    Thing ID             Title                          Type                   #props  #events #actions  Modified         ")
	fmt.Println("-------------   -------------------  -----------------------------  --------------------   ------  ------- --------  --------------------------")
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

		fmt.Printf("%-15s %-20s %-30s %-20s %7d  %7d  %7d   %-30s\n",
			tv.PublisherID,
			tdDoc.ID,
			tdDoc.Title,
			tdDoc.AtType,
			len(tdDoc.Properties),
			len(tdDoc.Events),
			len(tdDoc.Actions),
			timeStr,
		)
	}
	return nil
}

// HandleListThing lists a Thing in the directory
func HandleListThing(ctx context.Context, f svcconfig.AppFolders, pubID, thingID string) error {
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
	tv, err := rd.GetTD(ctx, pubID, thingID)
	if err != nil {
		return err
	}
	err = json.Unmarshal(tv.ValueJSON, &tdDoc)
	if err != nil {
		return err
	}
	fmt.Printf("%sTD of %s %s:%s\n", Blue, pubID, thingID, Reset)
	fmt.Printf(" title:       %s\n", tdDoc.Title)
	fmt.Printf(" description: %s\n", tdDoc.Description)
	fmt.Printf(" deviceType:  %s\n", tdDoc.AtType)
	fmt.Printf(" modified:    %s\n", tdDoc.Modified)
	fmt.Println("")

	fmt.Println("Properties:")
	fmt.Println(" ID                             Title                                    DataType   Default    Initial Value   ReadOnly   WriteOnly  Description")
	fmt.Println(" -----------------------------  ---------------------------------------  ---------  ---------  --------------  --------   ---------  -----------")

	keys := utils.OrderedMapKeys(tdDoc.Properties)
	for _, key := range keys {
		prop := tdDoc.Properties[key]
		fmt.Printf(" %-30.30s %-40.40s %-10s %-10v %-15.15v %-10v %-10v %-20.20s\n", key, prop.Title, prop.Type, prop.Default, prop.InitialValue, prop.ReadOnly, prop.WriteOnly, prop.Description)
	}

	fmt.Println("\nEvents:")
	fmt.Println(" ID                             Title                                    DataType   EventType       Description")
	fmt.Println(" -----------------------------  ---------------------------------------  ---------  --------------  -----------")
	keys = utils.OrderedMapKeys(tdDoc.Events)
	for _, key := range keys {
		ev := tdDoc.Events[key]
		dataType := "(n/a)"
		if ev.Data != nil {
			dataType = ev.Data.Type
		}
		fmt.Printf(" %-30.30s %-40.40s %-10.10s %-15.15s %s\n", key, ev.Title, dataType, ev.AtType, ev.Description)
	}

	fmt.Println("\nActions:")
	fmt.Println(" ID                             Title                                    Arg(s)     ActionType      Description")
	fmt.Println(" -----------------------------  ---------------------------------------  ---------  --------------  -----------")
	keys = utils.OrderedMapKeys(tdDoc.Actions)
	for _, key := range keys {
		action := tdDoc.Actions[key]
		dataType := "(n/a)"
		if action.Input != nil {
			dataType = action.Input.Type
		}
		fmt.Printf(" %-30.30s %-40.40s %-10.10s %-15.15s %s\n", key, action.Title, dataType, action.AtType, action.Description)
	}
	return err
}

// HandleListThingVerbose lists a Thing in the directory
func HandleListThingVerbose(ctx context.Context, f svcconfig.AppFolders, pubID, thingID string) error {
	var dir directory.IDirectory
	var rd directory.IReadDirectory

	conn, err := hubclient.ConnectToService(directory.ServiceName, f.Run)
	if err == nil {
		dir = capnpclient.NewDirectoryCapnpClient(ctx, conn)
		rd, err = dir.CapReadDirectory(ctx, "hubcli")
	}
	if err != nil {
		return err
	}
	tv, err := rd.GetTD(ctx, pubID, thingID)
	if err != nil {
		return err
	}
	fmt.Println("TD of", pubID, thingID)
	fmt.Printf("%s\n", tv.ValueJSON)
	return err
}
