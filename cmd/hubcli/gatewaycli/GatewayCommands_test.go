package gatewaycli

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/cmd/hubcli/certscli"
	"github.com/hiveot/hub/internal/dummy"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
	"github.com/hiveot/hub/pkg/resolver"
)

const testHomeDir = "/tmp/test-hubcli"

var testSocketDir = path.Join(testHomeDir, "run")
var resolverSocketPath = path.Join(testSocketDir, resolver.ServiceName+".socket")

func TestConnectToGateway(t *testing.T) {
	logging.SetLogging("info", "")
	_ = os.RemoveAll(testHomeDir)
	err := os.MkdirAll(testHomeDir, 0700)
	assert.NoError(t, err)

	// step1: generate a CA cert for testing
	f := svcconfig.GetFolders(testHomeDir, false)
	err = certscli.HandleCreateCACert(f.Certs, 1, true)
	assert.NoError(t, err)

	//testCACert, testCAKey, err := selfsigned.CreateHubCA(1)
	caCertPath := path.Join(f.Certs, hubapi.DefaultCaCertFile)
	caKeyPath := path.Join(f.Certs, hubapi.DefaultCaKeyFile)
	testCAKey, err2 := certsclient.LoadKeysFromPEM(caKeyPath)
	testCACert, err := certsclient.LoadX509CertFromPEM(caCertPath)
	require.NoError(t, err)

	// step 2: generate the gateway server cert
	certSvc := selfsigned.NewSelfSignedCertsService(testCACert, testCAKey)
	capServiceCert, err := certSvc.CapServiceCerts(context.Background(), "hubcli")
	testServiceKeys := certsclient.CreateECDSAKeys()
	testServicePubKeyPEM, _ := certsclient.PublicKeyToPEM(&testServiceKeys.PublicKey)
	testServicePrivKeyPEM, _ := certsclient.PrivateKeyToPEM(testServiceKeys)
	testServiceCertPEM, _, err := capServiceCert.CreateServiceCert(
		context.Background(), "hubcli-test", testServicePubKeyPEM, []string{"localhost", "127.0.0.1"}, 1)
	testServiceCert, err := tls.X509KeyPair([]byte(testServiceCertPEM), []byte(testServicePrivKeyPEM))
	require.NoError(t, err)

	// step 3: start the gateway
	authnService := dummy.NewDummyAuthnService()
	svc := service.NewGatewayService(resolverSocketPath, authnService)
	err = svc.Start()
	require.NoError(t, err)

	srvListener, err2 := net.Listen("tcp", "127.0.0.1:0")
	if err2 != nil {
		logrus.Panicf("Unable to create a listener, can't run test: %s", err2)
	}
	srvListener = listener.CreateTLSListener(srvListener, &testServiceCert, testCACert)
	go capnpserver.StartGatewayCapnpServer(svc, srvListener)

	// step 4: client connects
	gw, err := connectToGateway(f, srvListener.Addr().String())

	assert.NoError(t, err)
	assert.NotEmpty(t, gw)
	time.Sleep(time.Second)
	clientInfo, err := gw.Ping(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, clientInfo)

	// shutdown
	err = srvListener.Close()
	assert.NoError(t, err)
	gw.Release()

}
