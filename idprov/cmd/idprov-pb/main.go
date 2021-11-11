package main

import (
	"crypto/ecdsa"
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	idprovpb "github.com/wostzone/hub/idprov/pkg/idprov-pb"
	"github.com/wostzone/hub/lib/client/pkg/certs"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/idprovclient"
	"github.com/wostzone/hub/lib/client/pkg/proc"
)

// main Parse commandline options and launch IDProvisioning protocol binding service
func main() {
	var caKey *ecdsa.PrivateKey

	// Service configuration with defaults
	idpConfig := &idprovpb.IDProvPBConfig{
		IdpCerts:        idprovpb.DefaultCertStore,
		ClientID:        idprovpb.PluginID,
		EnableDiscovery: true,
		IdpPort:         idprovclient.DefaultPort,
		IdpAddress:      "",
		ValidForDays:    30,
	}

	// Commandline can override configuration
	// flag.StringVar(&idpConfig.Address, "address", "localhost", "Listening address of the provisioning server.")
	flag.StringVar(&idpConfig.IdpAddress, "idpAddress", idpConfig.IdpAddress, "IDP Server address. Default is Hub address")
	flag.StringVar(&idpConfig.IdpCerts, "idpCerts", idpConfig.IdpCerts, "Folder with provisioned certificates")
	flag.UintVar(&idpConfig.IdpPort, "idpPort", idpConfig.IdpPort, "Listening port of the provisioning server.")
	flag.StringVar(&idpConfig.ClientID, "clientID", idprovpb.PluginID, "Plugin Client ID")

	appPath, _ := os.Executable()
	appFolder := path.Dir(path.Dir(appPath))
	hubConfig, err := config.LoadAllConfig(nil, appFolder, idprovpb.PluginID, &idpConfig)
	if err != nil {
		logrus.Printf("bye bye")
		os.Exit(1)
	}
	// commandline overrides configfile
	// flag.Parse()
	serverCertPath := path.Join(hubConfig.CertsFolder, config.DefaultServerCertFile)
	serverKeyPath := path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)
	serverCert, err := certs.LoadTLSCertFromPEM(serverCertPath, serverKeyPath)
	if err == nil {
		caKeyPath := path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)
		caKey, err = certs.LoadKeysFromPEM(caKeyPath)
	}
	if err != nil {
		logrus.Fatalf("idprov-pb.main: Missing CA and/or server certificate. Unable to continue: %s", err)
		os.Exit(1)
	}
	pb := idprovpb.NewIDProvPB(idpConfig,
		hubConfig.MqttAddress,
		uint(hubConfig.MqttPortCert),
		uint(hubConfig.MqttPortWS),
		serverCert,
		hubConfig.CaCert,
		caKey)
	err = pb.Start()

	if err != nil {
		logrus.Printf("Failed starting IDProvServer: %s\n", err)
		os.Exit(1)
	}
	logrus.Printf("Successful started IDProvServer\n")
	proc.WaitForSignal()
	logrus.Printf("IDProvServer stopped\n")
	os.Exit(0)

}
