package service

import (
	"bytes"
	"crypto/x509"
	"errors"
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
	"github.com/sirupsen/logrus"
	"sync"
)

// GatewayHook is a hiveot hook for the mochi-co mqtt broker
type GatewayHook struct {
	mqtt.HookBase
	gatewayUrl   string
	sessionMutex sync.RWMutex
	sessions     map[string]*MqttSession
	caCert       *x509.Certificate
}

// ID returns the ID of the hook.
func (hook *GatewayHook) ID() string {
	return "mqttgateway"
}

// OnACLCheck returns true/allowed for all checks.
func (hook *GatewayHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return true
}

func (hook *GatewayHook) OnConnect(cl *mqtt.Client, pk packets.Packet) {
	session, err := NewMqttSession(hook.gatewayUrl, hook.caCert, cl)
	if err != nil {
		// reject the connection
		cl.Stop(errors.New("no connection with gateway"))
		return
	}
	hook.sessionMutex.Lock()
	defer hook.sessionMutex.Unlock()
	hook.sessions[cl.ID] = session
	logrus.Infof("New client connection clientID=%s, userName=%s", cl.ID, cl.Properties.Username)
}
func (hook *GatewayHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {

	clientID := string(pk.Connect.Username)
	password := string(pk.Connect.Password)
	hook.sessionMutex.Lock()
	session, found := hook.sessions[cl.ID]
	hook.sessionMutex.Unlock()
	if !found {
		logrus.Errorf("missing session for mqtt client connection %s", cl.ID)
		return false
	}
	session.Login(clientID, password)
	return true
}

func (hook *GatewayHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	hook.sessionMutex.Lock()
	defer hook.sessionMutex.Unlock()
	session := hook.sessions[cl.ID]
	delete(hook.sessions, cl.ID)
	session.OnDisconnect()
	logrus.Infof("Client disconnected id=%s", cl.ID)
}

func (hook *GatewayHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnACLCheck,
		mqtt.OnConnect,
		mqtt.OnConnectAuthenticate,
		mqtt.OnDisconnect,
		//mqtt.OnSubscribed,
		//mqtt.OnUnsubscribed,
		//mqtt.OnPublished,
		//mqtt.OnPublish,
	}, []byte{b})
}

// NewMochiHook returns a new instance of the mochi-co mqtt server hook
//
//	serviceID is required mqtt-optional ID prefix used to listen on tcp/ws ports
func NewMochiHook(gatewayUrl string, caCert *x509.Certificate) *GatewayHook {
	svc := &GatewayHook{
		caCert:       caCert,
		gatewayUrl:   gatewayUrl,
		sessionMutex: sync.RWMutex{},
		sessions:     make(map[string]*MqttSession),
	}
	return svc
}
