package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/yaml.v3"

	"github.com/hiveot/hub/pkg/gateway"
)

// CapabilitiesConfigSection describes the capabilities section in the config files
type CapabilitiesConfigSection struct {
	Capabilities map[string]struct {
		ClientType []string `yaml:"clientType"`
	} `yaml:"capabilities"`
}

// GatewayService implements the IGatewayService interface
type GatewayService struct {
	capList []gateway.CapabilityInfo
	// watcher for config directory
	configWatcher *fsnotify.Watcher
	// Directory containing service configuration files with export capabilities
	configDir string
	socketDir string
	capsMutex sync.RWMutex
}

// GetCapability returns the capability with the given name, if available.
// This method will return an interface for the service providing the capability.
//
// Note: this method only makes sense when using an RPC such as capnp,
// or when everything runs within the same process.
// If we get here, then assume the latter.
// The RPC Server normally intercepts this method and obtain the capability via RPC routing.
func (svc *GatewayService) GetCapability(
	_ context.Context, clientType string, service string) (interface{}, error) {
	svc.capsMutex.RLock()
	defer svc.capsMutex.RUnlock()
	_ = clientType
	_ = service

	// This is the plan:
	//
	// 1. determine the protocol used
	// 2. if capnp, then get the capnp capability
	//    if grpc then get the grpc service interface
	//    if neither, then instantiate the service directly in this process
	// 3. wrap the protocol in a pogs client and return it.
	// 4. the caller will wrap it in its own protocol and return it to the client
	//
	// Problem: without knowing the interface types, how to proceed with this?
	//
	// Probably best if ALL services support a capnp interface. The client can wrap the
	// capability in a capnp client for that request, assuming it knows the interface.
	// If this is indeed capnp central, then this pogs service can't do anything here and it
	// all has to come from the capnpserver wrapper.
	//
	// nothing we can do here now
	return nil, fmt.Errorf("not supported")
}

// GetGatewayInfo describes the capabilities and capacity of the gateway
func (svc *GatewayService) GetGatewayInfo(_ context.Context) (info gateway.GatewayInfo, err error) {
	svc.capsMutex.RLock()
	defer svc.capsMutex.RUnlock()
	info.Capabilities = svc.capList
	info.Latency = 0
	info.URL = "capnp://localhost/"
	return info, nil
}

// ExtractCapabilitiesFromConfig extracts the capabilities from the configuration file data .. duh
func (svc *GatewayService) ExtractCapabilitiesFromConfig(serviceName string, configData []byte) []gateway.CapabilityInfo {
	capList := make([]gateway.CapabilityInfo, 0)
	svcConfig := CapabilitiesConfigSection{}
	err := yaml.Unmarshal(configData, &svcConfig)
	if err != nil {
		logrus.Errorf("failed parsing config file data from: %s", serviceName)
		return capList
	}

	// add each capability of this service in turn
	for methodName, capConfig := range svcConfig.Capabilities {
		capInfo := gateway.CapabilityInfo{
			Service:    serviceName,
			Name:       methodName,
			ClientType: capConfig.ClientType,
		}
		capList = append(capList, capInfo)
	}
	return capList
}

func (svc *GatewayService) Ping(_ context.Context) (string, error) {
	logrus.Infof("ping")
	return "pong", nil
}

// Refresh the available capabilities
// This scans the configuration files in search of a capabilities section
// {servicename}.yaml
//
//	capabilities:
//	   method:
//	      clientType: names
func (svc *GatewayService) Refresh() error {
	svc.capsMutex.Lock()
	defer svc.capsMutex.Unlock()
	capList := make([]gateway.CapabilityInfo, 0)

	dirContent, err := os.ReadDir(svc.configDir)
	if err != nil {
		return err
	}
	// read each config file looking for capabilities
	for _, dirInfo := range dirContent {
		if !dirInfo.IsDir() {
			fileName := dirInfo.Name()
			fileExt := path.Ext(fileName)
			filePath := path.Join(svc.configDir, fileName)
			serviceName := strings.TrimSuffix(fileName, fileExt)
			if fileExt == ".yaml" {
				fileData, err2 := os.ReadFile(filePath)
				if err2 == nil {
					caps := svc.ExtractCapabilitiesFromConfig(serviceName, fileData)
					capList = append(capList, caps...)
					logrus.Infof("found service '%s' with %d capabilities", serviceName, len(caps))
				} else {
					logrus.Errorf("failed reading config file '%s': %s", filePath, err2)
				}
			}
		}
	}
	svc.capList = capList
	return nil
}

// Start a file watcher on the config folder
func (svc *GatewayService) Start(ctx context.Context) error {
	err := svc.Refresh()

	svc.configWatcher, _ = fsnotify.NewWatcher()
	err = svc.configWatcher.Add(svc.configDir)
	if err == nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					logrus.Infof("Gateway service watcher ended by context")
					return
				case event := <-svc.configWatcher.Events:
					logrus.Infof("event: %v", event)
					_ = svc.Refresh()
				case err := <-svc.configWatcher.Errors:
					logrus.Errorf("error: %s", err)
				}
			}
		}()
	}
	return err
}

func (svc *GatewayService) Stop(_ context.Context) (err error) {
	if svc.configWatcher != nil {
		err = svc.configWatcher.Close()
		svc.configWatcher = nil
	}
	return err
}

// NewGatewayService returns a new instance of the gateway service
func NewGatewayService(configDir string, socketDir string) *GatewayService {
	svc := &GatewayService{
		configDir: configDir,
		socketDir: socketDir,
	}
	return svc
}
