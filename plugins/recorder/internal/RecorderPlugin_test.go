package internal_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/plugins/recorder/internal"
)

var appFolder string

// Use the project app folder during testing
func init() {
	cwd, _ := os.Getwd()
	appFolder = path.Join(cwd, "../../../test")
}

func TestStartRecorderPlugin(t *testing.T) {
	config := &internal.RecorderConfig{}
	rec := internal.NewRecorder(config)
	rec.Start()
}

func TestStartRecorder(t *testing.T) {
	rec, err := internal.StartRecorder(appFolder)
	assert.NoError(t, err)
	rec.Stop()
}
