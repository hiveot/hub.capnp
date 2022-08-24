package certcli

import (
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"os"
	"path"

	"github.com/wostzone/hub/svc/certsvc/certconfig"
	"github.com/wostzone/hub/svc/certsvc/selfsigned"
	"github.com/wostzone/wost-go/pkg/certsclient"
)

//// CreateKeyPair generate a key pair in PEM format and save it to the cert folder
//// as <clientID>-pub.pem and <clientID>-key.pem
//// Returns the public key PEM content
//func CreateKeyPair(clientID string, certFolder string) (privKey *ecdsa.PrivateKey, err error) {
//	privKey = signing.CreateECDSAKeys()
//	privKeyFile := path.Join(certFolder, clientID+"-priv.pem")
//	pubKeyFile := path.Join(certFolder, clientID+"-pub.pem")
//	err = certsclient.SaveKeysToPEM(privKey, privKeyFile)
//	if err == nil {
//		pubKeyPem, _ := certsclient.PublicKeyToPEM(&privKey.PublicKey)
//		err = ioutil.WriteFile(pubKeyFile, []byte(pubKeyPem), 0644)
//	}
//	if err != nil {
//		fmt.Printf("Failed saving keys: %s\n", err)
//	}
//	if err == nil {
//		fmt.Printf("Generated public and private key pair as: %s and %s\n", pubKeyFile, privKeyFile)
//	}
//	return privKey, err
//}

// HandleCreateCertbundle generates the hub certificate bundle CA, Hub and Plugin keys
// and certificates.
//  If the CA certificate already exist it is NOT updated
//  If the Hub and Plugin certificates already exist, they are renewed
//func HandleCreateCertbundle(certsFolder string, sanName string) error {
//	err := selfsigned.CreateCertificateBundle([]string{sanName}, certsFolder, true)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Server and Plugin certificates generated in %s\n", certsFolder)
//	return nil
//}

func loadCA(certFolder string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	pemPath := path.Join(certFolder, certconfig.DefaultCaCertFile)
	caCert, err := certsclient.LoadX509CertFromPEM(pemPath)
	if err != nil {
		return nil, nil, err
	}
	pemPath = path.Join(certFolder, certconfig.DefaultCaKeyFile)
	caKey, err := certsclient.LoadKeysFromPEM(pemPath)
	if err != nil {
		return nil, nil, err
	}
	return caCert, caKey, nil
}

// Load public key or create a public/private key pair if not given
// If the path is a private key, then extract the public key from it
func loadOrCreateKey(keyFile string) (pubKey *ecdsa.PublicKey, generatedPrivKey *ecdsa.PrivateKey, err error) {
	// If a key file is given, use it, otherwise generate a pair
	if keyFile != "" {
		fmt.Printf("Using key file: %s\n", keyFile)
		pubKey, err = certsclient.LoadPublicKeyFromPEM(keyFile)
		// maybe this is a private key
		if err != nil {
			fmt.Println("not a public key, try loading as private key...")
			privKey, err2 := certsclient.LoadKeysFromPEM(keyFile)
			if err2 == nil {
				pubKey = &privKey.PublicKey
				err = nil
			}
		}
		// error out if this is neither a public nor a private key
		if err != nil {
			return nil, nil, err
		}
	} else {
		fmt.Printf("No public key file was provided. Creating a key pair ") // no newline
		generatedPrivKey = certsclient.CreateECDSAKeys()
		pubKey = &generatedPrivKey.PublicKey
	}
	return
}

// HandleCreateCACert generates the hub self-signed CA private key and certificate
// in the given folder.
// Use force to create the folder and overwrite existing certificate if it exists
func HandleCreateCACert(certsFolder string, sanName string, force bool) error {
	caCertPath := path.Join(certsFolder, certconfig.DefaultCaCertFile)
	caKeyPath := path.Join(certsFolder, certconfig.DefaultCaKeyFile)

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
		if _, err := os.Stat(caCertPath); err != nil {
			return fmt.Errorf("CA certificate already exists in '%s'", caCertPath)
		}
		if _, err := os.Stat(caKeyPath); err != nil {
			return fmt.Errorf("CA key alread exists in '%s'", caKeyPath)
		}
	}

	caCert, privKey, err := selfsigned.CreateHubCA()
	if err != nil {
		return err
	}
	err = certsclient.SaveX509CertToPEM(caCert, caCertPath)
	if err == nil {
		// this sets permissions to 0400 current user readonly
		err = certsclient.SaveKeysToPEM(privKey, caKeyPath)
	}

	fmt.Printf("Generated CA certificate '%s' and key '%s'\n", caCertPath, caKeyPath)
	return err
}

// HandleCreateClientCert creates a consumer client certificate and optionally private/public keypair
// This prints the certificate to stdout.
//
//  certFolder where to find the CA certificate and key used to sign the client certificate.
//  clientID for the CN of the client certificate. Used to identify the consumer.
//  keyFile with path to the client's public or private key
//  validity in days. 0 to use certconfig.DefaultClientCertDurationDays
func HandleCreateClientCert(certFolder string, clientID string, keyFile string, validityDays int) error {
	var pubKey *ecdsa.PublicKey
	var generatedPrivKey *ecdsa.PrivateKey
	var cert *x509.Certificate

	if validityDays == 0 {
		validityDays = certconfig.DefaultClientCertDurationDays
	}
	caCert, caKey, err := loadCA(certFolder)
	if err == nil {
		pubKey, generatedPrivKey, err = loadOrCreateKey(keyFile)
	}
	if err == nil {
		cert, err = selfsigned.CreateClientCert(clientID, certsclient.OUClient, pubKey, caCert, caKey, validityDays)
	}
	if err != nil {
		return err
	}
	certPem := certsclient.X509CertToPEM(cert)
	fmt.Printf("Certificate for %s, valid for %d days:\n", clientID, validityDays)
	fmt.Println(certPem)
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
		validityDays = certconfig.DefaultDeviceCertDurationDays
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
		validityDays = certconfig.DefaultServiceCertDurationDays
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
