package log4go

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/daviddengcn/go-colortext"
)

var stdout io.Writer = os.Stdout

type RecInfo struct {
	isQuit bool
	level  Level

	data string
}

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct {
	iow    io.Writer
	color  bool
	format string
	wg     sync.WaitGroup
	rec    chan *RecInfo // write queue
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {
	c := &ConsoleLogWriter{
		iow:    stdout,
		color:  false,
		format: "[%T %D] [%L] (%S) %M",
		rec:    make(chan *RecInfo, 256),
	}
	go func() {
		c.wg.Add(1)
	LOOP:
		for {
			select {
			case rec := <-c.rec:
				if rec.isQuit == true {
					c.wg.Done()
					break LOOP
				}
				if c.color {
					switch rec.level {
					case CRITICAL:
						ct.ChangeColor(ct.Red, true, ct.White, false)
					case ERROR:
						ct.ChangeColor(ct.Red, false, 0, false)
					case WARNING:
						ct.ChangeColor(ct.Yellow, false, 0, false)
					case INFO:
						ct.ChangeColor(ct.Green, false, 0, false)
					case DEBUG:
						ct.ChangeColor(ct.Magenta, false, 0, false)
					case TRACE:
						ct.ChangeColor(ct.Cyan, false, 0, false)
					default:
					}
					fmt.Fprint(c.iow, rec.data)
					ct.ResetColor()
				} else {
					fmt.Fprint(c.iow, rec.data)
				}
			}
		}
	}()
	return c
}

// Must be called before the first log message is written.
func (c *ConsoleLogWriter) SetColor(color bool) *ConsoleLogWriter {
	c.color = color
	return c
}

// Set the logging format (chainable).  Must be called before the first log
// message is written.
func (c *ConsoleLogWriter) SetFormat(format string) *ConsoleLogWriter {
	c.format = format
	return c
}

func (c *ConsoleLogWriter) Close() {
	c.rec <- &RecInfo{isQuit: true}
	c.wg.Wait()
}

func (c *ConsoleLogWriter) Flush() {
}

func (c *ConsoleLogWriter) LogWrite(rec *LogRecord) {
	c.rec <- &RecInfo{data: FormatLogRecord(c.format, rec), level: rec.Level}
}

