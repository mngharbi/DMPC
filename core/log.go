package core

/*
	Utilities for logging
*/

import (
	"log"
	"os"
)

/*
   Log levels
*/
type LogLevel int

const (
	FATAL LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
)

/*
	Generates a logging handler
	with new stdout/stderr streams
*/
func InitializeLogging() *LoggingHandler {
	return &LoggingHandler{
		logLevel:     FATAL,
		stderrStream: log.New(os.Stderr, "", log.LstdFlags),
		stdoutStream: log.New(os.Stdout, "", log.LstdFlags),
	}
}

/*
	Changes log level for a logging handler
*/
func (logHandler *LoggingHandler) SetLogLevel(logLevel LogLevel) {
	logHandler.logLevel = logLevel
}

/*
   Constants for logging
*/
const (
	fatalPrefix string = "FATAL: "
	errorPrefix string = "ERROR: "
	warnPrefix  string = "WARN: "
	infoPrefix  string = "INFO: "
	debugPrefix string = "DEBUG: "
)

/*
   Structure that keeps streams
   used to use the same streams across packages
*/
type LoggingHandler struct {
	logLevel     LogLevel
	stderrStream *log.Logger
	stdoutStream *log.Logger
}

/*
   Utilities for logging
*/
func (logHandler *LoggingHandler) Fatalf(format string, v ...interface{}) {
	logHandler.stderrStream.Fatalf(fatalPrefix+format, v...)
}

func (logHandler *LoggingHandler) Errorf(format string, v ...interface{}) {
	if logHandler.logLevel < ERROR {
		return
	}
	logHandler.stderrStream.Printf(errorPrefix+format, v...)
}

func (logHandler *LoggingHandler) Warnf(format string, v ...interface{}) {
	if logHandler.logLevel < WARN {
		return
	}
	logHandler.stdoutStream.Printf(warnPrefix+format, v...)
}

func (logHandler *LoggingHandler) Infof(format string, v ...interface{}) {
	if logHandler.logLevel < INFO {
		return
	}
	logHandler.stdoutStream.Printf(infoPrefix+format, v...)
}

func (logHandler *LoggingHandler) Debugf(format string, v ...interface{}) {
	if logHandler.logLevel < DEBUG {
		return
	}
	logHandler.stdoutStream.Printf(debugPrefix+format, v...)
}
