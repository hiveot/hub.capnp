package main

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
	"github.com/wostzone/wostlib-go/pkg/tlsclient"
)

// Generate certificates in the wost certs folder if they don't exist.
// If certificates already exists then no changes will be made.
func main() {
	// set flag for help
	ifName, mac, ip := tlsclient.GetOutboundInterface("")
	_ = ifName
	_ = mac
	san := ip.String()
	// flag.String("-c", "", "Location of hub.yaml config file")
	// flag.String("-home", "", "Location application home directory")
	flag.StringVar(&san, "-san", san, "Subject name or IP address to use in certificate. Default interface "+ifName)
	hc, err := hubconfig.LoadPluginConfig("", "", nil)
	if err != nil {
		os.Exit(1)
	}
	// generate error on invalid args
	flag.Parse()

	// setup certificates only if they don't exist
	caCertFile := path.Join(hc.CertsFolder, certsetup.CaCertFile)
	certNames := []string{san}
	if _, err := os.Stat(caCertFile); os.IsNotExist(err) {
		logrus.Infof("Generating certificates in %s", hc.CertsFolder)
		certsetup.CreateCertificateBundle(certNames, hc.CertsFolder)
	} else {
		logrus.Errorf("Not generating certificates. Certificates already exist in %s", hc.CertsFolder)
		os.Exit(1)
	}
	println("Certificates generated in ", hc.CertsFolder)
	os.Exit(0)
}
