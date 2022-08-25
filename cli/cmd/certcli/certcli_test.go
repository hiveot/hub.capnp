package certcli_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wostzone/hub/launcher/cmd/certcli"
	"github.com/wostzone/hub/svc/certsvc/certconfig"
	"github.com/wostzone/wost-go/pkg/certsclient"
)

func TestCreateCA(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	force := true
	sanName := "test"
	//_ = os.MkdirAll(certsFolder, 0700)
	//_ = os.Chdir(tempFolder)

	err := certcli.HandleCreateCACert(tempFolder, sanName, force)
	assert.NoError(t, err)

	certPath := path.Join(tempFolder, certconfig.DefaultCaCertFile)
	assert.FileExists(t, certPath)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateCA_ErrorExists(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	// create the cert
	force := true
	err := certcli.HandleCreateCACert(tempFolder, "test", force)
	assert.NoError(t, err)

	// error cert exists
	force = false
	err = certcli.HandleCreateCACert(tempFolder, "test", force)
	assert.Error(t, err)

	// error key exists
	os.Remove(path.Join(tempFolder, certconfig.DefaultCaCertFile))
	force = false
	err = certcli.HandleCreateCACert(tempFolder, "test", force)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateCA_FolderDoesntExists(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	_ = os.RemoveAll(tempFolder)

	force := false
	err := certcli.HandleCreateCACert(tempFolder, "test", force)
	assert.Error(t, err)
}

func TestCreateClientCert(t *testing.T) {
	clientID := "client"
	keyFile := ""
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	err := certcli.HandleCreateCACert(tempFolder, "test", true)
	assert.NoError(t, err)

	// create the cert
	err = certcli.HandleCreateClientCert(tempFolder, clientID, keyFile, 0)
	assert.NoError(t, err)

	// missing key file
	keyFile = "missingkeyfile.pem"
	err = certcli.HandleCreateClientCert(tempFolder, clientID, keyFile, 0)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateDeviceCert(t *testing.T) {
	deviceID := "urn:publisher:device1"
	keyFile := ""
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	err := certcli.HandleCreateCACert(tempFolder, "test", true)
	assert.NoError(t, err)

	err = certcli.HandleCreateDeviceCert(tempFolder, deviceID, keyFile, 0)
	assert.NoError(t, err)

	// missing key file
	keyFile = "missingkeyfile.pem"
	err = certcli.HandleCreateDeviceCert(tempFolder, deviceID, keyFile, 0)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateServiceCert(t *testing.T) {
	serviceID := "service25"
	keyFile := ""
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	err := certcli.HandleCreateCACert(tempFolder, "test", true)
	assert.NoError(t, err)

	err = certcli.HandleCreateServiceCert(tempFolder, serviceID, "127.0.0.1", keyFile, 0)
	assert.NoError(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateServiceCertWithKey(t *testing.T) {
	serviceID := "service25"
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	keyFile := path.Join(tempFolder, serviceID+".pem")
	err := certcli.HandleCreateCACert(tempFolder, "test", true)
	assert.NoError(t, err)

	privKey := certsclient.CreateECDSAKeys()
	err = certsclient.SaveKeysToPEM(privKey, keyFile)
	assert.NoError(t, err)
	// use a valid key
	err = certcli.HandleCreateServiceCert(tempFolder, serviceID, "", keyFile, 0)
	assert.NoError(t, err)

	// missing key file
	keyFile2 := path.Join(tempFolder, "keydoesntexist.pem")
	err = certcli.HandleCreateServiceCert(tempFolder, serviceID, "", keyFile2, 0)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateServiceCertMissingCA(t *testing.T) {
	serviceID := "service25"
	keyFile := ""
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	_ = os.RemoveAll(tempFolder)

	err := certcli.HandleCreateServiceCert(tempFolder, serviceID, "", keyFile, 1)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}
