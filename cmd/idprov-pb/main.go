package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	idprovpb "github.com/wostzone/hub/core/idprov-pb"
	"github.com/wostzone/idprov-go/pkg/idprov"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// main Parse commandline options and launch IDProvisioning protocol binding service
func main() {

	// Service configuration with defaults
	idpConfig := &idprovpb.IDProvPBConfig{
		IdpCerts:        idprovpb.DefaultCertStore,
		ClientID:        idprovpb.PluginID,
		EnableDiscovery: true,
		IdpPort:         idprov.DefaultPort,
		IdpAddress:      "",
		ValidForDays:    30,
	}

	// Commandline can override configuration
	// flag.StringVar(&idpConfig.Address, "address", "localhost", "Listening address of the provisioning server.")
	flag.StringVar(&idpConfig.IdpAddress, "idpAddress", idpConfig.IdpAddress, "IDP Server address. Default is Hub address")
	flag.StringVar(&idpConfig.IdpCerts, "idpCerts", idpConfig.IdpCerts, "Folder with provisioned certificates")
	flag.UintVar(&idpConfig.IdpPort, "idpPort", idpConfig.IdpPort, "Listening port of the provisioning server.")
	flag.StringVar(&idpConfig.ClientID, "clientID", idprovpb.PluginID, "Plugin Client ID")

	hubConfig, err := hubconfig.LoadCommandlineConfig("", idprovpb.PluginID, &idpConfig)
	if err != nil {
		logrus.Printf("bye bye")
		os.Exit(1)
	}
	// commandline overrides configfile
	// flag.Parse()

	pb := idprovpb.NewIDProvPB(idpConfig, hubConfig)
	err = pb.Start()

	if err != nil {
		logrus.Printf("Failed starting IDProvServer: %s\n", err)
		os.Exit(1)
	}
	logrus.Printf("Successful started IDProvServer\n")
	hubclient.WaitForSignal()
	logrus.Printf("IDProvServer stopped\n")
	os.Exit(0)
}
