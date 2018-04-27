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
func SetLogLevel(newLogLevel LogLevel) {
    logLevel = newLogLevel
}

/*
    Constants for logging
*/
const (
    fatalPrefix string = "FATAL: "
    errorPrefix string = "ERROR: "
    warnPrefix string = "WARN: "
    infoPrefix string = "INFO: "
    debugPrefix string = "DEBUG: "
)

/*
    Streams
*/
var (
    logLevel LogLevel
    fatalStream *log.Logger = log.New(os.Stderr, fatalPrefix, log.LstdFlags)
    errorStream *log.Logger = log.New(os.Stderr, errorPrefix, log.LstdFlags)
    warnStream *log.Logger = log.New(os.Stdout, warnPrefix, log.LstdFlags)
    infoStream *log.Logger = log.New(os.Stdout, infoPrefix, log.LstdFlags)
    debugStream *log.Logger = log.New(os.Stdout, debugPrefix, log.LstdFlags)
)

/*
    Log function definition
*/
type Logger func (format string, v ...interface{})

/*
    Utilities for logging
*/
func Fatal(format string, v ...interface{}) {
    fatalStream.Fatalf(format, v...)
}

func Error(format string, v ...interface{}) {
    if logLevel >= ERROR {
        return
    }
    errorStream.Printf(format, v...)
}

func Warn(format string, v ...interface{}) {
    if logLevel >= WARN {
        return
    }
    warnStream.Printf(format, v...)
}

func Info(format string, v ...interface{}) {
    if logLevel >= INFO {
        return
    }
    infoStream.Printf(format, v...)
}

func debug(format string, v ...interface{}) {
    if logLevel >= DEBUG {
        return
    }
    debugStream.Printf(format, v...)
}
