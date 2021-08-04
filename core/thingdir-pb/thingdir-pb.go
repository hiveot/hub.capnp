package thingdirpb

import (
	"path"

	"github.com/wostzone/thingdir-go/pkg/dirclient"
	"github.com/wostzone/thingdir-go/pkg/dirserver"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

const PluginID = "thingdir-pb"

// ThingDirPBConfig protocol binding configuration
type ThingDirPBConfig struct {
	// Directory server settings for the built-in directory server
	DisableDirServer bool   `yaml:"disableDirServer"` // Disable the built-in directory server and use an external server
	DirAddress       string `yaml:"dirAddress"`       // Directory server address, default is that of the mqtt server
	DirPort          uint   `yaml:"dirPort"`          // Directory server listening port, default is dirclient.DefaultPort
	EnableDiscovery  bool   `yaml:"enableDiscovery"`  // Enable server DNS-SD discovery
	ServiceName      string `yaml:"serviceName"`      // DNS-SD service name: "_{serviceName}._tcp" when using discovery
	ServerCertPath   string `yaml:"serverCertPath"`   // server cert location. Default is hub's server
	ServerKeyPath    string `yaml:"serverKeyPath"`    // server key location. Default is hub's key
	ServerCaPath     string `yaml:"serverCaPath"`     // server CA cert location for client auth. Default is hub's CA

	// directory client settings
	// If an external directory is used these fields must be set
	InstanceID        string `yaml:"instanceID"`  // Unique server instance ID, default is plugin ID
	DirClientCertPath string `yaml:"dirCertPath"` // Client certificate for connecting to the directory server.
	DirClientKeyPath  string `yaml:"dirKeyPath"`  // Client key location for connecting to the directory server
	DirCaPath         string `yaml:"dirCaPath"`   // Directory server CA cert location. Default is hub's CA

	// mqtt client settings
	MsgbusCertPath string `yaml:"msgbusCertPath"`   // Client certificate for connecting to the message bus.
	MsgbusKeyPath  string `yaml:"msgbusKeyPath"`    // Client key location for connecting to the message bus
	MsgbusCaPath   string `yaml:"msgbusCaCertPath"` // message bus CA cert location. Default is hub's CA

	//	VerifyPublisherInThingID bool   `yaml:"verifyPublisherInThingID"` // publisher must be the ThingID publisher
	DirectoryStoreFolder string `yaml:"storeFolder"` // location of directory files
}

// Thing Directory Protocol Binding for the WoST Hub
type ThingDirPB struct {
	config    ThingDirPBConfig
	hubConfig hubconfig.HubConfig
	dirServer *dirserver.DirectoryServer
	dirClient *dirclient.DirClient
	hubClient *hubclient.MqttHubClient
}

// Start the ThingDir service.
//  1. Launches the directory server, if enabled. disable to use an external directory
//  2. Creates a client to update the directory server
//  3. Creates a client to subscribe to TD updates on the message bus
// This automatically captures updates to TD documents published on the message bus
func (pb *ThingDirPB) Start() error {
	var err error
	// First get the directory server up and running, if not disabled
	if !pb.config.DisableDirServer {
		pb.dirServer = dirserver.NewDirectoryServer(
			pb.config.InstanceID,
			pb.config.DirectoryStoreFolder,
			pb.config.DirAddress, pb.config.DirPort,
			pb.config.ServiceName,
			pb.config.ServerCertPath, pb.config.ServerKeyPath, pb.config.ServerCaPath)

		err = pb.dirServer.Start()
		if err != nil {
			return err
		}
	}
	// connect a client to the directory server
	pb.dirClient = dirclient.NewDirClient(pb.config.DirAddress, pb.config.DirPort, pb.config.DirCaPath)
	err = pb.dirClient.ConnectWithClientCert(pb.config.DirClientCertPath, pb.config.DirClientKeyPath)
	if err != nil {
		return err
	}

	// last, start listening to TD updates on the message bus
	err = pb.hubClient.Connect()
	if err != nil {
		return err
	}
	pb.hubClient.SubscribeToTD("", pb.handleTDUpdate)

	return err
}

// Stop the ThingDir service
func (pb *ThingDirPB) Stop() {
	if pb.dirServer != nil {
		pb.dirServer.Stop()
	}
	if pb.dirClient != nil {
		pb.dirClient.Close()
	}
	if pb.hubClient != nil {
		pb.hubClient.Close()
	}
}

// NewThingDirPB creates a new Thing Directory protocol binding instance
// This uses the hub server certificate for the Thing Directory server. The server address must
// therefore match that of the certificate. Default is the hub's mqtt address.
//  config with the plugin configuration and overrides from the defaults
//  hubConfig with default server address and certificate folder
func NewThingDirPB(config *ThingDirPBConfig, hubConfig *hubconfig.HubConfig) *ThingDirPB {

	// Directory server defaults when using the built-in server
	if config.DirAddress == "" {
		config.DirAddress = hubConfig.MqttAddress
	}
	if config.DirPort == 0 {
		config.DirPort = dirclient.DefaultPort
	}
	if config.DirectoryStoreFolder == "" {
		config.DirectoryStoreFolder = hubConfig.ConfigFolder
	}
	if config.ServiceName == "" {
		config.ServiceName = dirclient.DefaultServiceName
	}
	if !config.EnableDiscovery {
		config.ServiceName = ""
	}
	if config.ServerCertPath == "" {
		config.ServerCertPath = path.Join(hubConfig.CertsFolder, certsetup.HubCertFile)
	}
	if config.ServerKeyPath == "" {
		config.ServerKeyPath = path.Join(hubConfig.CertsFolder, certsetup.HubKeyFile)
	}
	if config.ServerCaPath == "" {
		config.ServerCaPath = path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	}

	// Directory client defaults
	if config.InstanceID == "" {
		config.InstanceID = PluginID
	}
	if config.DirCaPath == "" {
		config.DirCaPath = path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	}
	if config.DirClientCertPath == "" {
		config.DirClientCertPath = path.Join(hubConfig.CertsFolder, certsetup.PluginCertFile)
	}
	if config.DirClientKeyPath == "" {
		config.DirClientKeyPath = path.Join(hubConfig.CertsFolder, certsetup.PluginKeyFile)
	}

	// Message bus client defaults
	if config.MsgbusCertPath == "" {
		config.MsgbusCertPath = path.Join(hubConfig.CertsFolder, certsetup.PluginCertFile)
	}
	if config.MsgbusKeyPath == "" {
		config.MsgbusKeyPath = path.Join(hubConfig.CertsFolder, certsetup.PluginKeyFile)
	}
	if config.MsgbusCaPath == "" {
		config.MsgbusCaPath = path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	}

	tdir := ThingDirPB{
		config:    *config,
		hubConfig: *hubConfig,
		hubClient: hubclient.NewMqttHubPluginClient(PluginID, hubConfig),
	}
	return &tdir
}
