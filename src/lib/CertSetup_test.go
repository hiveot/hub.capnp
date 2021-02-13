package lib

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTLSCertificateGeneration(t *testing.T) {
	hostname := "127.0.0.1"

	// test creating ca and server certificates
	caCertPEM, caKeyPEM := CreateWoSTCA()
	require.NotNilf(t, caCertPEM, "Failed creating CA certificate")
	caCert, err := tls.X509KeyPair(caCertPEM, caKeyPEM)
	_ = caCert
	require.NoErrorf(t, err, "Failed parsing CA certificate")

	clientCertPEM, clientKeyPEM, err := CreateClientCert(caCertPEM, caKeyPEM, hostname)
	require.NoErrorf(t, err, "Creating certificates failed:")
	require.NotNilf(t, clientCertPEM, "Failed creating client certificate")
	require.NotNilf(t, clientKeyPEM, "Failed creating client key")

	serverCertPEM, serverKeyPEM, err := CreateGatewayCert(caCertPEM, caKeyPEM, hostname)
	require.NoErrorf(t, err, "Failed creating server certificate")
	// serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	require.NoErrorf(t, err, "Failed creating server certificate")
	require.NotNilf(t, serverCertPEM, "Failed creating server certificate")
	require.NotNilf(t, serverKeyPEM, "Failed creating server private key")

	// verify the certificate
	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM(caCertPEM)
	require.True(t, ok, "Failed parsing CA certificate")

	serverBlock, _ := pem.Decode(serverCertPEM)
	require.NotNil(t, serverBlock, "Failed decoding server certificate PEM")

	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	require.NoError(t, err, "ParseCertificate for server failed")

	opts := x509.VerifyOptions{
		Roots:   certpool,
		DNSName: hostname,
		// DNSName:       "127.0.0.1",
		Intermediates: x509.NewCertPool(),
	}
	_, err = serverCert.Verify(opts)
	require.NoError(t, err, "Verify for server certificate failed")
}

func TestBadCert(t *testing.T) {
	hostname := "127.0.0.1"
	caCertPEM, caKeyPEM := CreateWoSTCA()
	// caCertPEM = pem.Encode( )[]byte{1, 2, 3}

	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "",
		Bytes: []byte{1, 2, 3},
	})
	caCertPEM = certPEMBuffer.Bytes()

	clientCertPEM, clientKeyPEM, err := CreateClientCert(caCertPEM, caKeyPEM, hostname)
	assert.Errorf(t, err, "Creating certificates should fail")
	assert.Nilf(t, clientCertPEM, "Created client certificate")
	assert.Nilf(t, clientKeyPEM, "Created client key")

}
