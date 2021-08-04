package main

import (
	"os"

	"github.com/sirupsen/logrus"
	thingdirpb "github.com/wostzone/hub/core/thingdir-pb"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// main entry point for thingdir-pb protocol binding
func main() {
	// with defaults
	thingdirConfig := &thingdirpb.ThingDirPBConfig{}
	hubConfig, err := hubconfig.LoadCommandlineConfig("", thingdirpb.PluginID, &thingdirConfig)
	if err != nil {
		logrus.Printf("bye bye")
		os.Exit(1)
	}
	// commandline overrides configfile
	// flag.Parse()

	pb := thingdirpb.NewThingDirPB(thingdirConfig, hubConfig)
	err = pb.Start()

	if err != nil {
		logrus.Printf("Failed starting Thing Directory server: %s\n", err)
		os.Exit(1)
	}
	logrus.Printf("Successful started Thing Directory server\n")
	hubclient.WaitForSignal()
	logrus.Printf("ThingDir server stopped\n")
	os.Exit(0)
}
