package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/proc"
	thingdirpb "github.com/wostzone/hub/thingdir/pkg/thingdir-pb"
)

func Main() {
	main()
}

// main entry point for the thingdir protocol binding service
func main() {
	// with defaults
	thingdirConfig := &thingdirpb.ThingDirPBConfig{}
	hubConfig, err := config.LoadAllConfig(os.Args, "", thingdirpb.PluginID, &thingdirConfig)
	if err != nil {
		logrus.Printf("bye bye")
		os.Exit(1)
	}
	// commandline overrides configfile
	// flag.Parse()

	pb := thingdirpb.NewThingDirPB(thingdirConfig, hubConfig)
	err = pb.Start()

	if err != nil {
		logrus.Printf("Failed starting Thing Directory server. Is the mosquitto broker running?: %s\n", err)
		os.Exit(1)
	}
	logrus.Printf("Successful started Thing Directory server\n")
	proc.WaitForSignal()

	pb.Stop()
	os.Exit(0)
}
