// Package certcli with certificate commandline definitions
package certscli

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/certs/capnpclient"
)

// CertCommands returns the certificate handling commands
func CertCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {

	cmd := &cli.Command{
		// certs ca | client | device --certs=folder --pubkey=path ID

		Name:  "certs",
		Usage: "Create certificates",
		Subcommands: []*cli.Command{
			CertCreateDeviceCommands(ctx, f),
			CertsCreateServiceCommand(ctx, f),
			CertsCreateUserCommand(ctx, f),
			CertsShowInfoCommand(ctx, f),
		},
	}
	return cmd
}

// CertsCreateUserCommand - requires the certs service to run
// hubcli certs client [--certs=CertFolder --pubkey=pubkeyfile] <loginID>
func CertsCreateUserCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	validityDays := certs.DefaultUserCertValidityDays

	return &cli.Command{
		Name:      "user",
		Usage:     "Create a user certificate",
		ArgsUsage: "<loginID>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "pubkey",
				Usage: "`file` with user public or private key in PEM format. When omitted a public/private key pair will be generated.",
			},
			&cli.IntFlag{
				Name:        "days",
				Usage:       "Number of days the certificate is valid.",
				Value:       validityDays,
				Destination: &validityDays,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() == 0 {
				return fmt.Errorf("Missing client login ID")
			}
			loginID := cCtx.Args().Get(0)
			pubKeyFile := cCtx.String("pubkey")
			err := HandleCreateUserCert(ctx, f, loginID, pubKeyFile, validityDays)
			return err
		},
	}
}

// CertCreateDeviceCommands
// hubcli certs device [--certs=CertFolder] --pubkey=pubkeyfile <deviceID>
func CertCreateDeviceCommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	validityDays := certs.DefaultDeviceCertValidityDays

	return &cli.Command{
		Name:      "device",
		Usage:     "Create an IoT device certificate",
		ArgsUsage: "<deviceID>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "pubkey",
				Usage: "`file` with device public or private key in PEM format. When omitted a public/private key pair will be generated.",
			},
			&cli.IntFlag{
				Name:        "days",
				Usage:       "Number of days the certificate is valid.",
				Value:       validityDays,
				Destination: &validityDays,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() == 0 {
				return fmt.Errorf("Missing device ID")
			}
			deviceID := cCtx.Args().Get(0)
			pubKeyFile := cCtx.String("pubkey")
			err := HandleCreateDeviceCert(ctx, f, deviceID, pubKeyFile, validityDays)
			return err
		},
	}
}

// CertsCreateServiceCommand
// hubcli certs service [--certs=CertFolder --pubkey=pubkeyfile] <serviceID>
func CertsCreateServiceCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	validityDays := certs.DefaultServiceCertValidityDays
	ipAddr := ""

	return &cli.Command{
		Name:      "service",
		Usage:     "Create a service certificate",
		ArgsUsage: "<serviceID>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "pubkey",
				Usage: "`file` with service public or private key in PEM format. When omitted a public/private key pair will be generated.",
			},
			&cli.StringFlag{
				Name:        "ipAddr",
				Usage:       "Optional service IP address in addition to localhost.",
				Destination: &ipAddr,
			},
			&cli.IntFlag{
				Name:        "days",
				Usage:       "Number of days the certificate is valid.",
				Value:       validityDays,
				Destination: &validityDays,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() == 0 {
				return fmt.Errorf("Missing service ID")
			}
			serviceID := cCtx.Args().Get(0)
			pubKeyFile := cCtx.String("pubkey")
			err := HandleCreateServiceCert(ctx, f, serviceID, ipAddr, pubKeyFile, validityDays)
			return err
		},
	}
}
func CertsShowInfoCommand(ctx context.Context, f svcconfig.AppFolders) *cli.Command {
	return &cli.Command{
		Name:      "info",
		Usage:     "Show certificate info",
		ArgsUsage: "<certFile>",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 1 {
				return fmt.Errorf("expected 1 argument. Got %d instead", cCtx.NArg())
			}
			return HandleShowCertInfo(ctx, cCtx.Args().First())
		},
	}
}

