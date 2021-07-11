package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/idprov-go/pkg/idprovoob"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// Commandline utility to set the out of band secret for provisioning
// This must be run from the hub plugin binary folder as it uses the plugin
// client certificate for the neccesary permission
// usage postOOB deviceID secret
func main() {

	// Some additional commandline arguments for this plugin
	appDir := path.Dir(os.Args[0])
	var hostname string = string(hubconfig.GetOutboundIP(""))
	var certFolder string = path.Join(appDir, "../certs")
	var port uint = 9678

	flag.StringVar(&hostname, "server", hostname, "Address of the provisioning server")
	flag.UintVar(&port, "server", port, "Port of the provisioning server")
	flag.StringVar(&certFolder, "certs", certFolder, "Folder with CA certificate and key")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Printf("Set out of band secret for device provisioning.\n")
		fmt.Printf("Usage: %s [options] <deviceID> <secret>\noptions:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	deviceID := args[0]
	secret := args[1]

	oobClient := idprovoob.NewOOBClient(hostname, port, certFolder)
	err := oobClient.Start()
	if err == nil {
		_, err = oobClient.PostOOB(deviceID, secret)
		oobClient.Stop()
	}

	if err != nil {
		logrus.Infof("Error setting OOB secret for device %s: %s", deviceID, err)
		os.Exit(1)
	}
	println("Success setting OOB secret for device %s", deviceID)
	os.Exit(0)
}
