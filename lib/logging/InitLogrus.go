package logging

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

// InitLogrus initializes and returns the logrus logger
//
//  1. sets the logging level
//
//  2. sets the timeFormat to ISO8601 YYYY-MM-DDTHH:MM:SS.sss-TZ
//
//  3. optionally write all logging to file
//
//  4. write debug and info to stdout
//
//  5. write warning, error, fatal to stderr
//
//     levelName is the requested logging level: "error", "warning", "info", "debug"
//     filename is the output log file full name including path
func InitLogrus(levelName string, logFile string) *logrus.Logger {
	loggingLevel := logrus.DebugLevel
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
	//var logOut io.Writer = os.Stdout

	logrus.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	if logFile != "" {
		logFileHandle, err2 := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
		if err2 != nil {
			logrus.Errorf("SetLogging: Unable to open logfile: %s", err2)
		} else {
			logrus.Infof("SetLogging: Send '%s' logging to '%s'", levelName, logFile)
			//logOut = io.MultiWriter(os.Stdout, logFileHandle)
			logrus.SetOutput(logFileHandle) // Send all logs to file
		}
	}

	// Customize logging output with source file and line number
	logrus.SetFormatter(
		&logrus.TextFormatter{
			DisableColors:   false,
			ForceColors:     true,
			PadLevelText:    true,
			TimestampFormat: "2006-01-02T15:04:05.0000-0700",
			FullTimestamp:   true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				funcName := f.Func.Name()

				// remove classname
				names := strings.Split(funcName, ".")
				if len(names) > 1 {
					funcName = names[len(names)-1]
				}
				// levelColor := 37
				// fileInfo := fmt.Sprintf(" \x1b[%dm%s:%v\x1b[0m", levelColor, path.Base(f.File), f.Line)
				// funcName = fmt.Sprintf("\x1b[%dm%s\x1b[0m()", levelColor, funcName)

				// remove the path from the function name
				_, funcName = path.Split(funcName)
				funcName += "(): "
				//funcName = fmt.Sprintf("%-30s", funcName)

				fileName := path.Base(f.File)
				//if len(fileName) > 15 {
				//	fileName = fileName[:10] + "..."
				//}
				fileInfo := fmt.Sprintf(" %s:%v", fileName, f.Line)
				fileInfo = fmt.Sprintf("%s", fileInfo)
				return funcName, fileInfo
			},
		})
	logrus.SetLevel(loggingLevel)
	//logrus.SetOutput(logOut)

	// send errors and fatal to stderr instead of stdout
	//logrus.SetOutput(ioutil.Discard) // Send all logs to nowhere by default
	logrus.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	})
	logrus.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	})
	return logrus.StandardLogger()
}
