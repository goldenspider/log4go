// Package log4go provides level-based and highly configurable logging.
//
// Enhanced Logging
//
// This is inspired by the logging functionality in Java.  Essentially, you create a Logger
// object and create output filters for it.  You can send whatever you want to the Logger,
// and it will filter that based on your settings and send it to the outputs.  This way, you
// can put as much debug code in your program as you want, and when you're done you can filter
// out the mundane messages so only the important ones show up.
//

package log4go

import (
	"encoding/json"
	"errors"
	"fmt"

	"path/filepath"
	"runtime"

	"time"
)

// Version information
const (
	L4G_VERSION = "log4go-v1.0.1"
	L4G_MAJOR   = 1
	L4G_MINOR   = 0
	L4G_BUILD   = 1
)

/****** Constants ******/

// These are the integer logging levels used by the logger
type Level int

const (
	DEBUG Level = iota
	TRACE
	INFO
	WARNING
	ERROR
	CRITICAL
)

// Default level passed to runtime.Caller
const DefaultFileDepth int = 3

// Logging level strings
var (
	levelStrings = [...]string{"DEBG", "TRAC", "INFO", "WARN", "EROR", "CRIT"}
)

func (l Level) String() string {
	if l < 0 || int(l) > len(levelStrings) {
		return "UNKNOWN"
	}
	return levelStrings[int(l)]
}

/****** Variables ******/
var (
	// LogBufferLength specifies how many log messages a particular log4go
	// logger can buffer at a time before writing them.
	LogBufferLength = 32
)

/****** LogRecord ******/

// A LogRecord contains all of the pertinent information for each message
type LogRecord struct {
	Level   Level     // The log level
	Created time.Time // The time at which the log message was created (nanoseconds)
	Source  string    // The message source
	Message string    // The log message
}

/****** LogWriter ******/

// This is an interface for anything that should be able to write logs
type LogWriter interface {
	// This will be called to log a LogRecord message.
	LogWrite(rec *LogRecord)

	// This should clean up anything lingering about the LogWriter, as it is called before
	// the LogWriter is removed.  LogWrite should not be called after Close.
	Close()
	Flush()
}

/****** Logger ******/

// A Filter represents the log level below which no log records are written to
// the associated LogWriter.
type Filter struct {
	Level Level

	rec     chan *LogRecord // write queue
	closing bool            // true if Socket was closed at API level

	LogWriter
}

func NewFilter(lvl Level, writer LogWriter) *Filter {
	f := &Filter{
		rec:     make(chan *LogRecord, LogBufferLength),
		closing: false,

		Level:     lvl,
		LogWriter: writer,
	}

	go f.run()
	return f
}

func (f *Filter) WriteToChan(rec *LogRecord) {
	if f.closing {
		//fmt.Fprintf(os.Stderr, "LogWriter: channel has been closed. Message is [%s]\n", rec.Message)
		return
	}
	f.rec <- rec
}

func (f *Filter) run() {
	for {
		select {
		case rec, ok := <-f.rec:
			if !ok {
				return
			}
			f.LogWrite(rec)
		}
	}
}

func (f *Filter) Close() {
	if f.closing {
		return
	}
	// sleep at most one second and let go routine running
	// drain the log channel before closing
	for i := 10; i > 0; i-- {
		time.Sleep(100 * time.Millisecond)
		if len(f.rec) <= 0 {
			break
		}
	}

	// block write channel
	f.closing = true

	defer f.LogWriter.Close()

	close(f.rec)

	if len(f.rec) <= 0 {
		return
	}
	// drain the log channel and write driect
	for rec := range f.rec {
		f.LogWrite(rec)
	}
}

func (f *Filter) Flush() {
	if f.closing {
		return
	}
	// sleep at most one second and let go routine running
	// drain the log channel before closing
	for i := 10; i > 0; i-- {
		time.Sleep(100 * time.Millisecond)
		if len(f.rec) <= 0 {
			break
		}
	}

	f.LogWriter.Flush()
}

