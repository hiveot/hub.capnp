package thingdir

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"

	"github.com/wostzone/hub/authz/pkg/aclstore"
	"github.com/wostzone/hub/authz/pkg/authorize"
	"github.com/wostzone/hub/thingdir/pkg/dirclient"
	"github.com/wostzone/hub/thingdir/pkg/dirserver"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/consumedthing"
	"github.com/wostzone/wost-go/pkg/hubnet"
	"github.com/wostzone/wost-go/pkg/mqttclient"
)

const PluginID = "thingdir"

// ThingDirConfig plugin configuration
type ThingDirConfig struct {
	//--- Directory server settings  ---

	// Service instance client ID used in publishing its own TD, discovery, and connecting to the message bus.
	// Must be unique on the local network. Default is the pluginID.
	InstanceID string `yaml:"clientID"`

	// Directory server listening address
	// Default is the outbound IP
	DirAddress string `yaml:"dirAddress"`

	// Directory server listening port,
	// Default is dirclient.DefaultPort
	DirPort uint `yaml:"dirPort"`

	// Directory authorization ACL file location.
	// Required. Panics if not provided.
	DirAclFile string `yaml:"aclFile"`

	// DirStoreFolder holds the location of directory database
	// Required. Panics if not provided.
	DirStoreFolder string `yaml:"storeFolder"`

	// Enable auth login on the server instead of using a separate auth server [false]
	//EnableLogin bool `yaml:"enableLogin"`

	//--- DNS-SD discovery settings ---

	// DNS-SD service name: as used in "_{serviceName}._tcp" when using discovery
	// Default is {dirclient.DefaultServiceName} (thingdir)
	DiscoveryServiceName string `yaml:"serviceName"`
	// Enable server DNS-SD discovery. Default is disabled.
	EnableDiscovery bool `yaml:"enableDiscovery"`

	//--- mqtt client settings ---
	// Mqtt server address, when different from the directory server address
	// Default is the Outbound IP
	MsgbusAddress string `yaml:"msgbusAddress"`
	// Certificate authentication auth port. Default is {config.DefaultMqttPortCert}
	MsgbusPortCert int `yaml:"msgbusPortCert"`
}

// ThingDir binds the mqtt message bus to the Directory Server on the WoST Hub.
// It stores published TD documents into the directory server and tracks the latest
// property values.
type ThingDir struct {
	// Plugin configuration
	config ThingDirConfig
	// The directory server implementing the directory HTTP API
	dirServer *dirserver.DirectoryServer
	// MQTT client to capture TD documents and property values
	mqttClient *mqttclient.MqttClient
	// Authorizer for the directory server
	//authorizer authorize.VerifyAuthorization
	// ACL for the directory server
	aclStore *aclstore.AclFileStore
	// CA certificates to validate client certs
	caCert *x509.Certificate
	// certificates for the directory server
	serverCert *tls.Certificate
	// certificates for the mqtt client (using cert auth)
	clientCert *tls.Certificate
}

// Start the ThingDir service.
// This service will capture TD documents from the message bus and updated the Directory
// Server store with received TD's and with received property values. If enabled, enabled
// DNS-SD discovery of the service.
//
//  1. Launches the directory server, if enabled.
//  2. Creates a client to update the directory server with received TDs.
//  3. Creates a client to subscribe to TD updates on the message bus.
func (tDir *ThingDir) Start() error {
	logrus.Infof("ThingDirPB.Start")
	var err error

	err = tDir.aclStore.Open()
	if err != nil {
		return err
	}

	err = tDir.dirServer.Start()
	if err != nil {
		return err
	}
	// Listen for TD updates on the message bus
	mqttHostPort := fmt.Sprintf("%s:%d", tDir.config.MsgbusAddress, tDir.config.MsgbusPortCert)
	err = tDir.mqttClient.ConnectWithClientCert(mqttHostPort, tDir.clientCert)
	if err != nil {
		return err
	}
	topic := consumedthing.CreateTopic("", mqttbinding.TopicTypeTD)
	tDir.mqttClient.Subscribe(topic, tDir.handleTDUpdate)

	// Listen for events
	topic = consumedthing.CreateTopic("", mqttbinding.TopicTypeEvent) + "/+"
	tDir.mqttClient.Subscribe(topic, tDir.handleEvent)

	return err
}

// Stop the ThingDir service
func (tDir *ThingDir) Stop() {
	logrus.Infof("ThingDirPB.Stop")
	if tDir.mqttClient != nil {
		tDir.mqttClient.Disconnect()
	}
	//if tDir.dirClient != nil {
	//	tDir.dirClient.Close()
	//}
	if tDir.dirServer != nil {
		tDir.dirServer.Stop()
	}
	//tDir.unpwStore.Close()
	tDir.aclStore.Close()
}

// NewThingDir creates a new Thing Directory service instance
// This uses the hub server certificate for the Thing Directory server. The server address must
// therefore match that of the certificate. Default is the hub's mqtt address.
// This will modify the provided config with values actually used.
//
//  tDirConfig with the plugin configuration and overrides from the defaults
//  caCert with CA to use for server verification
//  serverCert with directory server TLS certificate
//  clientCert to authenticate with the message bus
func NewThingDir(
	tDirConfig *ThingDirConfig,
	caCert *x509.Certificate,
	serverCert *tls.Certificate,
	clientCert *tls.Certificate,
) *ThingDir {

	// Directory server defaults when using the outbound IP
	if tDirConfig.DirAddress == "" {
		tDirConfig.DirAddress = hubnet.GetOutboundIP("").String()
	}
	if tDirConfig.DirPort == 0 {
		tDirConfig.DirPort = dirclient.DefaultPort
	}
	if tDirConfig.DirAclFile == "" {
		logrus.Panic("Missing directory authorization ACL file location")
	}
	if tDirConfig.DirStoreFolder == "" {
		logrus.Panic("Missing directory store location")
	}
	if tDirConfig.DiscoveryServiceName == "" {
		tDirConfig.DiscoveryServiceName = dirclient.DefaultServiceName
	}
	if !tDirConfig.EnableDiscovery {
		tDirConfig.DiscoveryServiceName = ""
	}

	// Message bus client defaults
	if tDirConfig.MsgbusAddress == "" {
		tDirConfig.MsgbusAddress = hubnet.GetOutboundIP("").String()
	}
	if tDirConfig.MsgbusPortCert == 0 {
		tDirConfig.MsgbusPortCert = config.DefaultMqttPortCert
	}

	aclStore := aclstore.NewAclFileStore(tDirConfig.DirAclFile, tDirConfig.InstanceID)
	authorizer := authorize.NewAuthorizer(aclStore).VerifyAuthorization

	dirServer := dirserver.NewDirectoryServer(
		tDirConfig.InstanceID,
		tDirConfig.DirStoreFolder,
		tDirConfig.DirAddress,
		tDirConfig.DirPort,
		tDirConfig.DiscoveryServiceName,
		serverCert,
		caCert,
		authorizer)

	tdir := ThingDir{
		config:     *tDirConfig,
		dirServer:  dirServer,
		mqttClient: mqttclient.NewMqttClient(PluginID, caCert, 0),
		aclStore:   aclStore,
		caCert:     caCert,
		serverCert: serverCert,
		clientCert: clientCert,
	}
	return &tdir
}
