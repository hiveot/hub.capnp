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
//func (svc *ResolverService) GetCapability(ctx context.Context, clientID, clientType, capabilityName string, args []string) (
//	capability capnp.Client, err error) {
//	var serviceName string
//
//	svc.capsMutex.RLock()
//	defer svc.capsMutex.RUnlock()
//
//	// determine which service this belongs to
//	var capInfo *resolver.CapabilityInfo
//	for name, capList := range svc.servicesCapabilities {
//		for _, info := range capList {
//			if info.CapabilityName == capabilityName {
//				capInfo = &info
//				serviceName = name
//				break
//			}
//		}
//	}
//
//	// unknown capability
//	if capInfo == nil {
//		err = fmt.Errorf("unknown capability '%s' requested by client '%s'", capabilityName, clientID)
//		logrus.Warning(err)
//		return capability, err
//	}
//
//	// get the HiveOT canpn client to use this service
//	capHiveOTService, _ := svc.connectedServices[serviceName]
//
//	if capHiveOTService == nil {
//		// no existing connection
//		err = fmt.Errorf("no connection to service '%s'", serviceName)
//		logrus.Warning(err)
//		return capability, err
//	} else if !capHiveOTService.IsValid() {
//		// this is no longer a valid service connection
//		err = fmt.Errorf("connection to service '%s' has been lost", capInfo.ServiceName)
//		logrus.Warning(err)
//		return capability, err
//	}
//
//	// now we have the service connection, request the capability
//	// avoid a deadlock if this service itself is requested
//	if serviceName == gateway.ServiceName {
//		// can't recurse into ourselves, so just return the service capability
//		capability = capnp.Client(*capHiveOTService) //.AddRef()
//	} else {
//		method, release := capHiveOTService.GetCapability(ctx,
//			func(params hubapi.CapHiveOTService_getCapability_Params) error {
//				_ = params.SetCapabilityName(capabilityName)
//				_ = params.SetClientType(clientType)
//				_ = params.SetArgs(caphelp.MarshalStringList(args))
//				err2 := params.SetClientID(clientID)
//				return err2
//			})
//		defer release()
//		// return the future.
//		capability = method.Cap().AddRef()
//	}
//	return capability, err
//}

// HandleUnknownMethod looks up the requested method and returns a stub that forwards
// the request to its remote destination.
// If the method is not known then nil is returned.
func (svc *ResolverService) HandleUnknownMethod(
	ctx context.Context, r capnp.Recv) *server.Method {

	var capProvider *hubapi.CapProvider
	var capInfo resolver.CapabilityInfo

	// lookup the method in our service inventory
	for serviceID, capList := range svc.servicesCapabilities {
		for _, capEntry := range capList {
			if r.Method.InterfaceID == capEntry.InterfaceID &&
				r.Method.MethodID == capEntry.MethodID {
				capProvider = svc.connectedServices[serviceID]
				capInfo = capEntry
				break
			}
		}
	}
	if capProvider == nil {
		return nil
	}
	// return a helper for forwarding the request
	forwarder := ForwarderMethod{
		Method: capnp.Method{
			InterfaceID:   capInfo.InterfaceID,
			InterfaceName: capInfo.InterfaceName,
			MethodID:      capInfo.MethodID,
			MethodName:    capInfo.MethodName,
		},
		destination: (*capnp.Client)(capProvider),
	}
	return &server.Method{
		Method: r.Method,
		Impl:   forwarder.Impl,
	}
}

// ListCapabilities returns list of capabilities of all connected services sorted by service and capability names
func (svc *ResolverService) ListCapabilities(_ context.Context, clientType string) ([]resolver.CapabilityInfo, error) {
	capList := make([]resolver.CapabilityInfo, 0)
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
	sort.Slice(capList, func(i, j int) bool {
		iName := capList[i].ServiceID + capList[i].MethodName
		jName := capList[j].ServiceID + capList[j].MethodName
		return iName < jName
	})
	return capList, nil
}

// Login to the gateway
func (svc *ResolverService) Login(_ context.Context, clientID, password string) (bool, error) {
	// TODO: add credentials check
	return true, nil
}

// Ping capability
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
			logrus.Error(err)
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
func (svc *ResolverService) Start(ctx context.Context) error {
	logrus.Infof("Starting resolver service")
	err := svc.Refresh(ctx)

	svc.socketWatcher, _ = fsnotify.NewWatcher()
	err = svc.socketWatcher.Add(svc.socketDir)
	if err == nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					logrus.Infof("socket watcher ended by context")
					return
				case event := <-svc.socketWatcher.Events:
					if svc.running {
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
	svc.running = true
	return err
}

// Stop releases the connections
func (svc *ResolverService) Stop() (err error) {
	logrus.Infof("Stopping resolver service")
	if svc.running {
		svc.running = false
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
