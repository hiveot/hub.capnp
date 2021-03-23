package main

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/pkg/certsetup"
	"github.com/wostzone/hubapi/pkg/hubconfig"
)

// Generate certificates if they don't exist
func main() {
	// set flag for help
	var hostname string = ""
	flag.String("-c", "", "Location of hub.yaml config file")
	flag.String("-home", "", "Location application home directory")
	flag.StringVar(&hostname, "-hostname", "localhost", "Hostname or IP to use in certificate. Default localhost for testing.")
	hc, err := hubconfig.LoadPluginConfig("", "", nil)
	if err != nil {
		os.Exit(1)
	}
	// generate error on invalid args
	flag.Parse()

	// setup certificates only if they don't exist
	caCertFile := path.Join(hc.Messenger.CertFolder, certsetup.CaCertFile)
	if _, err := os.Stat(caCertFile); os.IsNotExist(err) {
		logrus.Infof("Generating certificates in %s", hc.Messenger.CertFolder)
		certsetup.CreateCertificates("", hc.Messenger.CertFolder)
	} else {
		logrus.Errorf("Not generating certificates. Certificates already exist in %s", hc.Messenger.CertFolder)
		os.Exit(1)
	}
	println("Certificates generated in ", hc.Messenger.CertFolder)
	os.Exit(0)
}
