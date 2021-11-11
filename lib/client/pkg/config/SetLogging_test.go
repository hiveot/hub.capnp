package config_test

import (
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/config"
)

func TestLogging(t *testing.T) {
	wd, _ := os.Getwd()
	logFile := path.Join(wd, "../../test/logs/TestLogging.log")

	os.Remove(logFile)
	config.SetLogging("info", logFile)
	logrus.Info("Hello info")
	config.SetLogging("debug", logFile)
	logrus.Debug("Hello debug")
	config.SetLogging("warn", logFile)
	logrus.Warn("Hello warn")
	config.SetLogging("error", logFile)
	logrus.Error("Hello error")
	assert.FileExists(t, logFile)
	os.Remove(logFile)
}

func TestLoggingBadFile(t *testing.T) {
	logFile := "/root/cantloghere.log"

	err := config.SetLogging("info", logFile)
	assert.Error(t, err)
	os.Remove(logFile)
}
