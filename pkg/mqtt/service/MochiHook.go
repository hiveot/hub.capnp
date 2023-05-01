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

// MochiHook hooks into the mochi-co mqtt broker
type MochiHook struct {
	mqtt.HookBase
	gatewayUrl   string
	sessionMutex sync.RWMutex
	sessions     map[string]*MqttSession
	caCert       *x509.Certificate
}

func (hook *MochiHook) OnConnect(cl *mqtt.Client, pk packets.Packet) {

	session, err := NewMqttSession(hook.gatewayUrl, hook.caCert, cl)
	if err != nil {
		// reject the connection
		cl.Stop(errors.New("no connection with gateway"))
		return
	}
	hook.sessionMutex.Lock()
	hook.sessions[cl.ID] = session
	defer hook.sessionMutex.Unlock()
	logrus.Infof("New client connection clientID=%s, userName=%s", cl.ID, cl.Properties.Username)
}

func (hook *MochiHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	hook.sessionMutex.Lock()
	session := hook.sessions[cl.ID]
	session.OnDisconnect()
	delete(hook.sessions, cl.ID)
	defer hook.sessionMutex.Unlock()
	logrus.Infof("Client disconnected id=%s", cl.ID)
}

func (hook *MochiHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnect,
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
func NewMochiHook(gatewayUrl string, caCert *x509.Certificate) *MochiHook {
	svc := &MochiHook{
		caCert:       caCert,
		gatewayUrl:   gatewayUrl,
		sessionMutex: sync.RWMutex{},
		sessions:     make(map[string]*MqttSession),
	}
	return svc
}