// A Logger represents a collection of Filters through which log messages are
// written.
type Logger map[string]*Filter

// Create a new logger with a "stdout" filter configured to send log messages at
// or above lvl to standard output.
func NewDefaultLogger(lvl Level) Logger {
	return Logger{
		"stdout": NewFilter(lvl, NewConsoleLogWriter()),
	}
}

// Closes all log writers in preparation for exiting the program or a
// reconfiguration of logging.  Calling this is not really imperative, unless
// you want to guarantee that all log messages are written.  Close removes
// all filters (and thus all LogWriters) from the logger.
func (log Logger) Close() {
	// Close all open loggers
	for name, filt := range log {
		filt.Close()
		delete(log, name)
		fmt.Printf("Log close filter %s\n", name)
	}
}

func (log Logger) Flush() {
	// Flush all open loggers
	for name, filt := range log {
		filt.Flush()
		fmt.Printf("Log Flush filter %s\n", name)
	}
}

// Add a new LogWriter to the Logger which will only log messages at lvl or
// higher.  This function should not be called from multiple goroutines.
// Returns the logger for chaining.
func (log Logger) AddFilter(name string, lvl Level, writer LogWriter) Logger {
	log[name] = NewFilter(lvl, writer)
	return log
}

/******* Logging *******/

// Determine if any logging will be done
func (log Logger) skip(lvl Level) bool {
	for _, filt := range log {
		if lvl >= filt.Level {
			return false
		}
	}
	return true
}

// Dispatch the logs
func (log Logger) dispatch(rec *LogRecord) {
	for _, filt := range log {
		if rec.Level < filt.Level {
			continue
		}
		filt.WriteToChan(rec)
	}
}

// Send a formatted log message internally
func (log Logger) intLogf(lvl Level, format string, args ...interface{}) {
	if log.skip(lvl) {
		return
	}

	// Determine caller func
	pc, fullname, lineno, ok := runtime.Caller(DefaultFileDepth)
	src := ""
	if ok {
		src = fmt.Sprintf("%s %s:%d", fullname, filepath.Base(runtime.FuncForPC(pc).Name()), lineno)
	}

	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  src,
		Message: msg,
	}

	log.dispatch(rec)
}

// Send a log message with manual level, source, and message.
func (log Logger) Log(lvl Level, source, message string) {
	if log.skip(lvl) {
		return
	}

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  source,
		Message: message,
	}

	log.dispatch(rec)
}

// Send a log message with manual level, source, and message.
func (log Logger) Json(data []byte) {
	var rec LogRecord

	// Make the log record
	err := json.Unmarshal(data, &rec)
	if err != nil {
		// log to standard output
		msg := "Err: " + err.Error() + " - " + string(data[0:])
		log.intLogf(WARNING, msg)
		return
	}

	if log.skip(rec.Level) {
		return
	}

	log.dispatch(&rec)
}

//=================================================================
func (log Logger) debug(arg0 string, args ...interface{}) {
	log.intLogf(DEBUG, arg0, args...)

}

func (log Logger) trace(arg0 string, args ...interface{}) {
	log.intLogf(TRACE, arg0, args...)

}

func (log Logger) info(arg0 string, args ...interface{}) {
	log.intLogf(INFO, arg0, args...)
}

func (log Logger) warn(arg0 string, args ...interface{}) error {
	msg := fmt.Sprintf(arg0, args...)

	log.intLogf(WARNING, msg)
	return errors.New(msg)
}

func (log Logger) error(arg0 string, args ...interface{}) error {
	msg := fmt.Sprintf(arg0, args...)

	log.intLogf(ERROR, msg)
	return errors.New(msg)
}

func (log Logger) critical(arg0 string, args ...interface{}) error {
	msg := fmt.Sprintf(arg0, args...)

	log.intLogf(CRITICAL, msg)
	return errors.New(msg)
}
