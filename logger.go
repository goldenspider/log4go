package log4go

import (
	"fmt"
)

var log = make(Logger)

func StartLogServer(cfgfile ...string) {
	if len(cfgfile) == 0 {
		log.LoadConfig("config.toml")
	} else {
		log.LoadConfig(cfgfile[0])
	}
}

func StopLogServer() {
	log.Close()
}

func LogDebugf(format string, params ...interface{}) {
	log.debug(format, params...)
}

func LogTracef(format string, params ...interface{}) {
	log.trace(format, params...)
}

func LogInfof(format string, params ...interface{}) {
	log.info(format, params...)
}

func LogWarnf(format string, params ...interface{}) error {
	return log.warn(format, params...)
}

func LogErrorf(format string, params ...interface{}) error {
	return log.error(format, params...)
}

func LogCriticalf(format string, params ...interface{}) error {
	return log.critical(format, params...)
}

///////////////////////////////////////////////////
func LogDebug(v ...interface{}) {
	log.debug("%s", fmt.Sprint(v...))
}

func LogTrace(v ...interface{}) {
	log.trace("%s", fmt.Sprint(v...))
}

func LogInfo(v ...interface{}) {
	log.info("%s", fmt.Sprint(v...))
}

func LogWarn(v ...interface{}) error {
	return log.warn("%s", fmt.Sprint(v...))
}

func LogError(v ...interface{}) error {
	return log.error("%s", fmt.Sprint(v...))
}

func LogCritical(v ...interface{}) error {
	return log.critical("%s", fmt.Sprint(v...))
}

func LogFlush() {
	log.Flush()
}
