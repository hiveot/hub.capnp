package logging

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitZap initializes the zap logger
func InitZap(levelName string, filename string) *zap.Logger {
	var core zapcore.Core

	zapLogger, _ := zap.NewProduction()
	loggingLevel := zap.DebugLevel

	if levelName != "" {
		switch strings.ToLower(levelName) {
		case "error":
			loggingLevel = zap.ErrorLevel
		case "warn", "warning":
			loggingLevel = zap.WarnLevel
		case "info":
			loggingLevel = zap.InfoLevel
		case "debug":
			loggingLevel = zap.DebugLevel
		}
	}
	// setup logging to stdout

	config := zap.NewDevelopmentEncoderConfig()
	//config.EncodeDuration = zapcore.MillisDurationEncoder
	//config.MessageKey = "message"
	//config.LevelKey = "level"
	//config.EncodeDuration = zapcore.MillisDurationEncoder
	//config.EncodeCaller = zapcore.FullCallerEncoder
	//config.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewConsoleEncoder(config)
	core = zapcore.NewCore(encoder, os.Stdout, loggingLevel)

	// fileEncoder := zapcore.NewJSONEncoder(config)

	// writer := zapcore.NewMultiWriteSyncer(
	// 	zapcore.AddSync(os.Stdout),
	// 	GetWriteSyncer(filename),
	// )
	// optionally add logging to file
	if filename != "" {
		// use a core which logs to file
		logFile, err2 := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
		// optionally allow different config for file logging
		fileConfig := config
		//fileConfig.EncodeDuration = zapcore.MillisDurationEncoder
		//fileConfig.MessageKey = "message"
		//fileConfig.LevelKey = "level"
		//fileConfig.EncodeDuration = zapcore.MillisDurationEncoder
		//fileConfig.EncodeCaller = zapcore.FullCallerEncoder
		//fileConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		fileEncoder := zapcore.NewJSONEncoder(fileConfig)
		if err2 != nil {
			fmt.Printf("SetLogging: Unable to open logfile: " + err2.Error())
		} else {
			writer := zapcore.AddSync(logFile)
			// combine both cores into a new core
			core = zapcore.NewTee(
				core,
				zapcore.NewCore(fileEncoder, writer, loggingLevel),
			)
			zapLogger.Sugar().Info("SetLogging: Send '%s' logging to '%s'", levelName, filename)
		}
	}
	// show caller and stack trace -> can't this be done separately for console vs file?
	zapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return zapLogger
}
