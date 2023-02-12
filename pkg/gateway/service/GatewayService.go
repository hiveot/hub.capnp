package service

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/gateway"
)

// GatewayService implements the IGatewayService interface.
// A new instance is created by the capnp server for each incoming connection.
// This service is intended as a proxy for remote services to the local resolver.
type GatewayService struct {
	resolverPath string
	sessions     map[net.Conn]*GatewaySession
	// mutex for updating sessions
	sessionMutex sync.RWMutex
	// for testing
	testAuthn authn.IAuthnService
}

// OnIncomingConnection notifies the service of a new incoming connection.
// This is invoked by the underlying protocol and returns a new session to use
// with the connection.
// If this connection closes then capabilites added in this session are removed.
// Returns nil if session is not valid
func (svc *GatewayService) OnIncomingConnection(conn net.Conn) (session *GatewaySession) {
	//var err error
	var authType = hubapi.AuthTypeUnauthenticated
	var clientID = ""
	var clientCert *x509.Certificate

	// mutual auth with client cert. Get clientID and type from cert
	tlsc, istls := conn.(*tls.Conn)
	if istls {
		err := tlsc.Handshake()
		// not a valid TLS connection so drop it
		if err != nil {
			logrus.Warningf("dropping invalid TLS connection from '%s':%s", tlsc.RemoteAddr(), err)
			_ = conn.Close()
			return nil
		}
		cstate := tlsc.ConnectionState()
		certs := cstate.PeerCertificates
		if len(certs) > 0 {
			clientCert = certs[0]
			clientID = clientCert.Subject.CommonName
			if len(clientCert.Subject.OrganizationalUnit) > 0 {
				authType = clientCert.Subject.OrganizationalUnit[0]
			}
		}

	}
	newSession, err := StartGatewaySession(svc.resolverPath, clientID, authType, tlsc, svc.testAuthn)

	if err != nil {
		logrus.Warningf("Unable to create gateway session. Closing connection...: %s", err)
		_ = conn.Close()
		return nil
	}
	if clientCert != nil {
		// save the new session
		logrus.Infof("Incoming connection with client cert from client='%s', authType='%s'", clientID, authType)
	} else if tlsc != nil {
		logrus.Infof("Incoming TLS connection without peer cert from '%s'", tlsc.RemoteAddr())
	} else {
		logrus.Infof("Incoming connection without TLS from '%s'", conn.RemoteAddr())
	}
	newSession.clientID = clientID
	newSession.authType = authType
	svc.sessionMutex.Lock()
	defer svc.sessionMutex.Unlock()
	svc.sessions[conn] = newSession
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

// Start connects to the resolver and obtains its bootstrap for forwarding requests
func (svc *GatewayService) Start() error {
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

	if len(sessionIDList) > 0 {
		logrus.Warningf("Stopping gateway service. %d sessions remaining", len(sessionIDList))
	}
	for _, sessionID := range sessionIDList {
		svc.sessionMutex.Lock()
		session := svc.sessions[sessionID]
		delete(svc.sessions, sessionID)
		session.Release()
		svc.sessionMutex.Unlock()
	}
	return err
}

// NewGatewayService returns a new instance of the gateway service.
//
//	resolverPath is the path to the resolver unix socket to pass requests to.
//	testAuthn for creating authentication clients. Used for testing only. Do not release this service.
func NewGatewayService(resolverPath string, testAuthn authn.IAuthnService) *GatewayService {
	svc := &GatewayService{
		resolverPath: resolverPath,
		testAuthn:    testAuthn,
		sessions:     make(map[net.Conn]*GatewaySession),
	}
	return svc
}
