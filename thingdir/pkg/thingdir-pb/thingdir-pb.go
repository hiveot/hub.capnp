package thingdirpb

import (
	"fmt"
	"github.com/wostzone/hub/lib/client/pkg/certsclient"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/authn/pkg/unpwstore"
	"github.com/wostzone/hub/authz/pkg/aclstore"
	"github.com/wostzone/hub/authz/pkg/authorize"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/thingdir/pkg/dirclient"
	"github.com/wostzone/hub/thingdir/pkg/dirserver"
)

const PluginID = "thingdir-pb"

// ThingDirPBConfig protocol binding configuration
type ThingDirPBConfig struct {
	// Directory server settings for the built-in directory server
	DisableDirServer bool   `yaml:"disableDirServer"` // Disable the built-in directory server and use an external server
	DirAddress       string `yaml:"dirAddress"`       // Directory server address, default is that of the mqtt server
	DirPort          uint   `yaml:"dirPort"`          // Directory server listening port, default is dirclient.DefaultPort
	ServerCertPath   string `yaml:"serverCertPath"`   // server cert location. Default is hub's server
	ServerKeyPath    string `yaml:"serverKeyPath"`    // server key location. Default is hub's key
	ServerCaPath     string `yaml:"serverCaPath"`     // server CA cert location for client auth. Default is hub's CA
	EnableLogin      bool   `yaml:"enableLogin"`      // Enable login on the server instead of using a separate auth server [false]

	// DNS-SD discovery settings
	EnableDiscovery bool   `yaml:"enableDiscovery"` // Enable server DNS-SD discovery
	ServiceName     string `yaml:"serviceName"`     // DNS-SD service name: as used in "_{serviceName}._tcp" when using discovery

	// protocl binding client settings used to connect the protocol binding to the directory server
	// If an external directory is used these fields must be set. Defaults to the internal server
	PbClientID       string `yaml:"pbClientID"`       // Unique server instance ID, default is plugin ID
	PbClientCertPath string `yaml:"pbClientCertPath"` // Client certificate for connecting to the directory server.
	PbClientKeyPath  string `yaml:"pbClientKeyPath"`  // Client key location for connecting to the directory server
	PbClientCaPath   string `yaml:"pbClientCaPath"`   // Directory server CA cert location. Default is hub's CA

	// mqtt client settings
	MsgbusCertPath string `yaml:"msgbusCertPath"`   // Client certificate for connecting to the message bus.
	MsgbusKeyPath  string `yaml:"msgbusKeyPath"`    // Client key location for connecting to the message bus
	MsgbusCaPath   string `yaml:"msgbusCaCertPath"` // message bus CA cert location. Default is hub's CA

	//	VerifyPublisherInThingID bool   `yaml:"verifyPublisherInThingID"` // publisher must be the ThingID publisher
	// directory store settings
	DirectoryStoreFolder string `yaml:"storeFolder"` // location of directory files
}

// ThingDirPB Directory Protocol Binding for the WoST Hub
type ThingDirPB struct {
	config    ThingDirPBConfig
	hubConfig config.HubConfig
	dirServer *dirserver.DirectoryServer
	dirClient *dirclient.DirClient
	hubClient *mqttclient.MqttHubClient
	//authenticator authenticate.VerifyUsernamePassword
	authorizer authorize.VerifyAuthorization
	aclStore   *aclstore.AclFileStore
	unpwStore  *unpwstore.PasswordFileStore
}

// Start the ThingDir service.
//  1. Launches the directory server, if enabled. disable to use an external directory
//  2. Creates a client to update the directory server
//  3. Creates a client to subscribe to TD updates on the message bus
// This automatically captures updates to TD documents published on the message bus
func (pb *ThingDirPB) Start() error {
	logrus.Infof("ThingDirPB.Start")
	var err error

	serverCert, err := certsclient.LoadTLSCertFromPEM(pb.config.ServerCertPath, pb.config.ServerKeyPath)
	if err != nil {
		return err
	}

	// First get the directory server up and running, if not disabled
	if !pb.config.DisableDirServer {
		// Using external or internal login authenticator?
		//var loginAuth func(string, string) bool
		//if pb.config.EnableLogin {
		//	loginAuth = pb.authenticator
		//}
		err = pb.unpwStore.Open()
		if err != nil {
			return err
		}
		err = pb.aclStore.Open()
		if err != nil {
			return err
		}

		pb.dirServer = dirserver.NewDirectoryServer(
			pb.config.PbClientID,
			pb.config.DirectoryStoreFolder,
			pb.config.DirAddress, pb.config.DirPort,
			pb.config.ServiceName,
			serverCert, pb.hubConfig.CaCert,
			//loginAuth,
			pb.authorizer)

		err = pb.dirServer.Start()
		if err != nil {
			return err
		}
	}
	// connect a client to the directory server for use by the protocol binding
	dirHostPort := fmt.Sprintf("%s:%d", pb.config.DirAddress, pb.config.DirPort)
	pb.dirClient = dirclient.NewDirClient(dirHostPort, pb.hubConfig.CaCert)
	err = pb.dirClient.ConnectWithClientCert(pb.hubConfig.PluginCert)
	if err != nil {
		return err
	}

	// last, start listening to TD updates on the message bus; use the same client certificate
	mqttHostPort := fmt.Sprintf("%s:%d", pb.hubConfig.Address, pb.hubConfig.MqttPortCert)
	err = pb.hubClient.ConnectWithClientCert(mqttHostPort, pb.hubConfig.PluginCert)
	if err != nil {
		return err
	}
	pb.hubClient.SubscribeToTD("", pb.handleTDUpdate)

	return err
}

