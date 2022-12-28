package service

import (
	"context"
	"fmt"
	"sync"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/resolver"
)

// ResolverSession implements a session for incoming connections
type ResolverSession struct {

	// ID of the connected client
	clientID string

	// type of the connected client
	clientType string

	// optional capability provider
	registeredProvider hubapi.CapProvider

	// optional provider capabilities
	registeredCapabilities []resolver.CapabilityInfo

	// the resolver service is used to get capabilities from connected sessions
	resolverService *ResolverService

	rwmux sync.RWMutex
}

// Close the session
// This releases the capability connection if it exists
func (session *ResolverSession) Close() (err error) {
	logrus.Infof("closing session of client '%s'", session.clientID)
	if session.registeredProvider.IsValid() {
		session.registeredProvider.Release()
	}
	return nil
}

// GetCapability returns the capability with the given name, if available.
// This requests the capability from the resolver service which scans connected sessions.
// This method will return a 'future' interface for the service providing the capability.
func (session *ResolverSession) GetCapability(ctx context.Context,
	clientID, clientType, capabilityName string, args []string) (
	capability capnp.Client, err error) {

	return session.resolverService.GetCapability(ctx, clientID, clientType, capabilityName, args)
}

// GetRegisteredCapability is used by the resolver service to get a registered capability from its provider
func (session *ResolverSession) GetRegisteredCapability(ctx context.Context,
	clientID, clientType, capabilityName string, args []string) (
	capability capnp.Client, err error) {

	session.rwmux.RLock()
	defer session.rwmux.RUnlock()

	// determine which method this belongs to
	var capInfo *resolver.CapabilityInfo
	for _, info := range session.registeredCapabilities {
		if info.CapabilityName == capabilityName {
			capInfo = &info
			break
		}
	}

	// unknown capability
	if capInfo == nil {
		err = fmt.Errorf("unknown capability '%s' requested for client '%s'", capabilityName, clientID)
		logrus.Warning(err)
		return capability, err
	}

	// now the provider is found, request the capability
	method, release := session.registeredProvider.GetCapability(ctx,
		func(params hubapi.CapProvider_getCapability_Params) error {
			err2 := params.SetCapabilityName(capabilityName)
			_ = params.SetClientID(clientID)
			_ = params.SetClientType(clientType)
			_ = params.SetArgs(caphelp.MarshalStringList(args))
			return err2
		})
	//_ = release
	defer release()

	resp, err := method.Struct()
	if err == nil {
		// remember to use AddRef as it would otherwise be released with the method that carries it
		capability = resp.Capability().AddRef()
	}
	return capability, err
}

// ListCapabilities returns the capabilities from the resolver
func (session *ResolverSession) ListCapabilities(ctx context.Context) (
	capabilities []resolver.CapabilityInfo, err error) {

	return session.resolverService.ListCapabilities(ctx)
}

// ListRegisteredCapabilities returns the capabilities registered in this session
func (session *ResolverSession) ListRegisteredCapabilities(_ context.Context) (
	capabilities []resolver.CapabilityInfo, err error) {

	session.rwmux.RLock()
	defer session.rwmux.RUnlock()

	if session.registeredCapabilities == nil {
		err = fmt.Errorf("Client does not offer capabilities")
	}
	return session.registeredCapabilities, err
}

// Release the session
func (session *ResolverSession) Release() {
	if session.registeredProvider.IsValid() {
		session.registeredProvider.Release()
	}
}

// RegisterCapabilities makes capabilities available to others.
// This stores the capabilities along with the provider that supports getCapability.
// The session takes ownership of the provider and will release it on exit
//
//	clientID is the unique clientID of the capability provider
//	capInfo is the list with capabilities available through this provider
//	capProvider is the capnp capability provider callback interface used to obtain capabilities
func (session *ResolverSession) RegisterCapabilities(_ context.Context,
	clientID string, capInfo []resolver.CapabilityInfo, provider hubapi.CapProvider) error {

	session.rwmux.Lock()
	defer session.rwmux.Unlock()
	session.registeredCapabilities = capInfo
	session.clientID = clientID
	session.registeredProvider = provider
	return nil
}

// NewResolverSession returns a new session
func NewResolverSession(resolverService *ResolverService) *ResolverSession {
	svc := &ResolverSession{
		resolverService: resolverService,
		//sessionID:              id,
		registeredCapabilities: nil,
	}
	return svc
}
