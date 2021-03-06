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

// ContextHook to report source file...
type ContextHook struct{}

// Levels ...
func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire ...
func (hook ContextHook) Fire(entry *logrus.Entry) error {
	// stack depth depends on logging level: Warn=10, Info=9, Error=8   :/
	depth := 10
	if entry.Level == logrus.InfoLevel {
		depth = 9
	} else if entry.Level == logrus.ErrorLevel {
		depth = 8
	}
	if pc, file, line, ok := runtime.Caller(depth); ok {
		funcName := runtime.FuncForPC(pc).Name()
		fileInfo := fmt.Sprintf("%s:%v", path.Base(file), line)
		entry.Data["source"] = fmt.Sprintf("%s:%s", fileInfo, path.Base(funcName))
	}
	return nil
}

// SetLogging sets the logging level and output file for this publisher
// Intended for standardize logging in the gateway and plugins
//  levelName is the requested logging level: error, warning, info, debug
//  filename is the output log file full name including path, use "" for stderr
//  timeFormat default is ISO8601 YYYY-MM-DDTHH:MM:SS.sss-TZ
func SetLogging(levelName string, filename string, timeFormat string) error {
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

	logrus.SetFormatter(
		&logrus.TextFormatter{
			// LogFormat: "",
			// DisableColors: true,
			ForceColors: true,
			// DisableLevelTruncation: true,
			PadLevelText:    true,
			TimestampFormat: "2006-01-02T15:04:05.000-0700",
			FullTimestamp:   true,
			// ForceFormatting: true,
		})
	logrus.SetOutput(logOut)
	logrus.SetLevel(loggingLevel)

	var hook = ContextHook{}
	logrus.AddHook(hook)

	logrus.SetReportCaller(false) // publisher logging includes caller and file:line#

	return err
}
