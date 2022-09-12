// Package certcli with certificate handling commands
package certcli

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"time"

	"capnproto.org/go/capnp/v3/rpc"
	svc2 "github.com/hiveot/hub.capnp/go/capnp/svc"
	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/svc/certsvc/selfsigned"
	"github.com/hiveot/hub/pkg/svc/certsvc/service"
)

func loadCA(certFolder string) (caCert *x509.Certificate, caKey *ecdsa.PrivateKey, err error) {
	pemPath := path.Join(certFolder, service.DefaultCaCertFile)
	caCert, err = certsclient.LoadX509CertFromPEM(pemPath)
	if err == nil {
		pemPath = path.Join(certFolder, service.DefaultCaKeyFile)
		caKey, err = certsclient.LoadKeysFromPEM(pemPath)
	}
	if err != nil {
		return nil, nil, err
	}
	return caCert, caKey, nil
}

// Load public key or create a public/private key pair if not given.
// If the path is a private key, then extract the public key from it
func loadOrCreateKey(keyFile string) (pubKey *ecdsa.PublicKey, generatedPrivKey *ecdsa.PrivateKey, err error) {
	// If a key file is given, use it, otherwise generate a pair
	if keyFile != "" {
		logrus.Infof("Using key file: %s\n", keyFile)
		pubKey, err = certsclient.LoadPublicKeyFromPEM(keyFile)
		// maybe this is a private key
		if err != nil {
			//fmt.Println("not a public key, try loading as private key...")
			privKey, err2 := certsclient.LoadKeysFromPEM(keyFile)
			if err2 == nil {
				logrus.Infof("Keyfile '%s' is a private key", keyFile)
				pubKey = &privKey.PublicKey
				err = nil
			}
		}
		// error out if this is neither a public nor a private key
		if err != nil {
			return nil, nil, err
		}
	} else {
		logrus.Info("No public key file was provided. Creating a new key pair.")
		generatedPrivKey = certsclient.CreateECDSAKeys()
		pubKey = &generatedPrivKey.PublicKey
	}
	return
}

// HandleCreateCACert generates the hub self-signed CA private key and certificate
// in the given folder.
// Use force to create the folder and overwrite existing certificate if it exists
func HandleCreateCACert(certsFolder string, sanName string, validityDays int, force bool) error {
	caCertPath := path.Join(certsFolder, service.DefaultCaCertFile)
	caKeyPath := path.Join(certsFolder, service.DefaultCaKeyFile)

	// folder doesn't exist
	if _, err := os.Stat(certsFolder); err != nil {
		if force {
			os.Mkdir(certsFolder, 0744)
		} else {
			return fmt.Errorf("certificate folder '%s' doesn't exist", certsFolder)
		}
	}
	// do not overwrite existing certificate unless force is used
	if !force {
		if _, err := os.Stat(caCertPath); err == nil {
			return fmt.Errorf("CA certificate already exists in '%s'", caCertPath)
		}
		if _, err := os.Stat(caKeyPath); err == nil {
			return fmt.Errorf("CA key alread exists in '%s'", caKeyPath)
		}
	}

	caCert, privKey, err := selfsigned.CreateHubCA(validityDays)
	if err != nil {
		return err
	}
	err = certsclient.SaveX509CertToPEM(caCert, caCertPath)
	if err == nil {
		// this sets permissions to 0400 current user readonly
		err = certsclient.SaveKeysToPEM(privKey, caKeyPath)
	}

	logrus.Infof("Generated CA certificate '%s' and key '%s'\n", caCertPath, caKeyPath)
	return err
}

// HandleCreateClientCert creates a consumer client certificate and optionally private/public keypair
// This prints the certificate to stdout.
//
//  certFolder where to find the CA certificate and key used to sign the client certificate.
//  clientID for the CN of the client certificate. Used to identify the consumer.
//  keyFile with path to the client's public or private key
//  validity in days. 0 to use certconfig.DefaultClientCertDurationDays
//func HandleCreateClientCert(certFolder string, clientID string, keyFile string, validityDays int) error {
//	var pubKey *ecdsa.PublicKey
//	var generatedPrivKey *ecdsa.PrivateKey
//	var cert *x509.Certificate
//
//	if validityDays == 0 {
//		validityDays = service.DefaultClientCertDurationDays
//	}
//	caCert, caKey, err := loadCA(certFolder)
//	if err == nil {
//		pubKey, generatedPrivKey, err = loadOrCreateKey(keyFile)
//	}
//	if err == nil {
//		cert, err = selfsigned.CreateClientCert(clientID, certsclient.OUClient, pubKey, caCert, caKey, validityDays)
//	}
//	if err != nil {
//		return err
//	}
//	certPem := certsclient.X509CertToPEM(cert)
//	fmt.Printf("Certificate for %s, valid for %d days:\n", clientID, validityDays)
//	fmt.Println(certPem)
//	if generatedPrivKey != nil {
//		keyPem, _ := certsclient.PrivateKeyToPEM(generatedPrivKey)
//		fmt.Println()
//		fmt.Printf("Generated pub/private key pair:\n")
//		fmt.Println(keyPem)
//	}
//	return err
//}

