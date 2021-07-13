package main

import (
	"flag"
	"os"

	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
	"github.com/wostzone/wostlib-go/pkg/tlsclient"
)

// Generate certificates in the wost certs folder.
// If CA already exists then no changes will be made to the CA.
func main() {
	// set flag for help
	ifName, mac, ip := tlsclient.GetOutboundInterface("")
	_ = ifName
	_ = mac
	san := ip.String()
	// flag.String("-c", "", "Location of hub.yaml config file")
	// flag.String("-home", "", "Location application home directory")
	flag.StringVar(&san, "-san", san, "Subject name or IP address to use in certificate. Default interface "+ifName)
	hc, err := hubconfig.LoadHubConfig("", "")
	if err != nil {
		os.Exit(1)
	}
	// generate error on invalid args
	flag.Parse()

	// setup certificates using the mqtt server address
	sanNames := []string{hc.MqttAddress}
	certsetup.CreateCertificateBundle(sanNames, hc.CertsFolder)
	println("Certificates generated in ", hc.CertsFolder)
	os.Exit(0)
}
