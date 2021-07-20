package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/signing"
	"github.com/wostzone/wostlib-go/pkg/tlsclient"
)

// commandline commands
const CmdBundle = "bundle"
const CmdClient = "clientcert"

// Commandline utility to manage WoST certificates.
func main() {
	var err error
	binFolder := path.Dir(os.Args[0])
	homeFolder := path.Dir(binFolder)
	certsFolder := path.Join(homeFolder, "certs")
	ouRole := certsetup.OUClient
	genKeys := false
	ifName, mac, ip := tlsclient.GetOutboundInterface("")
	_ = ifName
	_ = mac
	san := ip.String()

	bundleCommand := flag.NewFlagSet(CmdBundle, flag.ExitOnError)
	bundleCommand.StringVar(&certsFolder, "-certsFolder", certsFolder, "Certificates directory")
	bundleCommand.StringVar(&san, "-san", san, "Subject name or IP address to use in hub certificate. Default interface "+ifName)

	clientCertCommand := flag.NewFlagSet("clientcert", flag.ExitOnError)
	clientCertCommand.StringVar(&certsFolder, "-certsFolder", certsFolder, "Directory with CA cert and key")
	clientCertCommand.StringVar(&ouRole, "-ou", ouRole, "OU the client belongs to. This affects permissions.")
	clientCertCommand.BoolVar(&genKeys, "genkeys", false, "Generate the ECDSA public/private key pair")

	if len(os.Args) < 2 {
		{
			fmt.Printf("Usage: \n")
			fmt.Printf("  gencerts %s [OPTIONS]\n", CmdBundle)
			fmt.Printf("  gencerts %s clientID pubKey [OPTIONS]\n", CmdClient)
			fmt.Printf("\n")
			fmt.Printf("OPTIONS for %s\n", CmdBundle)
			bundleCommand.PrintDefaults()
			fmt.Printf("OPTIONS for %s clientID publicKeyFile\n", CmdClient)
			clientCertCommand.PrintDefaults()
			os.Exit(1)
		}
	}
	switch os.Args[1] {
	case CmdBundle:
		err = bundleCommand.Parse(os.Args[2:])
		if err != nil {
			os.Exit(1)
		} else if len(os.Args) > 2 {
			err = fmt.Errorf("too many arguments")
		} else {
			err = handleBundleCommand(certsFolder, san)
		}
	case CmdClient:
		// FIXME: better way to input public key
		err = clientCertCommand.Parse(os.Args[2:])
		cmdArgs := clientCertCommand.Args()
		if err != nil {
			os.Exit(1)
		} else if genKeys && len(cmdArgs) == 1 {
			// clientID with genkeys option
			clientID := cmdArgs[0]
			privKey := signing.CreateECDSAKeys()
			privKeyFile := clientID + "-key.pem"
			pubKeyFile := clientID + "-pub.pem"
			err = signing.SavePrivateKeyToPEM(privKey, privKeyFile)
			if err == nil {
				pubKeyPem, _ := signing.PublicKeyToPEM(&privKey.PublicKey)
				err = ioutil.WriteFile(pubKeyFile, []byte(pubKeyPem), 0644)
			}
			if err != nil {
				fmt.Printf("Failed saving keys: %s", err)
			}
			if err == nil {
				fmt.Printf("Saved client keys at: %s\n", privKeyFile)
				err = handleClientCertCommand(certsFolder, clientID, pubKeyFile)
			}
		} else if len(cmdArgs) <= 1 {
			err = fmt.Errorf("too few arguments. Expected: clientID publicKeyFile")
		} else {
			clientID := os.Args[2]
			pubKeyFile := os.Args[3]
			if err != nil {
				err = fmt.Errorf("%s: error reading public key file '%s': %s", CmdClient, pubKeyFile, err)
			} else {
				err = handleClientCertCommand(certsFolder, clientID, pubKeyFile)
			}
		}
	default:
		err = fmt.Errorf("invalid command %s", os.Args[1])
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func handleBundleCommand(certsFolder string, sanName string) error {
	// sanNames := []string{hc.MqttAddress}
	err := certsetup.CreateCertificateBundle([]string{sanName}, certsFolder)
	if err != nil {
		return err
	}
	println("Certificates generated in ", certsFolder)
	return nil
}

func handleClientCertCommand(certFolder string, clientID string, clientPubKeyFile string) error {
	ou := certsetup.OUClient
	caCertPEM, err := certsetup.LoadPEM(certFolder, certsetup.CaCertFile)
	if err != nil {
		return err
	}
	caKeyPEM, err := certsetup.LoadPEM(certFolder, certsetup.CaKeyFile)
	if err != nil {
		return err
	}
	pubKeyPEM, err := certsetup.LoadPEM("", clientPubKeyFile)
	if err != nil {
		return err
	}
	durationDays := certsetup.DefaultCertDurationDays
	certPEM, err := certsetup.CreateClientCert(
		clientID, ou, pubKeyPEM, caCertPEM, caKeyPEM, time.Now(), durationDays)
	if err != nil {
		return err
	}
	clientCertFile := clientID + "-cert.pem"
	certsetup.SaveCertToPEM(certPEM, "./", clientCertFile)

	fmt.Printf("Client certificate saved at %s\n", clientCertFile)
	return nil
}
