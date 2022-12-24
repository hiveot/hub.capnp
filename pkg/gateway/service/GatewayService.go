package service

import (
	"context"
	"net"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/gateway"
)

// GatewayService implements the IGatewaySession interface.
// A new instance is created by the capnp server for each incoming connection.
// This service is intended as a proxy for remote services to the local resolver.
type GatewayService struct {
	resolverPath string
	// user authentication service connection. This instance is owned by the service and 'on loan' to the session
	userAuthn authn.IUserAuthn
	sessions  map[net.Conn]*GatewaySession
	// mutex for updating sessions
	sessionMutex sync.RWMutex
}

// OnIncomingConnection notifies the service of a new incoming connection.
// This is invoked by the underlying protocol and returns a new session to use
// with the connection.
// If this connection closes then capabilites added in this session are removed.
func (svc *GatewayService) OnIncomingConnection(conn net.Conn) (session gateway.IGatewaySession) {
	//var err error
	newSession, err := StartGatewaySession(svc.resolverPath, svc.userAuthn)
	svc.sessionMutex.Lock()
	defer svc.sessionMutex.Unlock()
	svc.sessions[conn] = newSession

	if err != nil {
		logrus.Warning("Unable to create gateway session. Closing connection...")
		conn.Close()
		return nil
	}
	return newSession
}

// OnConnectionClosed is invoked if the connection with the client has closed.
// The service will remove the session.
func (svc *GatewayService) OnConnectionClosed(conn net.Conn, session gateway.IGatewaySession) {
	_ = conn
	// remove service when connection closes
	svc.sessionMutex.Lock()
	defer svc.sessionMutex.Unlock()
	for id, s := range svc.sessions {
		if s == session {
			session.Release()
			delete(svc.sessions, id)
			break
		}
	}
}

// Start currently has nothing to do as the capnpserver listens for incoming connections
func (svc *GatewayService) Start(_ context.Context) error {
	logrus.Infof("Starting gateway service")
	return nil
}

// Stop closes all remaining sessions
func (svc *GatewayService) Stop() (err error) {

	svc.sessionMutex.RLock()
	sessionIDList := make([]net.Conn, 0, len(svc.sessions))
	for conn := range svc.sessions {
		sessionIDList = append(sessionIDList, conn)
	}
	svc.sessionMutex.RUnlock()

	logrus.Infof("Stopping gateway service. %d sessions remaining", len(sessionIDList))

	for _, sessionID := range sessionIDList {
		svc.sessionMutex.Lock()
		session := svc.sessions[sessionID]
		delete(svc.sessions, sessionID)
		session.Release()
		svc.sessionMutex.Unlock()
	}
	if svc.userAuthn != nil {
		svc.userAuthn.Release()
	}
	return err
}

// NewGatewayService returns a new instance of the gateway service.
//
//	resolverPath is the path to the resolver unix socket to pass requests to.
func NewGatewayService(resolverPath string, userAuthn authn.IUserAuthn) *GatewayService {
	svc := &GatewayService{
		resolverPath: resolverPath,
		userAuthn:    userAuthn,
		sessions:     make(map[net.Conn]*GatewaySession),
	}
	return svc
}
