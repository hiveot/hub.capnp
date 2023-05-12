package service

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/mqtt/mqttclient"
	"github.com/hiveot/hub/pkg/pubsub/service"
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
	"github.com/sirupsen/logrus"
	"sync"
)

// GatewayHook is a hiveot hook for the mochi-co mqtt broker
type GatewayHook struct {
	mqtt.HookBase
	sessionMutex sync.RWMutex
	sessions     map[string]*MqttSession
	caCert       *x509.Certificate
}

// ID returns the ID of the hook.
func (hook *GatewayHook) ID() string {
	return "mqttgateway"
}

func (hook *GatewayHook) OnAuthPacket(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	// TBD
	return pk, nil
}

// OnACLCheck returns true/allowed for all checks.
func (hook *GatewayHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return true
}

// OnConnect creates a new mqtt session
func (hook *GatewayHook) OnConnect(cl *mqtt.Client, pk packets.Packet) {
	session, err := NewMqttSession(hook.caCert, cl)
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
	err := session.Login(clientID, password)
	return err == nil
}

// OnDisconnect releases the session for this client
func (hook *GatewayHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	hook.sessionMutex.Lock()
	defer hook.sessionMutex.Unlock()
	session := hook.sessions[cl.ID]
	delete(hook.sessions, cl.ID)
	session.OnDisconnect()
	logrus.Infof("Client disconnected id=%s", cl.ID)
}

func (hook *GatewayHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (pkx packets.Packet, err error) {
	hook.sessionMutex.Lock()
	defer hook.sessionMutex.Unlock()
	session := hook.sessions[cl.ID]
	if session != nil {
		err = session.OnPublish(pk.TopicName, pk.Payload)
		if err != nil {
			logrus.Error(err)
			return pkx, err
		}
	}
	// FIXME: don't publish this on the mqtt bus as it has to go through the pubsub service.
	// ugh, undocumented stuff
	err = packets.ErrRejectPacket
	return pk, err
}

// OnSubscribe proxies the subscription to the pubsub service
// The topic format of the pubsub and mqtt can differ.
func (hook *GatewayHook) OnSubscribe(cl *mqtt.Client, pk packets.Packet) (pkx packets.Packet) {
	var err error
	hook.sessionMutex.Lock()
	defer hook.sessionMutex.Unlock()
	session := hook.sessions[cl.ID]
	mqttTopic := pk.Filters[0].Filter
	pubID, thingID, msgType, name, err := mqttclient.SplitTopic(mqttTopic)
	logrus.Infof("OnSubscribe to %s", mqttTopic)
	if err != nil {
		logrus.Errorf("invalid mqtt topic '%s'", mqttTopic)
		return pk
	}
	if session != nil && len(pk.Filters) > 0 {
		// pass the subscription to the pubsub service
		pubsubTopic := service.MakeThingTopic(pubID, thingID, msgType, name)
		err = session.OnSubscribe(pubsubTopic, func(ev thing.ThingValue) {
			logrus.Infof("OnSubscribe. Received pubsub event on %s", pubsubTopic)
			newPk := pk //packets.NewPacket()
			mqttTopic = mqttclient.MakeTopic(ev.PublisherID, ev.ThingID, "event", ev.ID)
			evJson, _ := json.Marshal(ev)
			newPk.Payload = evJson
			newPk.FixedHeader.Type = packets.Publish
			newPk.TopicName = mqttTopic
			if err = cl.WritePacket(newPk); err != nil {
				err = fmt.Errorf("unable to write packet to client: %w", err)
				logrus.Error(err)
			}
		})
		if err != nil {
			err = fmt.Errorf("unable to subscribe to topic %s: %w", pubsubTopic, err)
			logrus.Error(err)
		}
	}
	//
	return pk
}

func (hook *GatewayHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnACLCheck,
		mqtt.OnConnect,
		mqtt.OnConnectAuthenticate,
		mqtt.OnDisconnect,
		mqtt.OnSubscribe,
		//mqtt.OnUnsubscribe,
		mqtt.OnPublish,
	}, []byte{b})
}

// NewMochiHook returns a new instance of the mochi-co mqtt server hook
//
//	serviceID is required mqtt-optional ID prefix used to listen on tcp/ws ports
func NewMochiHook(caCert *x509.Certificate) *GatewayHook {
	svc := &GatewayHook{
		caCert:       caCert,
		sessionMutex: sync.RWMutex{},
		sessions:     make(map[string]*MqttSession),
	}
	return svc
}
