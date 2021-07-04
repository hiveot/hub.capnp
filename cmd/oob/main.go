package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/idprov-pb/pkg/oob"
)

// Commandline utility to set the out of band secret for provisioning
// This must be run from the hub plugin binary folder as it uses the plugin
// client certificate for the neccesary permission
// usage postOOB deviceID secret
func main() {

	// Some additional commandline arguments for this plugin
	appDir := path.Dir(os.Args[0])
	var hostname string = ""
	var certFolder string = path.Join(appDir, "../certs")

	flag.StringVar(&hostname, "server", "localhost:9678", "Hostname/IP:port of the provisioning server. Default localhost")
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

	oobClient := oob.NewOOBClient()
	err := oobClient.Start(hostname, certFolder)
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
