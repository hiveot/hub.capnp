package main

import (
	"crypto/ecdsa"
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/hub/idprov/pkg/idprovserver"
	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/proc"
)

// main Parse commandline options and launch IDProvisioning protocol binding service
func main() {
	var caKey *ecdsa.PrivateKey

	// Service configuration with defaults
	idpConfig := &idprovserver.IDProvConfig{
		//CertStoreFolder: idprovserver.DefaultCertStore,
		//InstanceID:      idprovpb.PluginID,
		DisableDiscovery: false,
		//IdpPort:         idprovclient.DefaultPort,
		//IdpAddress:      "",
		//ValidForDays:    30,
	}

	// Commandline can override configuration
	// flag.StringVar(&idpConfig.Address, "address", "localhost", "Listening address of the provisioning server.")
	flag.StringVar(&idpConfig.IdpAddress, "idpAddress", idpConfig.IdpAddress, "IDP Server address. Default is Hub address")
	flag.StringVar(&idpConfig.CertStoreFolder, "certStoreFolder", idpConfig.CertStoreFolder, "Folder with provisioned certificates")
	flag.UintVar(&idpConfig.IdpPort, "idpPort", idpConfig.IdpPort, "Listening port of the provisioning server.")
	flag.StringVar(&idpConfig.ClientID, "clientID", idprovserver.PluginID, "Unique Plugin Identifier")

	appPath, _ := os.Executable()
	appFolder := path.Dir(path.Dir(appPath))
	hubConfig, err := config.LoadAllConfig(nil, appFolder, idprovclient.IdprovServiceName, &idpConfig)
	if err != nil {
		logrus.Printf("bye bye")
		os.Exit(1)
	}
	logging.SetLogging(hubConfig.Loglevel, hubConfig.LogFile)

	serverCertPath := path.Join(hubConfig.CertsFolder, config.DefaultServerCertFile)
	serverKeyPath := path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)
	serverCert, err := certsclient.LoadTLSCertFromPEM(serverCertPath, serverKeyPath)
	if err == nil {
		caKeyPath := path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)
		caKey, err = certsclient.LoadKeysFromPEM(caKeyPath)
	}
	if err != nil {
		logrus.Fatalf("idprov.main: Missing CA and/or server certificate. Unable to continue: %s", err)
		os.Exit(1)
	}
	if idpConfig.IdpAddress == "" {
		idpConfig.IdpAddress = hubConfig.Address
	}
	pb := idprovserver.NewIDProvServer(
		idpConfig,
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
