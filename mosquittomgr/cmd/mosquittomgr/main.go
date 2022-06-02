// package main for the mosquitto manager
package main

import (
	"os"
	"path"

	"github.com/wostzone/hub/authz/pkg/aclstore"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/proc"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/hub/mosquittomgr/internal"
)

// Main entry to WoST plugin for managing Mosquitto
// This setup the configuration from file and commandline parameters and launches the service
func main() {
	logging.SetLogging("info", "")
	svcConfig := internal.MMConfig{}
	hubConfig, err := config.LoadAllConfig(os.Args, "", internal.PluginID, &svcConfig)
	logging.SetLogging(hubConfig.Loglevel, hubConfig.LogFile)

	// currently most are overridden with the hub config.
	svcConfig.Address = hubConfig.Address
	svcConfig.CaCertFile = hubConfig.CaCertFile
	if svcConfig.ClientID == "" {
		svcConfig.ClientID = internal.PluginID
	}
	if svcConfig.AclFile == "" {
		svcConfig.AclFile = path.Join(hubConfig.ConfigFolder, aclstore.DefaultAclFile)
	}
	svcConfig.LogFolder = hubConfig.LogFolder
	if svcConfig.MosquittoConfFile == "" {
		svcConfig.MosquittoConfFile = path.Join(hubConfig.ConfigFolder, internal.DefaultConfFile)
	}
	if svcConfig.MosquittoTemplateFile == "" {
		svcConfig.MosquittoTemplateFile = path.Join(hubConfig.ConfigFolder, internal.DefaultTemplateFile)
	}
	svcConfig.MqttPortCert = hubConfig.MqttPortCert
	svcConfig.MqttPortUnpw = hubConfig.MqttPortUnpw
	svcConfig.MqttPortWS = hubConfig.MqttPortWS
	svcConfig.MosqAuthPlugin = path.Join(hubConfig.BinFolder, internal.DefaultAuthPlugin)
	svcConfig.ServerCertFile = path.Join(hubConfig.CertsFolder, config.DefaultServerCertFile)
	svcConfig.ServerKeyFile = path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)

	svc := internal.NewMosquittoManager(svcConfig)
	if err != nil {
		logrus.Errorf("Mosquittomgr: Start aborted due to commandline error")
		os.Exit(1)
	}
	err = svc.Start()
	if err != nil {
		logrus.Errorf("Mosquittomgr: Failed to start: %s", err)
		os.Exit(1)
	}
	proc.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
