// Package logging with logging configuration
package logging

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

var Logger *logrus.Logger
var ZapLogger *zap.Logger

// SetLogging initializes the global logger
func SetLogging(levelName string, logFile string) {
	Logger = InitLogrus(levelName, logFile)
	ZapLogger = InitZap(levelName, "")
}
