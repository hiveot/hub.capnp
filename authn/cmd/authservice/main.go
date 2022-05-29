package main

import (
	"crypto/ecdsa"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/proc"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/authn/pkg/authservice"
	"github.com/wostzone/hub/authn/pkg/unpwstore"
	"github.com/wostzone/wost-go/pkg/certsclient"
)

const DefaultUserConfigFolderName = "configStore"

func Main() {
	main()
}

// Main entry point to start the authentication service
func main() {
	// with defaults
	authServiceConfig := authservice.AuthServiceConfig{}
	authServiceConfig.ConfigStoreEnabled = true
	hubConfig, err := config.LoadAllConfig(os.Args, "", authservice.PluginID, &authServiceConfig)
	if err != nil {
		logrus.Printf("bye bye")
		os.Exit(1)
	}
	logging.SetLogging(hubConfig.Loglevel, hubConfig.LogFile)

	// sensible defaults
	if authServiceConfig.ConfigStoreFolder == "" {
		// this service offers a configuration store for clients
		defaultUserConfigFolder := path.Join(hubConfig.ConfigFolder, DefaultUserConfigFolderName)
		authServiceConfig.ConfigStoreFolder = path.Join(hubConfig.ConfigFolder, defaultUserConfigFolder)
	}
	if authServiceConfig.PasswordFile == "" {
		authServiceConfig.PasswordFile = path.Join(hubConfig.ConfigFolder, unpwstore.DefaultPasswordFile)
	}
	// the
	if authServiceConfig.Address == "" {
		authServiceConfig.Address = hubConfig.Address
	}

	// this server must have a certificate
	serverCertPath := path.Join(hubConfig.CertsFolder, config.DefaultServerCertFile)
	serverKeyPath := path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)
	logrus.Printf("Loading authn server certf from: %s\n", serverCertPath)
	serverCert, err := certsclient.LoadTLSCertFromPEM(serverCertPath, serverKeyPath)
	if err != nil {
		logrus.Printf("Failed load TLS Server certificate for the Auth Service.: %s\n", err)
		os.Exit(1)
	}
	// signing of tokens is done using the server certificate, available to all servers
	signingKey := serverCert.PrivateKey.(*ecdsa.PrivateKey)
	pb := authservice.NewJwtAuthService(
		authServiceConfig,
		signingKey,
		serverCert,
		hubConfig.CaCert)
	err = pb.Start()

	if err != nil {
		logrus.Printf("Failed starting Auth Service.: %s\n", err)
		os.Exit(1)
	}
	logrus.Printf("Successful started authentication server\n")
	proc.WaitForSignal()

	pb.Stop()
	os.Exit(0)
}
