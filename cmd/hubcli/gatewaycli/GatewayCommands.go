package gatewaycli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hiveot/hub/lib/hubclient"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/config"
)

func loadCerts(certsFolder string) (clientCert *tls.Certificate, caCert *x509.Certificate) {
	var err error
	// load the CA
	fmt.Println("certfolder = ", certsFolder)
	caCertPath := path.Join(certsFolder, hubapi.DefaultCaCertFile)
	caKeyPath := path.Join(certsFolder, hubapi.DefaultCaKeyFile)
	caCert, err = certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Fatalf("unable to load the CA cert: %s", err)
	}
	// load the client cert
	clientKeyPath := path.Join(certsFolder, "hubcliKey.pem")
	clientCertPath := path.Join(certsFolder, "hubcliCert.pem")
	clientCert, err = certsclient.LoadTLSCertFromPEM(clientCertPath, clientKeyPath)

	// hubcli cert not found, create a temporary client cert
	if err != nil {
		caKey, err2 := certsclient.LoadKeysFromPEM(caKeyPath)
		err = err2
		// if no key is available then continue as unauthenticated client
		if err2 == nil {
			logrus.Warningf("CLI cert not found. Creating temporary client cert")
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

func GatewayListCommand(ctx context.Context, certsFolder *string, configFolder *string) *cli.Command {
	var fullURL = ""
	var duration = 1
	return &cli.Command{
		Name:     "lgw",
		Usage:    "List gateway capabilities",
		Category: "gateway",
		//ArgsUsage: "(no args)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "url",
				Usage:       "full URL of gateway. ",
				Value:       "",
				Destination: &fullURL,
			},
			&cli.IntFlag{
				Name:        "duration",
				Usage:       "Search duration for auto discovery",
				Value:       duration,
				Destination: &duration,
			},
		},
		Action: func(cCtx *cli.Context) error {

			if cCtx.NArg() != 0 {
				return fmt.Errorf("no arguments expected")
			}
			err := HandleListGateway(ctx, fullURL, duration, certsFolder, configFolder)
			return err
		},
	}
}

// HandleListGateway handles list capabilities request and print the list of gateway capabilities
func HandleListGateway(ctx context.Context, fullURL string, searchTimeSec int, certsFolder *string, configFolder *string) error {

	var clientCert *tls.Certificate
	var caCert *x509.Certificate
	gwConfig := config.NewGatewayConfig()
	configFile := path.Join(*configFolder, "gateway.yaml")
	cfgData, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(cfgData, &gwConfig)
	if err != nil {
		return err
	}
	if !gwConfig.NoTLS {
		clientCert, caCert = loadCerts(*certsFolder)
	}
	//fullUrl := fmt.Sprintf("%s:%d", gwConfig.Address, gwConfig.TcpPort)
	//fullUrl := fmt.Sprintf("wss://%s:%d%s", gwConfig.Address, gwConfig.WssPort, gwConfig.WssPath)
	capClient, err := hubclient.ConnectWithCapnpTCP(fullURL, clientCert, caCert)
	gw := capnpclient.NewGatewaySessionCapnpClient(capClient)
	if err != nil {
		return err
	}
	defer gw.Release()

	// ask as a service. we might want to make this a parameter
	capList, err := gw.ListCapabilities(ctx)
	fmt.Println("Capability                          Service                        AuthTypes")
	fmt.Println("--------                            -------                        ----       ")
	for _, capInfo := range capList {
		authTypeAsText := strings.Join(capInfo.AuthTypes, ",")
		fmt.Printf("%-35s %-30s %-30s\n",
			capInfo.MethodName,
			capInfo.ServiceID,
			authTypeAsText,
		)
	}
	return err
}
