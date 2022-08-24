package certcli_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wostzone/hub/launcher/cmd/certcli"
	"github.com/wostzone/hub/svc/certsvc/certconfig"
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
}

func TestNoArgs(t *testing.T) {

}
