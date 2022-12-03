package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/gateway"
)

// GatewayService implements the IGatewayService and IHiveOTService interfaces
type GatewayService struct {
	// the combined capabilities by service name
	servicesCapabilities map[string][]caphelp.CapabilityInfo

	// connected services, cache for providing capabilities by service name
	connectedServices map[string]*hubapi.CapHiveOTService

	// watcher for config directory
	configWatcher *fsnotify.Watcher
	// Directory containing local service sockets
	socketDir string
	// mutex for updating serviceCapabilities and connectedServices
	capsMutex sync.RWMutex
	// mutex for running Refresh scans
	refreshMutex sync.Mutex
	//
	running bool
}

// GetCapability returns the capability with the given name, if available.
// This method will return a 'future' interface for the service providing the capability.
// The provided capability must be released after use.
//
//	clientID is the ID of the client requesting the capability
//	clientType is the type of client requesting the capability
//
// Note: this method only makes sense when using an RPC such as capnp.
// The RPC Server normally intercepts this method and obtain the capability via RPC routing.
func (svc *GatewayService) GetCapability(ctx context.Context, clientID, clientType, capabilityName string, args []string) (
	capability capnp.Client, err error) {
	var serviceName string

	svc.capsMutex.RLock()
	defer svc.capsMutex.RUnlock()

	// determine which service this belongs to
	var capInfo *caphelp.CapabilityInfo
	for name, capList := range svc.servicesCapabilities {
		for _, info := range capList {
			if info.CapabilityName == capabilityName {
				capInfo = &info
				serviceName = name
				break
			}
		}
	}

	// unknown capability
	if capInfo == nil {
		err = fmt.Errorf("unknown capability '%s' requested by client '%s'", capabilityName, clientID)
		logrus.Warning(err)
		return capability, err
	}

	// get the HiveOT canpn client to use this service
	capHiveOTService, _ := svc.connectedServices[serviceName]

	if capHiveOTService == nil {
		// no existing connection
		err = fmt.Errorf("no connection to service '%s'", serviceName)
		logrus.Warning(err)
		return capability, err
	} else if !capHiveOTService.IsValid() {
		// this is no longer a valid service connection
		err = fmt.Errorf("connection to service '%s' has been lost", capInfo.ServiceName)
		logrus.Warning(err)
		return capability, err
	}

	// now we have the service connection, request the capability
	// avoid a deadlock if this service itself is requested
	if serviceName == gateway.ServiceName {
		// can't recurse into ourselves, so just return the service capability
		capability = capnp.Client(*capHiveOTService) //.AddRef()
	} else {
		method, release := capHiveOTService.GetCapability(ctx,
			func(params hubapi.CapHiveOTService_getCapability_Params) error {
				_ = params.SetCapabilityName(capabilityName)
				_ = params.SetClientType(clientType)
				_ = params.SetArgs(caphelp.MarshalStringList(args))
				err2 := params.SetClientID(clientID)
				return err2
			})
		defer release()
		// return the future.
		capability = method.Cap().AddRef()
	}
	return capability, err
}

// ListCapabilities returns an aggregated list of capabilities of all connected services
func (svc *GatewayService) ListCapabilities(_ context.Context, clientType string) ([]caphelp.CapabilityInfo, error) {
	capList := make([]caphelp.CapabilityInfo, 0)
	svc.capsMutex.RLock()
	defer svc.capsMutex.RUnlock()

	logrus.Infof("listing %d services", len(svc.servicesCapabilities))
	for _, serviceCaps := range svc.servicesCapabilities {
		for _, capInfo := range serviceCaps {
			// only include client types that are allowed
			allowedTypes := strings.Join(capInfo.ClientTypes, ",")
			isAllowed := clientType != "" && strings.Contains(allowedTypes, clientType)
			if isAllowed {
				capList = append(capList, capInfo)
			}
		}
	}
	return capList, nil
}

// Login to the gateway
func (svc *GatewayService) Login(_ context.Context, clientID, password string) (bool, error) {
	// TODO: add credentials check
	return true, nil
}

// Ping capability
func (svc *GatewayService) Ping(_ context.Context) (string, error) {
	return "pong", nil
}

