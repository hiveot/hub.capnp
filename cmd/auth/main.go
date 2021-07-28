package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/wostzone/hub/pkg/aclstore"
	"github.com/wostzone/hub/pkg/auth"
	"github.com/wostzone/hub/pkg/unpwstore"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/signing"
	"github.com/wostzone/wostlib-go/pkg/tlsclient"
)

// commandline commands
const CmdCertBundle = "certbundle"
const CmdClientcert = "clientcert"
const CmdSetPasswd = "setpasswd"
const CmdSetRole = "setrole"
const Version = `0.1-alpha`

func main() {
	binFolder := path.Dir(os.Args[0])
	homeFolder := path.Dir(binFolder)
	ParseArgs(homeFolder, os.Args[1:])
}

// Commandline utility to manage authentication
func ParseArgs(homeFolder string, args []string) {
	// var err error
	configFolder := path.Join(homeFolder, "config")
	certsFolder := path.Join(homeFolder, "certs")
	// configFolder := path.Join(homeFolder, "config")
	// ouRole := certsetup.OUClient
	// genKeys := false
	ifName, mac, ip := tlsclient.GetOutboundInterface("")
	_ = ifName
	_ = mac
	sanName := ip.String()
	var optConf struct {
		// commands
		Certbundle  bool
		Clientcert  bool
		Setpassword bool
		Setrole     bool
		// arguments
		Loginid  string
		Groupid  string
		Password string
		Role     string
		// options
		Aclfile string
		Config  string
		Certs   string
		Output  string
		Pubkey  string
		Iter    int
	}
	usage := `
Usage:
  auth certbundle [--certs=CertFolder]
  auth clientcert [--certs=CertFolder --pubkey=pubkeyfile] <loginID> 
  auth setpassword [-c configFolder] [-i iterations] <loginID> [<password>]
  auth setrole [-c configFolder --aclfile=aclfile] <loginID> <groupID> <role>
  auth --help | --version

Commands:
  certbundle   Generate or refresh the Hub certificate bundle
  clientcert   Generate a signed client certificate with pub/private keys if not given
  setpassword  Set user password
  setrole      Set user role in group

Arguments:
  loginID      used as the certificate CN and login name 
  groupID      group for access control 
  outname      name of output certificate and pub/private key files if pubkey is not provided
  pubkeyFile   location of the client public key file
  role         one of viewer, editor, manager, thing, or none to delete

Options:
  --aclfile=AclFile              use a different acl file instead of the default config/` + aclstore.DefaultAclFilename + `
  -e --certs=CertFolder      location of Hub certificates [default: ` + certsFolder + `]
  -c --config=ConfigFolder   location of Hub config folder [default: ` + configFolder + `]
  -p --pubkey=PubKeyfile     use public key file instead of generating a key pair
	-i --iter=iterations       Number of iterations for generating password [default: 10]
  -h --help                  show this help
  --version                  show app version
`
	opts, err := docopt.ParseArgs(usage, args, Version)
	if err != nil {
		fmt.Printf("Parse Error: %s\n", err)
		os.Exit(1)
	}

	err = opts.Bind(&optConf)

	if err != nil {
		fmt.Printf("Bind Error: %s\n", err)
		os.Exit(1)
	}
	_ = opts
	if optConf.Certbundle {
		fmt.Printf("Generating Certificate Bundle. Certfolder=%s\n", optConf.Certs)
		err = HandleCreateCertbundle(optConf.Certs, sanName)
	} else if optConf.Clientcert {
		fmt.Printf("Generating Client Certificate using CA from %s\n", optConf.Certs)
		err = HandleCreateClientCert(optConf.Certs, optConf.Loginid, optConf.Pubkey)
	} else if optConf.Setpassword {
		fmt.Printf("Set user password\n")
		err = HandleSetPasswd(optConf.Config, optConf.Loginid, optConf.Password, optConf.Iter)
	} else if optConf.Setrole {
		fmt.Printf("Set user role in group\n")
		err = HandleSetRole(optConf.Config, optConf.Loginid, optConf.Groupid, optConf.Role, optConf.Aclfile)
	} else {
		err = fmt.Errorf("invalid command")
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

// CreateKeyPair generate a key pair in PEM format
// This saves the private and public key files as <clientID>-pub.pem and <clientID>-key.pem
// Returns the public key PEM content
func CreateKeyPair(clientID string) (pubKeyPem string, err error) {
	privKey := signing.CreateECDSAKeys()
	privKeyFile := clientID + "-priv.pem"
	pubKeyFile := clientID + "-pub.pem"
	err = signing.SavePrivateKeyToPEM(privKey, privKeyFile)
	if err == nil {
		pubKeyPem, _ = signing.PublicKeyToPEM(&privKey.PublicKey)
		err = ioutil.WriteFile(pubKeyFile, []byte(pubKeyPem), 0644)
	}
	if err != nil {
		fmt.Printf("Failed saving keys: %s", err)
	}
	if err == nil {
		fmt.Printf("Generated public and private key pair as: %s and %s\n", pubKeyFile, privKeyFile)
	}
	return pubKeyPem, err
}

// Set the login name and password for a consumer
func HandleSetPasswd(configFolder string, username string, passwd string, iterations int) error {
	var pwHash string
	var err error
	unpwFilePath := path.Join(configFolder, unpwstore.DefaultUnpwFilename)
	unpwStore := unpwstore.NewPasswordFileStore(unpwFilePath, "auth.main.HandleSetPasswd")
	err = unpwStore.Open()
	if err == nil {
		pwHash, err = auth.CreatePasswordHash(passwd, auth.PWHASH_ARGON2id, uint(iterations))
	}
	if err == nil {
		err = unpwStore.SetPasswordHash(username, pwHash)
	}
	if err == nil {
		unpwStore.Close()
		fmt.Printf("Password updated for user %s", username)
	}
	return err
}

// Generate the hub certificate bundle CA, Hub and Plugin keys and certificates
// If the CA certificate already exist it is NOT updated
// If the Hub and Plugin certificates already exist, they are renewed
func HandleCreateCertbundle(certsFolder string, sanName string) error {
	err := certsetup.CreateCertificateBundle([]string{sanName}, certsFolder)
	if err != nil {
		return err
	}
	fmt.Printf("Certificates generated in %s", certsFolder)
	return nil
}

// Create a consumer client certificate and optionall private/public keypair
//  certFolder where to find the CA certificate and key used to sign the client certificate
//  clientID for the CN of the client certificate. Used to identify the consumer.
//  clientPubKeyFile with the path to the public key of the client
func HandleCreateClientCert(certFolder string, clientID string, clientPubKeyFile string) error {
	var pubKeyPEM string
	ou := certsetup.OUClient
	caCertPEM, err := certsetup.LoadPEM(certFolder, certsetup.CaCertFile)
	if err != nil {
		return err
	}
	caKeyPEM, err := certsetup.LoadPEM(certFolder, certsetup.CaKeyFile)
	if err != nil {
		return err
	}
	// If a public key file is given, use it, otherwise generate a pair
	if clientPubKeyFile != "" {
		fmt.Printf("Using public key file: %s\n", clientPubKeyFile)
		pubKeyPEM, err = certsetup.LoadPEM("", clientPubKeyFile)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("No public key file was provided. ") // no newline
		pubKeyPEM, err = CreateKeyPair(clientID)
		if err != nil {
			return err
		}
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

// Create a device client certificate for thing devices
// This is similar to creating a consumer certificate
func HandleCreateThingCert(certFolder string, deviceID string, pubKeyFile string) error {
	const deviceCertValidityDays = 30
	caCertPEM, err := certsetup.LoadPEM(certFolder, certsetup.CaCertFile)
	if err != nil {
		return err
	}
	caKeyPEM, err := certsetup.LoadPEM(certFolder, certsetup.CaKeyFile)
	if err != nil {
		return err
	}
	pubKeyPEM, err := certsetup.LoadPEM("", pubKeyFile)
	if err != nil {
		return err
	}
	certPEM, err := certsetup.CreateClientCert(
		deviceID, certsetup.OUIoTDevice, pubKeyPEM,
		caCertPEM, caKeyPEM,
		time.Now().Add(-10*time.Second), deviceCertValidityDays)
	if err != nil {
		return err
	}
	clientCertFile := deviceID + "-cert.pem"
	certsetup.SaveCertToPEM(certPEM, "./", clientCertFile)

	fmt.Printf("Device certificate saved at %s\n", clientCertFile)
	return nil
}

// Set the role of a client in a group.
func HandleSetRole(configFolder string, clientID string, groupID string, role string, aclFile string) error {
	if role != auth.GroupRoleEditor && role != auth.GroupRoleViewer &&
		role != auth.GroupRoleManager && role != auth.GroupRoleThing && role != auth.GroupRoleNone {
		err := fmt.Errorf("invalid role '%s'", role)
		return err
	}
	aclFilePath := path.Join(configFolder, aclstore.DefaultAclFilename)
	if aclFile != "" {
		// option to specify an acl file wrt home
		aclFilePath = path.Join(path.Dir(configFolder), aclFile)
	}
	aclStore := aclstore.NewAclFileStore(aclFilePath, "auth.main.HandleSetRole")
	err := aclStore.Open()
	if err == nil {
		err = aclStore.SetRole(clientID, groupID, role)
	}
	if err == nil {
		fmt.Printf("Client '%s' role set to '%s' for group '%s'\n", clientID, role, groupID)
	}
	aclStore.Close()
	return err
}
