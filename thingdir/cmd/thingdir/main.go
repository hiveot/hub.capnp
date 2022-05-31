package main

import (
	"os"
	"path"

	"github.com/wostzone/hub/authz/pkg/aclstore"
	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/proc"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/hub/thingdir/pkg/thingdir"
)

func Main() {
	main()
}

// Commandline entry point for the Thing Directory service
func main() {
	// Load the service configuration and use defaults from hubConfig
	thingDirConfig := &thingdir.ThingDirConfig{}
	hubConfig, err := config.LoadAllConfig(os.Args, "", thingdir.PluginID, &thingDirConfig)
	if err != nil {
		logrus.Fatal("thingdir configuration error: %s", err)
		os.Exit(1)
	}
	if thingDirConfig.DirAddress == "" {
		thingDirConfig.DirAddress = hubConfig.Address
	}
	if thingDirConfig.DirAclFile == "" {
		thingDirConfig.DirAclFile = path.Join(hubConfig.ConfigFolder, aclstore.DefaultAclFile)
	}
	if thingDirConfig.DirStoreFolder == "" {
		thingDirConfig.DirStoreFolder = hubConfig.ConfigFolder
	}
	if thingDirConfig.MsgbusAddress == "" {
		thingDirConfig.MsgbusAddress = hubConfig.Address
	}
	if thingDirConfig.MsgbusPortCert == 0 {
		thingDirConfig.MsgbusPortCert = hubConfig.MqttPortCert
	}

	// TODO: include server cert in hubConfig?
	serverCertPath := path.Join(hubConfig.CertsFolder, config.DefaultServerCertFile)
	serverKeyPath := path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)
	serverCert, err := certsclient.LoadTLSCertFromPEM(serverCertPath, serverKeyPath)
	if err != nil {
		logrus.Errorf("Unable to load server certificate: %s", err)
		os.Exit(1)
	}

	pb := thingdir.NewThingDir(thingDirConfig, hubConfig.CaCert, serverCert, hubConfig.PluginCert)
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
