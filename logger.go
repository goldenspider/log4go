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

//func collecParams(dst []interface{}, params ...interface{})
func LogDebugf(format string, params ...interface{}) {
	log.Debug(format, params...)
}

func LogTracef(format string, params ...interface{}) {
	log.Trace(format, params...)
}

func LogInfof(format string, params ...interface{}) {
	log.Info(format, params...)
}

func LogWarnf(format string, params ...interface{}) {
	log.Warn(format, params...)
}

func LogErrorf(format string, params ...interface{}) {
	log.Error(format, params...)
}

func LogCriticalf(format string, params ...interface{}) {
	log.Critical(format, params...)
}

///////////////////////////////////////////////////
func LogDebug(v ...interface{}) {
	log.Debug("%s", fmt.Sprint(v...))
}

func LogTrace(v ...interface{}) {
	log.Trace("%s", fmt.Sprint(v...))
}

func LogInfo(v ...interface{}) {
	log.Info("%s", fmt.Sprint(v...))
}

func LogWarn(v ...interface{}) {
	log.Warn("%s", fmt.Sprint(v...))
}

func LogError(v ...interface{}) {
	log.Error("%s", fmt.Sprint(v...))
}

func LogCritical(v ...interface{}) {
	log.Critical("%s", fmt.Sprint(v...))
}

func LogFlush() {
	log.Flush()
}