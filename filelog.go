package log4go

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	BUFFERSIZE = 4 * 1024 * 1024
)

type FileLogWriter struct {
	filename string
	path     string
	bufsize  int
	iow      *bytes.Buffer
	format   string
	compress bool
	wg       sync.WaitGroup
}

// This creates a new FileLogWriter
func NewFileLogWriter(fname string) *FileLogWriter {
	c := &FileLogWriter{
		filename: fname,
		path:     "",
		bufsize:  BUFFERSIZE,
		iow:      nil,
		format:   "[%T %D %Z] [%L] (%S) %M",
		compress: false,
	}
	return c
}

func (c *FileLogWriter) SetFormat(format string) *FileLogWriter {
	c.format = format
	return c
}

func (c *FileLogWriter) SetBufSize(bufsize int) {
	if bufsize == 0 {
		c.bufsize = BUFFERSIZE
	} else {
		c.bufsize = bufsize
	}
	return
}

func (c *FileLogWriter) SetCompress(compress bool) {
	c.compress = compress
	return
}

func (c *FileLogWriter) SetPath(path string) {
	c.path = filepath.Clean(path) + "/"
	return
}

func (c *FileLogWriter) Close() {
	c.wg.Wait()

	if c.iow == nil || c.iow.Len() == 0 {
		return
	}

	sfilename := c.MakeFileName()
	fd, err := os.OpenFile(sfilename, os.O_WRONLY|os.O_CREATE, 0660)

	defer fd.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%s): %s\n", sfilename, err)
		return
	}

	tmp := c.iow
	c.iow = bytes.NewBuffer(make([]byte, 0, c.bufsize))

	tmp.WriteTo(fd)
	fd.Sync()
	time.Sleep(200 * time.Millisecond)
}

func (c *FileLogWriter) Flush() {
	c.Close()
}

// Set the logging format (chainable).  Must be called before the first log
// message is written.
//example-20160314160255-814856400.log
func (c *FileLogWriter) MakeFileName() string {
	out := bytes.NewBuffer(make([]byte, 0, 64))
	t := time.Now()
	//fmt.Println(time.Now().String())
	out.WriteString(fmt.Sprintf("%04d%02d%02d", t.Year(), t.Month(), t.Day()))
	out.WriteString(fmt.Sprintf("%02d%02d%02d", t.Hour(), t.Minute(), t.Second()))
	out.WriteString(fmt.Sprintf("-%d", t.Nanosecond()))
	sfilename := fmt.Sprintf("%s%s-%s.log", c.path, c.filename, out.String())
	return sfilename
}

func (c *FileLogWriter) LogWrite(rec *LogRecord) {
	s := FormatLogRecord(c.format, rec)
	if c.iow == nil {
		c.iow = bytes.NewBuffer(make([]byte, 0, c.bufsize))
	}
	c.iow.WriteString(s)

	if c.iow.Len() > c.bufsize {
		tmp := c.iow
		c.iow = bytes.NewBuffer(make([]byte, 0, c.bufsize))
		c.wg.Add(1)
		go func() {
			sfilename := c.MakeFileName()

			fd, err := os.OpenFile(sfilename, os.O_WRONLY|os.O_CREATE, 0660)
			defer fd.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "FileLogWriter(%s): %s\n", sfilename, err)
				return
			}

			tmp.WriteTo(fd)
			fd.Sync()
			c.wg.Done()
		}()
	}
}
