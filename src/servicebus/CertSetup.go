// Package servicebus with creation of self signed certificate chain
// Source: https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251
package servicebus

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"
)

const keySize = 2048 // 4096
const caDurationYears = 10
const certDurationYears = 10

// func GenCARoot() (*x509.Certificate, []byte, *rsa.PrivateKey) {
// 	if _, err := os.Stat("someFile"); err == nil {
// 		//read PEM and cert from file
// 	}
// 	var rootTemplate = x509.Certificate{
// 		SerialNumber: big.NewInt(1),
// 		Subject: pkix.Name{
// 			Country:      []string{"SE"},
// 			Organization: []string{"Company Co."},
// 			CommonName:   "Root CA",
// 		},
// 		NotBefore:             time.Now().Add(-10 * time.Second),
// 		NotAfter:              time.Now().AddDate(10, 0, 0),
// 		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
// 		BasicConstraintsValid: true,
// 		IsCA:                  true,
// 		MaxPathLen:            2,
// 		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
// 	}
// 	priv, err := rsa.GenerateKey(rand.Reader, 2048)
// 	if err != nil {
// 		panic(err)
// 	}
// 	rootCert, rootPEM := genCert(&rootTemplate, &rootTemplate, &priv.PublicKey, priv)
// 	return rootCert, rootPEM, priv
// }

// CreateWoSTCA creates WoST CA and certificate and private key for signing server certificates
func CreateWoSTCA() (certPEM []byte, keyPEM []byte) {
	// set up our CA certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:  []string{"WoST Zone"},
			Country:       []string{"CA"},
			Province:      []string{"BC"},
			Locality:      []string{"Project"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "Root CA",
		},
		NotBefore: time.Now().Add(-10 * time.Second),
		NotAfter:  time.Now().AddDate(caDurationYears, 0, 0),
		// KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		// ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		// MaxPathLen: 2,
		// 		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create the CA private key
	privKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		logrus.Errorf("CertSetup: Unable to create private key: %s", err)
		return nil, nil
	}

	// PEM encode private key
	privKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(privKeyPEMBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privKey.PublicKey, privKey)
	if err != nil {
		logrus.Errorf("CertSetup: Unable to create CA cert: %s", err)
		return nil, nil
	}

	// pem encode certificate
	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	return certPEMBuffer.Bytes(), privKeyPEMBuffer.Bytes()
}

// CreateGatewayCert creates Wost gateway server key and certificate
func CreateGatewayCert(caCertPEM []byte, caKeyPEM []byte, hostname string) (pkPEM []byte, certPEM []byte, err error) {
	// We need the CA key and certificate
	caPrivKeyBlock, _ := pem.Decode(caKeyPEM)
	caPrivKey, err := x509.ParsePKCS1PrivateKey(caPrivKeyBlock.Bytes)
	certBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"WoST Zone"},
			Country:       []string{"CA"},
			Province:      []string{"BC"},
			Locality:      []string{"WoST Gateway"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    hostname,
		},
		// IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(certDurationYears, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &privKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	privKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(privKeyPEMBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	return certPEMBuffer.Bytes(), privKeyPEMBuffer.Bytes(), nil

	// serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	// if err != nil {
	// 	return nil, nil, err
	// }

	// serverTLSConf = &tls.Config{
	// 	Certificates: []tls.Certificate{serverCert},
	// }

	// certpool := x509.NewCertPool()
	// certpool.AppendCertsFromPEM(caPEM.Bytes())
	// clientTLSConf = &tls.Config{
	// 	RootCAs: certpool,
	// }

	// return
}

// func createCert() {
// 	// get our ca and server certificate
// 	serverTLSConf, clientTLSConf, err := certsetup()
// 	if err != nil {
// 		panic(err)
// 	}

// 	// set up the httptest.Server using our certificate signed by our CA
// 	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintln(w, "success!")
// 	}))
// 	server.TLS = serverTLSConf
// 	server.StartTLS()
// 	defer server.Close()

// 	// communicate with the server using an http.Client configured to trust our CA
// 	transport := &http.Transport{
// 		TLSClientConfig: clientTLSConf,
// 	}
// 	http := http.Client{
// 		Transport: transport,
// 	}
// 	resp, err := http.Get(server.URL)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// verify the response
// 	respBodyBytes, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		panic(err)
// 	}
// 	body := strings.TrimSpace(string(respBodyBytes[:]))
// 	if body == "success!" {
// 		fmt.Println(body)
// 	} else {
// 		panic("not successful!")
// 	}
// // }
