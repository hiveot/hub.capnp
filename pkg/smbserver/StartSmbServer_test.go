package smbserver_test

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/pkg/config"
	"github.com/wostzone/hub/pkg/smbserver"
)

var homeFolder string

// Use the project app folder during testing
func init() {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
}
func TestStartSmbServer(t *testing.T) {
	// test on a different port as to not interfere with running application or test server
	os.Args = append(os.Args[0:1], strings.Split("-hostname localhost:9998", " ")...)

	hubConfig, err := config.SetupConfig(homeFolder, "", nil)
	assert.NoError(t, err)
	smb, err := smbserver.StartSmbServer(hubConfig)
	assert.NoError(t, err)
	time.Sleep(time.Second)
	smb.Stop()
}
