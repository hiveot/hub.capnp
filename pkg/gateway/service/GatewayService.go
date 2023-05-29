package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/pkg/authn/service/jwtauthn"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
	"net"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/authn"
)

const GatewayTokenValiditySec = 3600 * 24 * 14

// GatewayService implements the IGatewayService interface.
// A new instance is created by the capnp server for each incoming connection.
// This service is intended as a proxy for remote services to the local resolver.
type GatewayService struct {
	//resolverPath string
	// sessions by session token
	sessions map[string]*GatewaySession
	// mutex for updating sessions
	sessionMutex sync.RWMutex
	// for token management
	tokenMng *jwtauthn.JWTAuthn
	// Cached user authn capability for login and refresh
	authnService authn.IAuthnService
	userAuthn    authn.IUserAuthn
	// resolver capability for use by sessions
	resCap *capnpclient.ResolverCapnpClient
}

// Provide the user authentication service
//func (svc *GatewayService) getUserAuthn(
//	ctx context.Context, clientID string) (userAuthn authn.IUserAuthn, err error) {
//
//	if svc.userAuthn == nil {
//		// if the authn service is available ask for the capability, otherwise fail
//		if svc.authnService == nil {
//			err = errors.New("authn service not available") // not available
//		} else {
//			svc.userAuthn, err = svc.authnService.CapUserAuthn(ctx, clientID)
//		}
//	}
//	if err != nil {
//		err = fmt.Errorf("can't connect to the authn service: %s", err)
//		logrus.Error(err)
//	}
//	return svc.userAuthn, err
//}

// AuthNoAuth returns an 'unauthenticated' session token
func (svc *GatewayService) AuthNoAuth(clientID string) string {
	t, _ := svc.tokenMng.CreateToken(clientID, hubapi.AuthTypeUnauthenticated, GatewayTokenValiditySec)
	return t
}

// AuthProxy returns a session token for another service
func (svc *GatewayService) AuthProxy(c net.Conn, clientID string, clientCertPEM string) (sessionToken string, err error) {
	clientType := "" // unauthenticated

	// connection must have a valid peer certificate
	peerCert, certClientID, ou, err := listener.GetPeerCert(c)
	// lots of validation
	if err != nil {
		return "", err
	} else if peerCert == nil || certClientID == "" {
		return "", errors.New("invalid peer certificate")
	} else if ou != certsclient.OUService {
		// TODO: add a OUProxyService
		return "", errors.New("peer is not a proxy service")
	}

	// the given client certificate must match the given client ID
	ccert, err := certsclient.LoadX509CertFromPEM(clientCertPEM)
	if err != nil {
		return "", err
	}
	if clientID != ccert.Subject.CommonName {
		return "", errors.New("client certificate belongs to different client")
	} else if len(ccert.Subject.OrganizationalUnit) > 0 {
		clientType = ccert.Subject.OrganizationalUnit[0]
	}

	t, _ := svc.tokenMng.CreateToken(clientID, clientType, GatewayTokenValiditySec)

	return t, nil
}

// AuthRefresh refreshes a session token
func (svc *GatewayService) AuthRefresh(clientID, oldSessionToken string) (string, error) {
	// token must be valid
	_, claims, err := svc.tokenMng.ValidateToken(clientID, oldSessionToken)
	if err != nil {
		return "", err
	}
	newToken, err := svc.tokenMng.CreateToken(clientID, claims.Audience, GatewayTokenValiditySec)
	return newToken, err
}

// AuthWithCert returns a session token using the peer certificate
func (svc *GatewayService) AuthWithCert(c net.Conn) (string, error) {
	// connection must have a valid peer cert
	peerCert, clientID, authType, err := listener.GetPeerCert(c)
	if err != nil || peerCert == nil {
		return "", err
	}
	newToken, err := svc.tokenMng.CreateToken(clientID, authType, GatewayTokenValiditySec)
	return newToken, err
}

// AuthWithPassword returns a session token for user login/password auth
func (svc *GatewayService) AuthWithPassword(loginID string, password string) (string, error) {
	// need authn capability to login

	if svc.userAuthn == nil {
		return "", fmt.Errorf("user authentication is not available")
	}
	_, refreshToken, err := svc.userAuthn.Login(context.Background(), password)
	// todo: ensure this token is valid for NewSession. maybe not use it?
	if err != nil {
		return "", err
	}
	return refreshToken, err
}