// Stop the ThingDir service
func (pb *ThingDirPB) Stop() {
	logrus.Infof("ThingDirPB.Stop")
	if pb.hubClient != nil {
		pb.hubClient.Close()
	}
	if pb.dirClient != nil {
		pb.dirClient.Close()
	}
	if pb.dirServer != nil {
		pb.dirServer.Stop()
	}
	pb.unpwStore.Close()
	pb.aclStore.Close()

}

// NewThingDirPB creates a new Thing Directory service instance
// This uses the hub server certificate for the Thing Directory server. The server address must
// therefore match that of the certificate. Default is the hub's mqtt address.
//  config with the plugin configuration and overrides from the defaults
//  hubConfig with default server address and certificate folder
func NewThingDirPB(thingdirconf *ThingDirPBConfig, hubConfig *config.HubConfig) *ThingDirPB {

	// Directory server defaults when using the built-in server
	if thingdirconf.DirAddress == "" {
		thingdirconf.DirAddress = hubConfig.Address
	}
	if thingdirconf.DirPort == 0 {
		thingdirconf.DirPort = dirclient.DefaultPort
	}
	if thingdirconf.DirectoryStoreFolder == "" {
		thingdirconf.DirectoryStoreFolder = hubConfig.ConfigFolder
	}
	if thingdirconf.ServiceName == "" {
		thingdirconf.ServiceName = dirclient.DefaultServiceName
	}
	if !thingdirconf.EnableDiscovery {
		thingdirconf.ServiceName = ""
	}
	if thingdirconf.ServerCertPath == "" {
		thingdirconf.ServerCertPath = path.Join(hubConfig.CertsFolder, config.DefaultServerCertFile)
	}
	if thingdirconf.ServerKeyPath == "" {
		thingdirconf.ServerKeyPath = path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)
	}
	if thingdirconf.ServerCaPath == "" {
		thingdirconf.ServerCaPath = path.Join(hubConfig.CertsFolder, config.DefaultCaCertFile)
	}

	// Directory client defaults
	if thingdirconf.PbClientID == "" {
		thingdirconf.PbClientID = PluginID
	}
	if thingdirconf.PbClientCaPath == "" {
		thingdirconf.PbClientCaPath = path.Join(hubConfig.CertsFolder, config.DefaultCaCertFile)
	}
	if thingdirconf.PbClientCertPath == "" {
		thingdirconf.PbClientCertPath = path.Join(hubConfig.CertsFolder, config.DefaultPluginCertFile)
	}
	if thingdirconf.PbClientKeyPath == "" {
		thingdirconf.PbClientKeyPath = path.Join(hubConfig.CertsFolder, config.DefaultPluginKeyFile)
	}

	// Message bus client defaults
	if thingdirconf.MsgbusCertPath == "" {
		thingdirconf.MsgbusCertPath = path.Join(hubConfig.CertsFolder, config.DefaultPluginCertFile)
	}
	if thingdirconf.MsgbusKeyPath == "" {
		thingdirconf.MsgbusKeyPath = path.Join(hubConfig.CertsFolder, config.DefaultPluginKeyFile)
	}
	if thingdirconf.MsgbusCaPath == "" {
		thingdirconf.MsgbusCaPath = path.Join(hubConfig.CertsFolder, config.DefaultCaCertFile)
	}

	// The file based stores are the only option for now

	aclFile := path.Join(hubConfig.ConfigFolder, aclstore.DefaultAclFile)
	aclStore := aclstore.NewAclFileStore(aclFile, "ThingDirPB")

	unpwFile := path.Join(hubConfig.ConfigFolder, unpwstore.DefaultPasswordFile)
	unpwStore := unpwstore.NewPasswordFileStore(unpwFile, "ThingDirPB")

	tdir := ThingDirPB{
		config:    *thingdirconf,
		hubConfig: *hubConfig,
		hubClient: mqttclient.NewMqttHubClient(PluginID, hubConfig.CaCert),
		aclStore:  aclStore,
		//authenticator: authenticate.NewAuthenticator(unpwStore).VerifyUsernamePassword,
		unpwStore:  unpwStore,
		authorizer: authorize.NewAuthorizer(aclStore).VerifyAuthorization,
	}
	return &tdir
}
