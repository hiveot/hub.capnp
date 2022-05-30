package main

import (
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/proc"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/thingdir/pkg/thingdir"
)

func Main() {
	main()
}

// commandline entry point for the thingdir service
func main() {
	// with defaults
	thingdirConfig := &thingdir.ThingDirConfig{}
	hubConfig, err := config.LoadAllConfig(os.Args, "", thingdir.PluginID, &thingdirConfig)
	if err != nil {
		logrus.Printf("bye bye")
		os.Exit(1)
	}
	// commandline overrides configfile
	// flag.Parse()

	pb := thingdir.NewThingDirPB(thingdirConfig, hubConfig)
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
