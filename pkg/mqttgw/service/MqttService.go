package service

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hiveot/hub/lib/resolver"
	"github.com/hiveot/hub/pkg/directory"
	capnpclient3 "github.com/hiveot/hub/pkg/directory/capnpclient"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/history"
	capnpclient4 "github.com/hiveot/hub/pkg/history/capnpclient"
	"github.com/hiveot/hub/pkg/pubsub"
	capnpclient2 "github.com/hiveot/hub/pkg/pubsub/capnpclient"
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/listeners"
	"sync"
)

const serviceName = "mqttgw"

// MqttService hooks into the mochi-co mqttgw broker
type MqttService struct {
	mochiServer  *mqtt.Server
	mochiHook    *GatewayHook
	sessionMutex sync.RWMutex
	sessions     map[string]*MqttSession
	caCert       *x509.Certificate
	mux          sync.Mutex
}

// Start the mqttgw service.
// If unable to start, this exits with a Fatal message
//
// Precondition: the global resolver must be connected to the gateway or other service resolver.
//
//	mqttTcpPort and mqttWsPort are the listening ports for TCP and websocket connections
//	serverCert holds the TLS server certificate and key.
//	caCert holds the CA certificate used to generate the TLS cert.
//
// Returns an error if startup fails
func (svc *MqttService) Start(
	mqttTcpPort, mqttWsPort int, serverCert *tls.Certificate, caCert *x509.Certificate) error {

	svc.mux.Lock()
	defer svc.mux.Unlock()
	svc.caCert = caCert

	//srvOptions := &mqttgw.Options{Capabilities: &mqttgw.Capabilities{}}
	svc.mochiServer = mqtt.New(nil)

	// setup TLS listener for tcp and websocket
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{*serverCert},
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS13,
		RootCAs:      caCertPool,
		ServerName:   "HiveOT MQTT Gateway",
	}
	lisCfg := &listeners.Config{TLSConfig: &tlsConfig}
	if mqttTcpPort != 0 {
		tcplis := listeners.NewTCP(
			serviceName+"-tcp", fmt.Sprintf(":%d", mqttTcpPort), lisCfg)
		err := svc.mochiServer.AddListener(tcplis)
		if err != nil {
			return err
		}
	}
	// Create a WS listener on the given websocket port
	if mqttWsPort != 0 {
		wslis := listeners.NewWebsocket(serviceName+"-wss", fmt.Sprintf(":%d", mqttWsPort), lisCfg)
		err := svc.mochiServer.AddListener(wslis)
		if err != nil {
			return err
		}
	}
	// add the hiveot hook to manage client sessions and access control
	svc.mochiHook = NewMochiHook(svc.caCert)
	err := svc.mochiServer.AddHook(svc.mochiHook, map[string]any{})
	if err != nil {
		return err
	}
	//
	err = svc.mochiServer.Serve()
	return err
}

// Stop the mqttgw broker
func (svc *MqttService) Stop() error {
	svc.mux.Lock()
	defer svc.mux.Unlock()

	err := svc.mochiServer.Close()
	return err
}

// NewMqttGatewayService returns a new instance of the mqttgw gateway service.
// Use Start() to run.
//
//	serviceID is required mqttgw-optional ID prefix used to listen on tcp/ws ports
func NewMqttGatewayService() *MqttService {
	// initialize the global resolver with marshallers user by this service
	resolver.RegisterCapnpMarshaller[gateway.IGatewaySession](capnpclient.NewGatewaySessionCapnpClient, "")
	resolver.RegisterCapnpMarshaller[pubsub.IDevicePubSub](capnpclient2.NewDevicePubSubCapnpClient, "")
	resolver.RegisterCapnpMarshaller[pubsub.IUserPubSub](capnpclient2.NewUserPubSubCapnpClient, "")
	resolver.RegisterCapnpMarshaller[directory.IReadDirectory](capnpclient3.NewReadDirectoryCapnpClient, "")
	resolver.RegisterCapnpMarshaller[history.IReadHistory](capnpclient4.NewReadHistoryCapnpClient, "")

	svc := &MqttService{
		sessionMutex: sync.RWMutex{},
		sessions:     make(map[string]*MqttSession),
	}
	return svc
}
