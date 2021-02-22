// Package lib with gateway logging helper functions
package lib

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// SetLogging sets the logging level and output file for this publisher
// Intended for standardize logging in the gateway and plugins
//  levelName is the requested logging level: error, warning, info, debug
//  filename is the output log file full name including path, use "" for stderr
func SetLogging(levelName string, filename string) error {
	loggingLevel := logrus.DebugLevel
	var err error

	if levelName != "" {
		switch strings.ToLower(levelName) {
		case "error":
			loggingLevel = logrus.ErrorLevel
		case "warn", "warning":
			loggingLevel = logrus.WarnLevel
		case "info":
			loggingLevel = logrus.InfoLevel
		case "debug":
			loggingLevel = logrus.DebugLevel
		}
	}
	var logOut io.Writer = os.Stdout
	if filename != "" {
		logFileHandle, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			err = fmt.Errorf("Publisher.SetLogging: Unable to open logfile: %s", err)
		} else {
			logrus.Warnf("Publisher.SetLogging: Send '%s' logging to '%s'", levelName, filename)
			logOut = io.MultiWriter(logOut, logFileHandle)
		}
	}

	// Work around bug in textformatter losing the color output
	// TODO: configure json logging output
	customFormatter := logrus.StandardLogger().Formatter.(*logrus.TextFormatter)
	// customFormatter := new(logrus.TextFormatter)
	customFormatter.FullTimestamp = true
	customFormatter.ForceColors = true
	// customFormatter.PadLevelText = true
	customFormatter.DisableColors = false
	customFormatter.DisableLevelTruncation = false
	logrus.SetFormatter(customFormatter)

	// logrus.SetFormatter(
	// 	&logrus.TextFormatter{
	// 		// LogFormat: "",
	// 		DisableColors: true,
	// 		// DisableLevelTruncation: true,
	// 		// PadLevelText:    true,
	// 		// TimestampFormat: "2006-01-02 15:04:05.000",
	// 		// FullTimestamp: true,
	// 		// ForceFormatting: true,
	// 	})
	logrus.SetOutput(logOut)
	logrus.SetLevel(loggingLevel)

	logrus.SetReportCaller(false) // publisher logging includes caller and file:line#
	return err
}
