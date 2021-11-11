package testenv

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// SetLogging sets the logging level and output file
// This sets the timeFormat to ISO8601 YYYY-MM-DDTHH:MM:SS.sss-TZ
// Intended for standardize logging in the hub and plugins
//  levelName is the requested logging level: error, warning, info, debug
//  filename is the output log file full name including path, use "" for stderr
func SetLogging(levelName string, filename string) {
	loggingLevel := logrus.DebugLevel
	// logrus.SetReportCaller(true)
	var logOut io.Writer = os.Stdout

	logrus.SetOutput(logOut)
	logrus.SetLevel(loggingLevel)
}
