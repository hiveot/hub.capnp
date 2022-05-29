package main

import (
	"flag"
	"fmt"
	"github.com/wostzone/wost-go/pkg/certsclient"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/idprov/pkg/oobclient"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/hubnet"
)

// Commandline utility to set the out of band secret for provisioning
// This uses the plugin client certificate for the neccesary permission
//  > usage postOOB deviceID secret
func main() {

	// Some additional commandline arguments for this plugin
	appDir := path.Dir(os.Args[0])
	var hostname string = hubnet.GetOutboundIP("").String()
	var certFolder string = path.Join(appDir, "../certs")
	var port uint = 9678

	flag.StringVar(&hostname, "server", hostname, "Address of the provisioning server")
	flag.UintVar(&port, "port", port, "Port of the provisioning server")
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
	addrPort := fmt.Sprintf("%s:%d", hostname, port)
	pluginCertPath := path.Join(certFolder, config.DefaultPluginCertFile)
	pluginKeyPath := path.Join(certFolder, config.DefaultPluginKeyFile)
	clientCert, err := certsclient.LoadTLSCertFromPEM(pluginCertPath, pluginKeyPath)
	if err != nil {
		logrus.Infof("Unable to load the plugin certificate/key at %s: %s", pluginCertPath, err)
		os.Exit(1)
	}
	caCertPath := path.Join(certFolder, config.DefaultCaCertFile)
	caCert, err := certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Infof("Unable to load the CA certificate at %s: %s", caCertPath, err)
		os.Exit(1)
	}
	oobClient := oobclient.NewOOBClient(addrPort, clientCert, caCert)
	err = oobClient.Start()
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