// HandleCreateClientCert creates a consumer client certificate and optionally private/public
// keypair through the service via Capnp protocol.
// This prints the certificate to stdout.
//
//  clientID for the CN of the client certificate. Used to identify the consumer.
//  keyFile with path to the client's public or private key
//  validity in days. 0 to use certconfig.DefaultClientCertDurationDays
func HandleCreateClientCert(clientID string, keyFile string, validityDays int) error {
	// Set up a connection to the server.
	// for UDS need the schema prefix: https://github.com/grpc/grpc-go/issues/1846
	// address := "unix:///tmp/certsvc.socket"

	// fmt.Println("Connecting to: ", address)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	// defer cancel()
	// conn, err := grpc.DialContext(ctx, address,
	// 	grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	// if err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
	// defer conn.Close()

	network := "unix"
	address := "/tmp/certsvc.socket"
	clientSideConn, err := net.Dial(network, address)
	transport := rpc.NewStreamTransport(clientSideConn)
	clientConn := rpc.NewConn(transport, nil)
	defer clientConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	certClient := svc2.CertServiceCap(clientConn.Bootstrap(ctx))

	pubKey, generatedPrivKey, err := loadOrCreateKey(keyFile)
	if err != nil {
		log.Fatal(err)
	}
	pubKeyPEM, err := certsclient.PublicKeyToPEM(pubKey)
	if err != nil {
		log.Fatal(err)
	}
	// create the client capability (eg message)
	resp, release := certClient.CreateClientCert(ctx,
		func(params svc2.CertServiceCap_createClientCert_Params) error {
			fmt.Println("CertServiceCap_createClientCert_Params")
			err = params.SetClientID(clientID)
			err = params.SetPubKeyPEM(pubKeyPEM)
			return err
		})
	defer release()

	result, err := resp.Struct()
	if err != nil {
		log.Fatalf("error getting response struct: %v", err)
	}
	fmt.Printf("Certificate for %s, valid for %d days:\n", clientID, validityDays)
	fmt.Println(result.CertPEM())

	if generatedPrivKey != nil {
		keyPem, _ := certsclient.PrivateKeyToPEM(generatedPrivKey)
		fmt.Println()
		fmt.Printf("Generated pub/private key pair:\n")
		fmt.Println(keyPem)
	}
	return err
}

// HandleCreateDeviceCert creates a device client certificate for a device and save it in the certFolder
// This is similar to creating a consumer certificate
func HandleCreateDeviceCert(certFolder string, deviceID string, keyFile string, validityDays int) error {
	var pubKey *ecdsa.PublicKey
	var generatedPrivKey *ecdsa.PrivateKey
	var cert *x509.Certificate

	if validityDays == 0 {
		validityDays = service.DefaultDeviceCertDurationDays
	}
	caCert, caKey, err := loadCA(certFolder)
	if err == nil {
		pubKey, generatedPrivKey, err = loadOrCreateKey(keyFile)
	}
	if err == nil {
		cert, err = selfsigned.CreateClientCert(deviceID, certsclient.OUIoTDevice, pubKey, caCert, caKey, validityDays)
	}
	if err != nil {
		return err
	}
	certPem := certsclient.X509CertToPEM(cert)
	fmt.Printf("Device certificate for %s, valid for %d days:\n", deviceID, validityDays)
	fmt.Println(certPem)
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
//  certFolder where to find the CA certificate and key used to sign the certificate.
//  serviceID for the CN of the certificate. Used to identify the service.
//  ipAddr optional IP address in addition to localhost
//  keyFile with path to the public or private key of the service, if exists.
//  validity in days. 0 to use certconfig.DefaultServiceCertDurationDays
func HandleCreateServiceCert(certFolder string, serviceID string, ipAddr string, keyFile string, validityDays int) error {
	var names = []string{"localhost"}
	var pubKey *ecdsa.PublicKey
	var generatedPrivKey *ecdsa.PrivateKey
	var cert *x509.Certificate

	if ipAddr != "" {
		names = append(names, ipAddr)
	}
	if validityDays == 0 {
		validityDays = service.DefaultServiceCertDurationDays
	}
	caCert, caKey, err := loadCA(certFolder)
	if err == nil {
		pubKey, generatedPrivKey, err = loadOrCreateKey(keyFile)
	}
	if err == nil {
		cert, err = selfsigned.CreateServiceCert(serviceID, names, pubKey, caCert, caKey, validityDays)
	}
	if err != nil {
		return err
	}
	certPem := certsclient.X509CertToPEM(cert)
	fmt.Printf("Certificate for %s, valid for %d days:\n", serviceID, validityDays)
	fmt.Println(certPem)
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
//  certFile certificate to get info for
func HandleShowCertInfo(certFile string) error {
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
