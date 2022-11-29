package caphelp

import (
	"context"
	"fmt"
	"strings"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// IHiveOTService defines the POGS interface that all HiveOT Services must implement.
// The reason is to be able to dynamically obtain and invoke the capabilites provided
// by the service.
type IHiveOTService interface {

	// GetCapability obtains the capability with the given name.
	// This returns the client for that capability, or nil if the capability is not available.
	// The client must cast the result to the appropriate interface.
	//
	// All capabilities must be released after use.
	//
	// clientID is the ID of the client, in case further ID related auth is needed.
	// clientType is the type of authenticated client
	// capabilityName is the name to retrieve
	// args is an array with arguments as per API with the same name
	GetCapability(ctx context.Context,
		clientID string, clientType string, capabilityName string, args []string) (
		capability capnp.Client, err error)

	// ListCapabilities lists the available capabilities of the service
	// Returns a list of capabilities that can be obtained through the service
	ListCapabilities(ctx context.Context) (infoList []CapabilityInfo, err error)

	// Stop the service and free its resources
	Stop(ctx context.Context) error
}

//type CallInfo struct {
//	//callType struct{}
//	//method   func(context.Context, capnp.Struct) error
//	method interface{}
//	//method func(ctx context.Context, args ...string) (interface{}, error)
//}

// HiveOTServiceCapnpServer implements the capnp server listing and getting capabilities
// available to remote users (through the gateway).
// Embed this server within the service's capnp server and use ExportCapability to enable
// to use it with ListCapabilities and GetCapability.
//
// This implements the capnp HiveOTService_Server interface
type HiveOTServiceCapnpServer struct {
	svc                  IHiveOTService
	exportedCapabilities map[string]CapabilityInfo
	knownMethods         map[string]server.Method
	serviceName          string
}

// ExportCapability adds the method to the list of exported capabilities and allow it to be
// returned in GetCapability.
func (capsrv *HiveOTServiceCapnpServer) ExportCapability(methodName string, clientTypes []string) {
	_, found := capsrv.knownMethods[methodName]
	if !found {
		err := fmt.Errorf("method '%s' is not a known method. Unable to enable it", methodName)
		logrus.Error(err)
		panic(err)
	}
	newCap := CapabilityInfo{
		CapabilityName: methodName,
		ClientTypes:    clientTypes,
		ServiceName:    capsrv.serviceName,
	}
	capsrv.exportedCapabilities[methodName] = newCap
}

// GetCapability invokes the requested method to return the capability it provides
// This returns an error if the capability is not found or not available to the client type
func (capsrv *HiveOTServiceCapnpServer) GetCapability(
	ctx context.Context, call hubapi.CapHiveOTService_getCapability) (err error) {

	args := call.Args()
	capabilityName, _ := args.CapabilityName()
	clientType, _ := args.ClientType()
	methodArgsCapnp, _ := args.Args()
	methodArgs := UnmarshalStringList(methodArgsCapnp)
	_ = methodArgs

	capInfo, found := capsrv.exportedCapabilities[capabilityName]
	if !found {
		err = fmt.Errorf("capability '%s' not found", capabilityName)
		return err
	}

	// validate the client type is allowed on this method
	allowedTypes := strings.Join(capInfo.ClientTypes, ",")
	isAllowed := clientType != "" && strings.Contains(allowedTypes, clientType)
	if !isAllowed {
		err = fmt.Errorf("capability '%s' is not available to clients of type '%s'", capabilityName, clientType)
		return err
	}
	// invoke the method to get the capability
	// the clientID and clientType arguments from this call are passed on to the capability request
	// TODO: is there a need for further arugments?
	method, _ := capsrv.knownMethods[capabilityName]
	mc := call.Call
	//for i, argText := range methodArgs {
	//	mc.Args().SetText()
	//}
	err = method.Impl(ctx, mc)
	return err
}

// ListCapabilities returns the list of registered capabilities
// Only capabilities with a client type are returned.
func (capsrv *HiveOTServiceCapnpServer) ListCapabilities(
	_ context.Context, call hubapi.CapHiveOTService_listCapabilities) error {
	var err error

	infoList := make([]CapabilityInfo, 0, len(capsrv.exportedCapabilities))
	for _, capInfo := range capsrv.exportedCapabilities {
		infoList = append(infoList, capInfo)
	}

	if err == nil {
		resp, err2 := call.AllocResults()
		err = err2
		if err == nil {
			infoListCapnp := MarshalCapabilities(infoList)
			err = resp.SetInfoList(infoListCapnp)
		}
	}
	return err
}

// RegisterKnownMethods stores the known methods for use by getCapability.
// ExportCapability must be called for a method to be exposed
func (capsrv *HiveOTServiceCapnpServer) RegisterKnownMethods(methods []server.Method) {
	for _, method := range methods {
		capsrv.knownMethods[method.MethodName] = method
	}
}

// NewHiveOTServiceCapnpServer creates a capnp server handler that implements the HiveOTService interface
// for retrieving capabilities by remote users through the gateway.
// Use ExportCapability to allow remote use of the method.
//
//	methods is the list of known methods generated by capnp.
//	serviceName is the name of the service whose methods are made available
func NewHiveOTServiceCapnpServer(serviceName string) HiveOTServiceCapnpServer {
	srv := HiveOTServiceCapnpServer{
		exportedCapabilities: make(map[string]CapabilityInfo),
		knownMethods:         make(map[string]server.Method),
		serviceName:          serviceName,
	}
	return srv
}