// HandleCreateDeviceCert creates an IoT device certificate and optionally private/public keypair
// This prints the certificate to stdout.
//
//	certFolder where to find the CA certificate and key used to sign the client certificate.
//	deviceID for the CN of the certificate. Used to identify the device.
//	keyFile with path to the client's public or private key
//	validity in days. 0 to use certconfig.DefaultClientCertDurationDays
func HandleCreateDeviceCert(ctx context.Context, f svcconfig.AppFolders, deviceID string, keyFile string, validityDays int) error {
	var pubKeyPEM string
	var generatedPrivKey *ecdsa.PrivateKey
	var certPEM string
	var cc certs.ICerts
	var dc certs.IDeviceCerts

	conn, err := listener.CreateLocalClientConnection(certs.ServiceName, f.Run)
	if err == nil {
		cc, err = capnpclient.NewCertServiceCapnpClient(conn)
	}
	if err == nil {
		dc = cc.CapDeviceCerts(ctx)
	}
	if err != nil {
		return err
	}
	if err == nil {
		pubKeyPEM, generatedPrivKey, err = loadOrCreateKey(keyFile)
	}
	// finally, create the user certificate
	if err == nil {
		certPEM, _, err = dc.CreateDeviceCert(ctx, deviceID, pubKeyPEM, validityDays)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Certificate for %s, valid for %d days:\n", deviceID, validityDays)
	fmt.Println(certPEM)
	if generatedPrivKey != nil {
		keyPem, _ := certsclient.PrivateKeyToPEM(generatedPrivKey)
		fmt.Println()
		fmt.Printf("Generated pub/private key pair:\n")
		fmt.Println(keyPem)
	}
	return err
}

// HandleCreateServiceCert creates a Hub service certificate and optionally private/public keypair
// This prints the certificate to stdout. The certificate is valid for localhost.
//
//	f.Certs where to find the CA certificate and key used to sign the certificate.
//	serviceID for the CN of the certificate. Used to identify the service.
//	ipAddr optional IP address in addition to localhost
//	keyFile with path to the client's public or private key
//	validity in days. 0 to use certconfig.DefaultClientCertDurationDays
func HandleCreateServiceCert(ctx context.Context, f svcconfig.AppFolders,
	serviceID string, ipAddr string, keyFile string, validityDays int) error {

	var names = []string{"localhost"}
	var pubKeyPEM string
	var generatedPrivKey *ecdsa.PrivateKey
	var certPEM string
	var cc certs.ICerts
	var sc certs.IServiceCerts

	conn, err := listener.CreateLocalClientConnection(certs.ServiceName, f.Run)
	if err == nil {
		cc, err = capnpclient.NewCertServiceCapnpClient(conn)
	}
	if err == nil {
		sc = cc.CapServiceCerts(ctx)
	}
	if err != nil {
		return err
	}
	if err == nil {
		pubKeyPEM, generatedPrivKey, err = loadOrCreateKey(keyFile)
	}
	// finally, create the user certificate
	if err == nil {
		certPEM, _, err = sc.CreateServiceCert(ctx, serviceID, pubKeyPEM, names, validityDays)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Certificate for %s, valid for %d days:\n", serviceID, validityDays)
	fmt.Println(certPEM)
	if generatedPrivKey != nil {
		keyPem, _ := certsclient.PrivateKeyToPEM(generatedPrivKey)
		fmt.Println()
		fmt.Printf("Generated pub/private key pair:\n")
		fmt.Println(keyPem)
	}
	return err
}

// HandleCreateUserCert creates a consumer client certificate and optionally private/public keypair
// This prints the certificate to stdout.
//
//	certFolder where to find the CA certificate and key used to sign the client certificate.
//	clientID for the CN of the client certificate. Used to identify the consumer.
//	keyFile with path to the client's public or private key
//	validity in days. 0 to use certconfig.DefaultClientCertDurationDays
func HandleCreateUserCert(ctx context.Context, f svcconfig.AppFolders, clientID string, keyFile string, validityDays int) error {
	var pubKeyPEM string
	var generatedPrivKey *ecdsa.PrivateKey
	var certPEM string
	var cc certs.ICerts
	var uc certs.IUserCerts

	conn, err := listener.CreateLocalClientConnection(certs.ServiceName, f.Run)
	if err == nil {
		cc, err = capnpclient.NewCertServiceCapnpClient(conn)
	}
	if err == nil {
		uc = cc.CapUserCerts(ctx)
	}
	if err != nil {
		return err
	}
	if err == nil {
		pubKeyPEM, generatedPrivKey, err = loadOrCreateKey(keyFile)
	}
	// finally, create the user certificate
	if err == nil {
		certPEM, _, err = uc.CreateUserCert(ctx, clientID, pubKeyPEM, validityDays)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Certificate for %s, valid for %d days:\n", clientID, validityDays)
	fmt.Println(certPEM)
	if generatedPrivKey != nil {
		keyPem, _ := certsclient.PrivateKeyToPEM(generatedPrivKey)
		fmt.Println()
		fmt.Printf("Generated pub/private key pair:\n")
		fmt.Println(keyPem)
	}
	return err
}

// HandleShowCertInfo shows certificate details
// Simplified version of openssl x509 -in cert -noout -text
//
//	certFile certificate to get info for
func HandleShowCertInfo(ctx context.Context, certFile string) error {
	cmd := exec.Command("openssl", "x509", "-in", certFile, "-noout", "-text")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("ERROR: %s.\n", err)
		fmt.Fprintf(os.Stderr, "%s", stderr.String())
	} else {
		fmt.Printf("%s\n", out)
	}
	return err
}

// Load public key or create a public/private key pair if not given.
// If the path is a private key, then extract the public key from it
func loadOrCreateKey(keyFile string) (
	pubKeyPEM string, generatedPrivKey *ecdsa.PrivateKey, err error) {
	var keyAsBytes []byte
	var pubKey *ecdsa.PublicKey

	// If a key file is given, use it, otherwise generate a pair
	if keyFile != "" {
		logrus.Infof("Using key file: %s\n", keyFile)
		keyAsBytes, err = ioutil.ReadFile(keyFile)
		if err != nil {
			logrus.Errorf("Failed loading Keyfile '%s': %s", keyFile, err)
			return "", nil, err
		}
		pubKeyPEM = string(keyAsBytes)

		// verify that this isn't a private key
		pubKey, err = certsclient.PublicKeyFromPEM(pubKeyPEM)
		if err != nil {
			logrus.Warningf("not a public key, try loading as private key...")
			privKey, err2 := certsclient.PrivateKeyFromPEM(pubKeyPEM)
			err = err2
			if err2 != nil {
				logrus.Errorf("Keyfile '%s' is a also not a private key: %s", keyFile, err2)
			} else {
				logrus.Infof("Keyfile '%s' is a private key", keyFile)
				pubKey = &privKey.PublicKey
				pubKeyPEM, err = certsclient.PublicKeyToPEM(pubKey)
			}
		}
		// error out if this is neither a public nor a private key
		if err != nil {
			return "", nil, err
		}
	} else {
		logrus.Info("No public key file was provided. Creating a new key pair.")
		generatedPrivKey = certsclient.CreateECDSAKeys()
		pubKey = &generatedPrivKey.PublicKey
		pubKeyPEM, err = certsclient.PublicKeyToPEM(pubKey)
	}
	return pubKeyPEM, generatedPrivKey, err
}
