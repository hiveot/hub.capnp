package internal_test

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/plugins/smbus/internal"
)

var appFolder string

// Use the project app folder during testing
func init() {
	cwd, _ := os.Getwd()
	appFolder = path.Join(cwd, "../../../test")
}
func TestStartSmbus(t *testing.T) {
	// test on a different port as to not interfere with running application or test server
	os.Args = append(os.Args[0:1], strings.Split("-hostname localhost:9998", " ")...)

	smb, err := internal.StartSmbus(appFolder)
	assert.NoError(t, err)
	time.Sleep(time.Second)
	smb.Stop()
}
