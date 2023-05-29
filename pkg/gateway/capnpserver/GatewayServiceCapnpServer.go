package capnpserver

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"context"
	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/service"
	"github.com/sirupsen/logrus"
	"net"
)

// GatewayServiceCapnpServer provides the capnp RPC server for gateway capabilities
// This implements the capnproto generated interface GatewayService_Server

type GatewayServiceCapnpServer struct {
	svc *service.GatewayService
}

func (capsrv *GatewayServiceCapnpServer) NewSession(
	_ context.Context, call hubapi.CapGatewayService_newSession) error {
	sessionToken, _ := call.Args().SessionToken()
	clientID, _ := call.Args().ClientID()
	// get the session
	session, err := capsrv.svc.NewSession(clientID, sessionToken)
	if err != nil {
		return err
	}
	// wrap the session in a capnp server and hook into its HandleUnknownMethod forwarder
	capSessionSrv := NewGatewaySessionCapnpServer(session)

	// instead of using CapProvider_ServerToClient, this lets us hook into its HandleUnknownMethod handler
	//capability := hubapi.CapProvider_ServerToClient(capSessionSrv)
	c, _ := hubapi.CapProvider_Server(capSessionSrv).(server.Shutdowner)
	methods := hubapi.CapProvider_Methods(nil, capSessionSrv)
	clientHook := server.New(methods, capSessionSrv, c)
	clientHook.HandleUnknownMethod = capSessionSrv.HandleUnknownMethod

	resClient := capnp.NewClient(clientHook)
	capability := hubapi.CapProvider(resClient)

	// return the new session
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetSession(capability)
	}
	return err
}

func (capsrv *GatewayServiceCapnpServer) AuthNoAuth(
	_ context.Context, call hubapi.CapGatewayService_authNoAuth) error {
	clientID, _ := call.Args().ClientID()
	sessionToken := capsrv.svc.AuthNoAuth(clientID)
	res, err := call.AllocResults()
	if err != nil {
		return err
	}
	err = res.SetSessionToken(sessionToken)
	return err
}

func (capsrv *GatewayServiceCapnpServer) AuthProxy(
	_ context.Context, call hubapi.CapGatewayService_authProxy) error {

	clientID, _ := call.Args().ClientID()
	clientCertPEM, _ := call.Args().ClientCertPEM()
	sessionToken, err := capsrv.svc.AuthProxy(nil, clientID, clientCertPEM)
	if err != nil {
		return err
	}
	res, err := call.AllocResults()
	if err != nil {
		return err
	}
	err = res.SetSessionToken(sessionToken)
	return err
}

func (capsrv *GatewayServiceCapnpServer) AuthRefresh(
	_ context.Context, call hubapi.CapGatewayService_authRefresh) error {
	clientID, _ := call.Args().ClientID()
	oldSessionToken, _ := call.Args().SessionToken()
	sessionToken, err := capsrv.svc.AuthRefresh(clientID, oldSessionToken)
	if err != nil {
		return err
	}
	res, err := call.AllocResults()
	if err != nil {
		return err
	}
	err = res.SetSessionToken(sessionToken)
	return err
}

func (capsrv *GatewayServiceCapnpServer) AuthWithCert(
	_ context.Context, call hubapi.CapGatewayService_authWithCert) error {
	sessionToken, err := capsrv.svc.AuthWithCert(nil)
	if err != nil {
		return err
	}
	res, err := call.AllocResults()
	if err != nil {
		return err
	}
	err = res.SetSessionToken(sessionToken)
	return err
}

func (capsrv *GatewayServiceCapnpServer) AuthWithPassword(
	_ context.Context, call hubapi.CapGatewayService_authWithPassword) error {
	clientID, _ := call.Args().ClientID()
	password, _ := call.Args().Password()
	sessionToken, err := capsrv.svc.AuthWithPassword(clientID, password)
	if err != nil {
		return err
	}
	res, err := call.AllocResults()
	if err != nil {
		return err
	}
	err = res.SetSessionToken(sessionToken)
	return err
}

// StartGatewayServiceCapnpServer starts the capnp protocol server for the gateway service
func StartGatewayServiceCapnpServer(
	svc *service.GatewayService, lis net.Listener, wssPath string) (err error) {

	serviceName := gateway.ServiceName

	if wssPath != "" {
		logrus.Infof("listening on Websocket address %s%s", lis.Addr(), wssPath)
	} else {
		logrus.Infof("listening on TCP address %s", lis.Addr())
	}

	capsrv := &GatewayServiceCapnpServer{
		svc: svc,
	}
	boot := capnp.Client(hubapi.CapGatewayService_ServerToClient(capsrv))
	if wssPath != "" {
		err = listener.ServeWS(serviceName, lis, wssPath, boot, nil, nil)
	} else {
		err = listener.Serve(serviceName, lis, boot, nil, nil)
	}
	return err
}