// NewSession returns a new session for the given token
// This returns an error if the token is invalid or already be in use.
func (svc *GatewayService) NewSession(clientID string, sessionToken string) (*GatewaySession, error) {
	// test token
	token, claims, err := svc.tokenMng.ValidateToken(clientID, sessionToken)
	_ = token
	if err != nil {
		err = fmt.Errorf("new session for clientID '%s' has invalid session token: %w", clientID, sessionToken)
		return nil, err
	}
	_, hasSession := svc.sessions[sessionToken]
	if hasSession {
		return nil, fmt.Errorf("session token is already in use")
	}

	// only capabilities with a matching authType are available in this session
	authType := claims.Subject
	// todo, get available capabilities for this authType

	// create session
	newSession := NewGatewaySession(clientID, authType, svc.resCap)
	svc.sessionMutex.Lock()
	defer svc.sessionMutex.Unlock()
	svc.sessions[sessionToken] = newSession
	return newSession, nil
}

// OnIncomingConnection notifies the service of a new incoming connection.
// This is invoked by the underlying protocol and returns a new session to use
// with the connection.
// If this connection closes then capabilites added in this session are removed.
// Returns nil if session is not valid
//func (svc *GatewayService) OnIncomingConnection(conn net.Conn) (session *GatewaySession) {
//	//var err error
//	var clientID = ""
//
//	tlsc := conn.(*tls.Conn)
//	if tlsc == nil {
//		logrus.Warningf("connection from '%s' is not TLS", conn.RemoteAddr().String())
//		return
//	}
//	peerCert, clientID, authType, err := listener.GetPeerCert(conn)
//	newSession, err := StartGatewaySession(svc.resolverPath, clientID, authType, tlsc, svc.testAuthn)
//
//	if err != nil {
//		logrus.Warningf("Unable to create gateway session. Closing connection...: %s", err)
//		_ = conn.Close()
//		return nil
//	}
//	if peerCert != nil {
//		// save the new session
//		logrus.Infof("Incoming connection with client cert from client='%s', authType='%s'", clientID, authType)
//	} else if tlsc != nil {
//		logrus.Infof("Incoming TLS connection without peer cert from '%s'", tlsc.RemoteAddr())
//	} else {
//		logrus.Infof("Incoming connection without TLS from '%s'", conn.RemoteAddr())
//	}
//	newSession.clientID = clientID
//	newSession.authType = authType
//	svc.sessionMutex.Lock()
//	defer svc.sessionMutex.Unlock()
//	svc.sessions[conn] = newSession
//	return newSession
//}

// OnSessionClosed is invoked if the session with the client has been released.
// The service will remove the session.
// This must be invoked by the protocol binding when its connect/client closes.
func (svc *GatewayService) OnSessionClosed(sessionID string) {

	// remove service when connection closes
	svc.sessionMutex.Lock()
	defer svc.sessionMutex.Unlock()
	for id, session := range svc.sessions {
		if id == sessionID {
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

	// determine which sessions still need to be closed
	svc.sessionMutex.RLock()
	sessionIDList := make([]string, 0, len(svc.sessions))
	for sessionToken := range svc.sessions {
		sessionIDList = append(sessionIDList, sessionToken)
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
//	resolverCap is the capnp client to the resolver for use by sessions to get capabilities
//	userAuthn to verify password login
func NewGatewayService(resolverCap *capnpclient.ResolverCapnpClient, userAuthn authn.IUserAuthn) *GatewayService {
	// session tokens are signed by a randomly generated signing key and are valid
	// until they expire or the service has restarted.
	// validity period uses the default 14 days
	signingKey := certsclient.CreateECDSAKeys()
	tokenMng := jwtauthn.NewJWTAuthn(signingKey, 0, 0)

	svc := &GatewayService{
		resCap:    resolverCap,
		userAuthn: userAuthn,
		sessions:  make(map[string]*GatewaySession),
		tokenMng:  tokenMng,
	}

	return svc
}
