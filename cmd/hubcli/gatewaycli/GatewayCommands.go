package gatewaycli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/config"
)

func loadCerts(f svcconfig.AppFolders) (clientCert *tls.Certificate, caCert *x509.Certificate) {
	var err error
	// load the CA
	caCertPath := path.Join(f.Certs, hubapi.DefaultCaCertFile)
	caKeyPath := path.Join(f.Certs, hubapi.DefaultCaKeyFile)
	caCert, err = certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Fatalf("unable to load the CA cert: %s", err)
	}
	// load the client cert
	clientKeyPath := path.Join(f.Certs, "hubcliKey.pem")
	clientCertPath := path.Join(f.Certs, "hubcliCert.pem")
	clientCert, err = certsclient.LoadTLSCertFromPEM(clientCertPath, clientKeyPath)

	// hubcli cert not found, create a temporary client cert
	if err != nil {
		caKey, err2 := certsclient.LoadKeysFromPEM(caKeyPath)
		err = err2
		// if no key is available then continue as unauthenticated client
		if err2 == nil {
			logrus.Warningf("Creating temporary client cert")
			clientKey := certsclient.CreateECDSAKeys()
			svc := selfsigned.NewServiceCertsService(caCert, caKey)
			pubKeyPem, _ := certsclient.PublicKeyToPEM(&clientKey.PublicKey)
			clientKeyPem, _ := certsclient.PrivateKeyToPEM(clientKey)
			clientCertPem, _, _ := svc.CreateServiceCert(
				context.TODO(), "hubcli", pubKeyPem, []string{"localhost", "127.0.0.1"}, 1)

			c, err3 := tls.X509KeyPair([]byte(clientCertPem), []byte(clientKeyPem))
			clientCert = &c
			err = err3
		}
	}
	if err != nil {
		logrus.Warningf("unable to load or create the hubcli client cert: %s", err)
	}
	return clientCert, caCert
}

func GatewayListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "listgw",
		Aliases:   []string{"lgw"},
		Usage:     "List gateway capabilities",
		Category:  "gateway",
		ArgsUsage: "(no args)",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListGateway(ctx, f)
			return err
		},
	}
}

// HandleListGateway handles list capabilities request and print the list of gateway capabilities
func HandleListGateway(ctx context.Context, f svcconfig.AppFolders) error {
	logrus.Infof("f.Config: %v", f.Config)
	var clientCert *tls.Certificate
	var caCert *x509.Certificate

	gwConfig := config.NewGatewayConfig()
	f.ConfigFile = path.Join(f.Config, "gateway.yaml")
	f.LoadConfig(&gwConfig)

	fullUrl := "tcp://" + gwConfig.Address
	if !gwConfig.NoWS {
		if gwConfig.NoTLS {
			fullUrl = "ws://" + gwConfig.WSAddress
		} else {
			fullUrl = "wss://" + gwConfig.WSAddress
		}
	}
	if !gwConfig.NoTLS {
		clientCert, caCert = loadCerts(f)
	}
	gw, err := capnpclient.ConnectToGateway(fullUrl, clientCert, caCert)
	if err != nil {
		return err
	}
	defer gw.Release()
	//logrus.Infof("Sending request")

	// ask as a service. we might want to make this a parameter
	capList, err := gw.ListCapabilities(ctx)
	fmt.Println("Capability                          Service                        ClientTypes")
	fmt.Println("--------                            -------                        ----       ")
	for _, capInfo := range capList {
		clientTypeAsText := strings.Join(capInfo.ClientTypes, ",")
		fmt.Printf("%-35s %-30s %-30s\n",
			capInfo.MethodName,
			capInfo.ServiceID,
			clientTypeAsText,
		)
	}
	return err
}
