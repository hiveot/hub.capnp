// Package config with logging configuration functions
package config

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// SetLogging sets the logging level and output file
// Intended for standardize logging in the hub and plugins
//  levelName is the requested logging level: error, warning, info, debug
//  filename is the output log file full name including path, use "" for stderr
//  timeFormat default is ISO8601 YYYY-MM-DDTHH:MM:SS.sss-TZ
func SetLogging(levelName string, filename string, timeFormat string) error {
	loggingLevel := logrus.DebugLevel
	var err error
	logrus.SetReportCaller(true)

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
			err = fmt.Errorf("SetLogging: Unable to open logfile: %s", err)
		} else {
			logrus.Warnf("SetLogging: Send '%s' logging to '%s'", levelName, filename)
			logOut = io.MultiWriter(logOut, logFileHandle)
		}
	}

	// Customize logging output with source file and line number
	logrus.SetFormatter(
		&logrus.TextFormatter{
			DisableColors:   false,
			ForceColors:     true,
			PadLevelText:    true,
			TimestampFormat: "2006-01-02T15:04:05.000-0700",
			FullTimestamp:   true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				funcName := f.Func.Name()
				// remove classname
				names := strings.Split(funcName, ".")
				if len(names) > 1 {
					funcName = names[len(names)-1]
				}
				levelColor := 37
				fileInfo := fmt.Sprintf(" \x1b[%dm%s:%v\x1b[0m", levelColor, path.Base(f.File), f.Line)
				funcName = fmt.Sprintf("\x1b[%dm%s\x1b[0m()", levelColor, funcName)
				// fileInfo := fmt.Sprintf(" %s:%v", path.Base(f.File), f.Line)
				return funcName, fileInfo
			},
		})
	logrus.SetOutput(logOut)
	logrus.SetLevel(loggingLevel)

	// var hook = ContextHook{}
	// logrus.AddHook(hook)

	// logrus.SetReportCaller(false) // publisher logging includes caller and file:line#

	return err
}