// Refresh updates the map of available capabilities
// This first adds the gateway as a capability then scans the sockets, makes a connection
// and read the available capabilities from the connection.
func (svc *GatewayService) Refresh(ctx context.Context) error {
	logrus.Infof("refreshing gateway using '%s' as socket dir", svc.socketDir)
	newCapabilityMap := make(map[string][]caphelp.CapabilityInfo)
	svc.refreshMutex.Lock()
	defer svc.refreshMutex.Unlock()

	// read available sockets
	dirContent, err := os.ReadDir(svc.socketDir)
	if err != nil {
		return err
	}
	for _, socketFile := range dirContent {
		socketName := socketFile.Name()
		socketPath := path.Join(svc.socketDir, socketName)
		serviceName := strings.TrimSuffix(socketName, filepath.Ext(socketName))
		// don't recurse into ourselves. That will deadlock.
		if serviceName != gateway.ServiceName {
			capList, err := svc.scanService(ctx, serviceName, socketPath)
			if err == nil {
				newCapabilityMap[serviceName] = capList
			}
		}
	}
	svc.capsMutex.Lock()
	svc.servicesCapabilities = newCapabilityMap
	svc.capsMutex.Unlock()
	return err
}

// scanService establishes a connection to the service and update its capabilities
// If a connection already exists then use it first. If it fails to connect then
// try to reconnect.
// The connection is stored and can be used for obtaining capabilities of the service.
func (svc *GatewayService) scanService(ctx context.Context,
	serviceName string, socketPath string) (newCaps []caphelp.CapabilityInfo, err error) {
	ctxWT, _ := context.WithTimeout(ctx, time.Second)

	logrus.Infof("scanning capabilities of service '%s'", serviceName)
	newCaps = make([]caphelp.CapabilityInfo, 0)

	// Validate the last established connection if it exists
	capHiveOTService, found := svc.connectedServices[serviceName]
	if found && !capHiveOTService.IsValid() {
		logrus.Infof("connection to '%s' is no longer valid. Cleaning up", serviceName)
		// is releasing still needed?
		capHiveOTService.Release()
		capHiveOTService = nil
		delete(svc.connectedServices, serviceName)
	}
	// establish a connection if it doesn't exist
	if capHiveOTService == nil {
		udsConn, err := net.DialTimeout("unix", socketPath, time.Second)
		if err != nil {
			err = fmt.Errorf("connection to service on socket '%s failed: %s",
				socketPath, err)
			logrus.Error(err)
			return nil, err
		} else {
			//logrus.Infof("connection with '%s' established", serviceName)
		}
		transport := rpc.NewStreamTransport(udsConn)
		rpcConn := rpc.NewConn(transport, nil)

		// use the service bootstrap capability
		bootCap := hubapi.CapHiveOTService(rpcConn.Bootstrap(ctxWT))
		capHiveOTService = &bootCap
		svc.connectedServices[serviceName] = capHiveOTService
	}
	// with a valid connection, query the capabilities
	if capHiveOTService != nil {
		logrus.Infof("requesting capabilities from '%s'", serviceName)
		method, release := capHiveOTService.ListCapabilities(ctxWT, nil)
		defer release()
		resp, err2 := method.Struct()
		logrus.Infof("done. err=%s", err2)
		if err = err2; err == nil {
			capInfoList, err2 := resp.InfoList()
			if err = err2; err2 == nil {
				newCaps = caphelp.UnmarshalCapabilities(capInfoList)
			}
		}
	}
	logrus.Infof("found '%d' capabilities on '%s'", len(newCaps), serviceName)
	return newCaps, err
}

// Start a file watcher on the socket folder
func (svc *GatewayService) Start(ctx context.Context) error {
	logrus.Infof("Starting gateway service")
	err := svc.Refresh(ctx)

	svc.configWatcher, _ = fsnotify.NewWatcher()
	err = svc.configWatcher.Add(svc.socketDir)
	if err == nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					logrus.Infof("socket watcher ended by context")
					return
				case event := <-svc.configWatcher.Events:
					if svc.running {
						logrus.Infof("event: %v", event)
						_ = svc.Refresh(ctx)
					} else {
						// configWatcher is nil when the service has stopped
						logrus.Infof("socket watcher stopped")
						return
					}
				case err := <-svc.configWatcher.Errors:
					logrus.Errorf("error: %s", err)
				}
			}
		}()
	}
	svc.running = true
	return err
}

// Stop releases the connections
func (svc *GatewayService) Stop(_ context.Context) (err error) {
	logrus.Infof("Stopping gateway service")
	if svc.running {
		svc.running = false
		err = svc.configWatcher.Close()

		for _, hiveService := range svc.connectedServices {
			hiveService.Release()
		}
		svc.connectedServices = nil
	}
	return err
}

// NewGatewayService returns a new instance of the gateway service
func NewGatewayService(socketDir string) *GatewayService {
	svc := &GatewayService{
		servicesCapabilities: make(map[string][]caphelp.CapabilityInfo),
		connectedServices:    make(map[string]*hubapi.CapHiveOTService),
		configWatcher:        nil,
		socketDir:            socketDir,
		capsMutex:            sync.RWMutex{},
	}
	return svc
}
