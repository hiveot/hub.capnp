package gatewaycli

import (
	"context"
	"crypto/tls"
	"fmt"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/config"
)

// load the hubcli key and certificate or use a temporary certificate to connect as hubcli service to the gateway
func connectToGateway(f svcconfig.AppFolders, gwAddr string) (gateway.IGatewaySession, error) {

	// load the CA
	caCertPath := path.Join(f.Certs, hubapi.DefaultCaCertFile)
	caKeyPath := path.Join(f.Certs, hubapi.DefaultCaKeyFile)
	caCert, err := certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Fatalf("unable to load the CA cert: %s", err)
	}
	// load the client cert
	clientKeyPath := path.Join(f.Certs, "hubcliKey.pem")
	clientCertPath := path.Join(f.Certs, "hubcliCert.pem")
	clientCert, err := certsclient.LoadTLSCertFromPEM(clientCertPath, clientKeyPath)

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
				nil, "hubcli", pubKeyPem, []string{"localhost", "127.0.0.1"}, 1)

			c, err3 := tls.X509KeyPair([]byte(clientCertPem), []byte(clientKeyPem))
			clientCert = &c
			err = err3
		}
	}
	if err != nil {
		logrus.Warningf("unable to load or create the hubcli client cert: %s", err)
	}
	gw, err := capnpclient.ConnectToGatewayTLS("tcp", gwAddr, clientCert, caCert)
	if err == nil {
		clientInfo, err2 := gw.Ping(context.Background())
		err = err2
		logrus.Infof("Connected to gateway as client='%s' clientType='%s'", clientInfo.ClientID, clientInfo.ClientType)
	}
	return gw, err
}

func GatewayCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	cmd := &cli.Command{
		Name:  "gw",
		Usage: "List gateway capabilities",
		Subcommands: []*cli.Command{
			GatewayListCommand(ctx, f),
		},
	}
	return cmd
}

func GatewayListCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List gateway capabilities",
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

func HandleListGateway(ctx context.Context, f svcconfig.AppFolders) error {

	gwConfig := config.NewGatewayConfig(f.Run, f.Certs)

	gw, err := connectToGateway(f, gwConfig.Address)
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
	return nil
}
