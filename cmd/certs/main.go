package main

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/sirupsen/logrus"

	"github.com/wostzone/hub/cmd/certs/certsetup"

	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/hubnet"
	"github.com/wostzone/wost-go/pkg/signing"
)

const Version = `0.3-alpha`

func main() {
	binFolder := path.Dir(os.Args[0])
	homeFolder := path.Dir(binFolder)
	ParseArgs(homeFolder, os.Args[1:])
}

// ParseArgs to handle commandline arguments
func ParseArgs(homeFolder string, args []string) {
	// var err error
	configFolder := path.Join(homeFolder, "config")
	certsFolder := path.Join(homeFolder, "certs")
	// configFolder := path.Join(homeFolder, "config")
	// ouRole := certsetup.OUClient
	// genKeys := false
	ifName, mac, ip := hubnet.GetOutboundInterface("")
	_ = ifName
	_ = mac
	sanName := ip.String()
	var optConf struct {
		// commands
		Certbundle bool
		Clientcert bool
		Devicecert bool
		// arguments
		Loginid  string
		Deviceid string
		// options
		Config   string
		Certs    string
		Hostname string
		Output   string
		Pubkey   string
		Iter     int
		Verbose  bool
	}

	usage := `
Usage:
  certs certbundle [-v --certs=CertFolder] [--hostname=hostname]
  certs clientcert [-v --certs=CertFolder --pubkey=pubkeyfile] <loginID> 
  certs devicecert [-v --certs=CertFolder --pubkey=pubkeyfile] <deviceID> 
  certs --help | --version

Commands:
  certbundle   Generate or refresh the Hub certificate bundle, optionally provide a subject hostname or ip
  clientcert   Generate a signed client certificate, with pub/private keys if not given
  devicecert   Generate a signed device certificate, with pub/private keys if not given

Arguments:
  loginID      used as the client certificate CN and filename (loginID-cert.pem)

Options:
  -e --certs=CertFolder      location of Hub certificates [default: ` + certsFolder + `]
  -c --config=ConfigFolder   location of Hub config folder [default: ` + configFolder + `]
  -n --hostname=Hostname     hostname or IP address to use on the certificate. Default is outbound IP
  -p --pubkey=PubKeyfile     use this public key file to generate certificate, instead of a new key pair
  -h --help                  show this help
  -v --verbose               show info logging
`
	opts, err := docopt.ParseArgs(usage, args, Version)
	if err != nil {
		fmt.Printf("Parse Error: %s\n", err)
		os.Exit(1)
	}

	err = opts.Bind(&optConf)

	if optConf.Verbose {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	if err != nil {
		fmt.Printf("Bind Error: %s\n", err)
		os.Exit(1)
	}
	_ = opts
	if optConf.Certbundle {
		if optConf.Hostname != "" {
			sanName = optConf.Hostname
		}
		fmt.Printf("Generating certificate Bundle. Certfolder=%s. names=%s\n", optConf.Certs, sanName)
		err = HandleCreateCertbundle(optConf.Certs, sanName)
	} else if optConf.Clientcert {
		fmt.Printf("Generating Client certificate using CA from %s\n", optConf.Certs)
		err = HandleCreateClientCert(optConf.Certs, optConf.Loginid, optConf.Pubkey)
	} else if optConf.Devicecert {
		fmt.Printf("Generating Thing device certificate using CA from %s\n", optConf.Certs)
		err = HandleCreateDeviceCert(optConf.Certs, optConf.Deviceid, optConf.Pubkey)
	} else {
		err = fmt.Errorf("invalid command")
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

// CreateKeyPair generate a key pair in PEM format and save it to the cert folder
// as <clientID>-pub.pem and <clientID>-key.pem
// Returns the public key PEM content
func CreateKeyPair(clientID string, certFolder string) (privKey *ecdsa.PrivateKey, err error) {
	privKey = signing.CreateECDSAKeys()
	privKeyFile := path.Join(certFolder, clientID+"-priv.pem")
	pubKeyFile := path.Join(certFolder, clientID+"-pub.pem")
	err = certsclient.SaveKeysToPEM(privKey, privKeyFile)
	if err == nil {
		pubKeyPem, _ := certsclient.PublicKeyToPEM(&privKey.PublicKey)
		err = ioutil.WriteFile(pubKeyFile, []byte(pubKeyPem), 0644)
	}
	if err != nil {
		fmt.Printf("Failed saving keys: %s\n", err)
	}
	if err == nil {
		fmt.Printf("Generated public and private key pair as: %s and %s\n", pubKeyFile, privKeyFile)
	}
	return privKey, err
}

// HandleCreateCertbundle generates the hub certificate bundle CA, Hub and Plugin keys
// and certificates.
//  If the CA certificate already exist it is NOT updated
//  If the Hub and Plugin certificates already exist, they are renewed
func HandleCreateCertbundle(certsFolder string, sanName string) error {
	err := certsetup.CreateCertificateBundle([]string{sanName}, certsFolder, true)
	if err != nil {
		return err
	}
	fmt.Printf("Server and Plugin certificates generated in %s\n", certsFolder)
	return nil
}

// HandleCreateClientCert creates a consumer client certificate and optionally private/public keypair
//  certFolder where to find the CA certificate and key used to sign the client certificate
//  clientID for the CN of the client certificate. Used to identify the consumer.
//  pubKeyFile with path to the client's public key of the certificate
func HandleCreateClientCert(certFolder string, clientID string, pubKeyFile string) error {
	var pubKey *ecdsa.PublicKey
	ou := certsclient.OUClient
	pemPath := path.Join(certFolder, config.DefaultCaCertFile)
	caCert, err := certsclient.LoadX509CertFromPEM(pemPath)
	if err != nil {
		return err
	}
	pemPath = path.Join(certFolder, config.DefaultCaKeyFile)
	caKey, err := certsclient.LoadKeysFromPEM(pemPath)
	if err != nil {
		return err
	}
	// If a public key file is given, use it, otherwise generate a pair
	if pubKeyFile != "" {
		fmt.Printf("Using public key file: %s\n", pubKeyFile)
		pubKey, err = certsclient.LoadPublicKeyFromPEM(pubKeyFile)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("No public key file was provided. Creating a key pair ") // no newline
		privKey, err := CreateKeyPair(clientID, "")
		pubKey = &privKey.PublicKey
		if err != nil {
			return err
		}
	}
	durationDays := certsetup.DefaultCertDurationDays
	cert, err := certsetup.CreateHubClientCert(
		clientID, ou, pubKey, caCert, caKey, time.Now(), durationDays)
	if err != nil {
		return err
	}
	pemPath = path.Join(".", clientID+"-cert.pem")
	err = certsclient.SaveX509CertToPEM(cert, pemPath)

	fmt.Printf("Client certificate saved at %s\n", pemPath)
	return err
}

// HandleCreateDeviceCert creates a device client certificate for a device and save it in the certFolder
// This is similar to creating a consumer certificate
func HandleCreateDeviceCert(certFolder string, deviceID string, pubKeyFile string) error {
	const deviceCertValidityDays = 30
	var pubKey *ecdsa.PublicKey
	pemPath := path.Join(certFolder, config.DefaultCaCertFile)
	caCert, err := certsclient.LoadX509CertFromPEM(pemPath)
	if err != nil {
		return err
	}
	pemPath = path.Join(certFolder, config.DefaultCaKeyFile)
	caKey, err := certsclient.LoadKeysFromPEM(pemPath)
	if err != nil {
		return err
	}
	// If a public key file is given, use it, otherwise generate a pair
	if pubKeyFile != "" {
		fmt.Printf("Using public key file: %s\n", pubKeyFile)
		pubKey, err = certsclient.LoadPublicKeyFromPEM(pubKeyFile)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("No public key file was provided. Creating a new key pair ") // no newline
		privKey, err := CreateKeyPair(deviceID, "")
		pubKey = &privKey.PublicKey
		if err != nil {
			return err
		}
	}

	certPEM, err := certsetup.CreateHubClientCert(
		deviceID, certsclient.OUIoTDevice, pubKey,
		caCert, caKey,
		time.Now(), deviceCertValidityDays)
	if err != nil {
		return err
	}
	// save the new certificate
	pemPath = path.Join(".", deviceID+"-cert.pem")
	err = certsclient.SaveX509CertToPEM(certPEM, pemPath)

	fmt.Printf("Device certificate saved at %s\n", pemPath)
	return err
}
