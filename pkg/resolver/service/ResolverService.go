package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// ResolverService implements the IResolverService interface
type ResolverService struct {
	// the combined capabilities by service name
	servicesCapabilities map[string][]resolver.CapabilityInfo

	// connected services, cache for providing capabilities by service name
	connectedServices map[string]*hubapi.CapProvider

	// watcher for config directory
	socketWatcher *fsnotify.Watcher
	// Directory containing local service sockets
	socketDir string
	// mutex for updating serviceCapabilities and connectedServices
	capsMutex sync.RWMutex
	// mutex for running Refresh scans
	refreshMutex sync.Mutex
	//
	running atomic.Bool
}

// HandleUnknownMethod looks up the requested method and returns a stub that forwards
// the request to its remote destination.
// If the method is not known then nil is returned.
func (svc *ResolverService) HandleUnknownMethod(m capnp.Method) *server.Method {

	var capProvider *hubapi.CapProvider
	svc.capsMutex.RLock()
	// lookup the method in our service inventory
	for serviceID, capList := range svc.servicesCapabilities {
		for _, capInfo := range capList {
			if m.InterfaceID == capInfo.InterfaceID &&
				m.MethodID == capInfo.MethodID {
				capProvider = svc.connectedServices[serviceID]
				// add the names for logging
				m.InterfaceName = capInfo.InterfaceName
				m.MethodName = capInfo.MethodName
				break
			}
		}
	}
	svc.capsMutex.RUnlock()
	if capProvider == nil {
		return nil
	}
	// return a helper for forwarding the request
	forwarder := NewForwarderMethod(m, (*capnp.Client)(capProvider))
	return forwarder
}

// ListCapabilities returns list of capabilities of all connected services sorted by service and capability names
func (svc *ResolverService) ListCapabilities(_ context.Context, clientType string) ([]resolver.CapabilityInfo, error) {
	capList := make([]resolver.CapabilityInfo, 0)
	svc.capsMutex.RLock()
	defer svc.capsMutex.RUnlock()
	logrus.Infof("clientType=%s", clientType)
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
	sort.Slice(capList, func(i, j int) bool {
		iName := capList[i].ServiceID + capList[i].MethodName
		jName := capList[j].ServiceID + capList[j].MethodName
		return iName < jName
	})
	logrus.Infof("listing %d capabilities from %d services", len(capList), len(svc.servicesCapabilities))
	return capList, nil
}

// Ping the resolver
func (svc *ResolverService) Ping(_ context.Context) (string, error) {
	return "pong", nil
}

// Refresh updates the map of available capabilities
// This first adds the gateway as a capability then scans the sockets, makes a connection
// and read the available capabilities from the connection.
func (svc *ResolverService) Refresh(ctx context.Context) error {
	var nrCaps int
	newCapabilityMap := make(map[string][]resolver.CapabilityInfo)
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
		capList, err := svc.scanService(ctx, serviceName, socketPath)
		if err == nil {
			nrCaps += len(capList)
			newCapabilityMap[serviceName] = capList
		} else {
			logrus.Errorf("socket '%s' offers no services: %s", socketPath, err)
		}
	}
	logrus.Infof("Found '%d' capabilities from '%d' services", nrCaps, len(dirContent))
	svc.capsMutex.Lock()
	svc.servicesCapabilities = newCapabilityMap
	svc.capsMutex.Unlock()
	return err
}

// scanService establishes a connection to the service and update its capabilities
// If a connection already exists then use it first. If it fails to connect then
// try to reconnect.
// The connection is stored and can be used for obtaining capabilities of the service.
func (svc *ResolverService) scanService(ctx context.Context,
	serviceName string, socketPath string) (newCaps []resolver.CapabilityInfo, err error) {

	newCaps = make([]resolver.CapabilityInfo, 0)

	// Validate the last established connection if it exists
	capHiveOTService, found := svc.connectedServices[serviceName]
	if found && !capHiveOTService.IsValid() {
		// is releasing still needed?
		capHiveOTService.Release()
		capHiveOTService = nil
		delete(svc.connectedServices, serviceName)
	}
	// establish a connection if it doesn't exist
	if capHiveOTService == nil {
		udsConn, err := net.Dial("unix", socketPath)
		if err != nil {
			err = fmt.Errorf("connection to service on socket '%s failed: %s",
				socketPath, err)
			return nil, err
		}
		transport := rpc.NewStreamTransport(udsConn)
		rpcConn := rpc.NewConn(transport, nil)

		// use the service bootstrap capability
		bootCap := hubapi.CapProvider(rpcConn.Bootstrap(ctx))
		capHiveOTService = &bootCap
		svc.connectedServices[serviceName] = capHiveOTService
	}
	// with a valid connection, query the capabilities
	if capHiveOTService != nil {
		method, release := capHiveOTService.ListCapabilities(ctx, nil)
		defer release()
		resp, err2 := method.Struct()
		if err = err2; err == nil {
			capInfoList, err2 := resp.InfoList()
			if err = err2; err2 == nil {
				newCaps = capserializer.UnmarshalCapabilyInfoList(capInfoList)

			}
		}
	}
	return newCaps, err
}

// Start a file watcher on the socket folder
// create the folder if it doesn't exist.
func (svc *ResolverService) Start(ctx context.Context) error {
	logrus.Infof("Starting resolver service")
	os.MkdirAll(svc.socketDir, 0700)
	err := svc.Refresh(ctx)

	svc.socketWatcher, _ = fsnotify.NewWatcher()
	err = svc.socketWatcher.Add(svc.socketDir)
	if err == nil {
		svc.running.Store(true)
		go func() {
			for {
				select {
				case <-ctx.Done():
					logrus.Infof("socket watcher ended by context")
					return
				case event := <-svc.socketWatcher.Events:
					isRunning := svc.running.Load()
					if isRunning {
						logrus.Infof("watcher event: %v", event)
						_ = svc.Refresh(ctx)
					} else {
						// socketWatcher is nil when the service has stopped
						logrus.Infof("socket watcher stopped")
						return
					}
				case err := <-svc.socketWatcher.Errors:
					logrus.Errorf("error: %s", err)
				}
			}
		}()
	}
	return err
}

// Stop releases the connections
func (svc *ResolverService) Stop() (err error) {
	logrus.Infof("Stopping resolver service")
	isRunning := svc.running.Load()
	if isRunning {
		svc.running.Store(false)
		err = svc.socketWatcher.Close()

		for _, hiveService := range svc.connectedServices {
			hiveService.Release()
		}
		svc.connectedServices = nil
	}
	return err
}

// NewResolverService returns a new instance of the service
func NewResolverService(socketDir string) *ResolverService {
	svc := &ResolverService{
		servicesCapabilities: make(map[string][]resolver.CapabilityInfo),
		connectedServices:    make(map[string]*hubapi.CapProvider),
		socketWatcher:        nil,
		socketDir:            socketDir,
		capsMutex:            sync.RWMutex{},
	}
	return svc
}
