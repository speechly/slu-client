package logger

import (
	"fmt"
	"io"
	"os"
)

// stdLogger is a simple logger that uses fmt.Fprintln / fmt.Fprintf to log messages.
// Destination can be specified by the caller.
type stdLogger struct {
	dst io.Writer
}

// NewStderrLogger returns a new StdLogger that prints to STDERR.
func NewStderrLogger() Logger {
	return NewStdLogger(os.Stderr)
}

// NewStdLogger returns a new StdLogger that prints to dst.
func NewStdLogger(dst io.Writer) Logger {
	return &stdLogger{
		dst: dst,
	}
}

func (l *stdLogger) Debug(v ...interface{}) {
	fmt.Fprintln(l.dst, v...)
}

func (l *stdLogger) Debugf(f string, v ...interface{}) {
	fmt.Fprintf(l.dst, f+"\n", v...)
}

func (l *stdLogger) Info(v ...interface{}) {
	fmt.Fprintln(l.dst, v...)
}

func (l *stdLogger) Infof(f string, v ...interface{}) {
	fmt.Fprintf(l.dst, f+"\n", v...)
}

func (l *stdLogger) Warn(v ...interface{}) {
	fmt.Fprintln(l.dst, v...)
}

func (l *stdLogger) Warnf(f string, v ...interface{}) {
	fmt.Fprintf(l.dst, f+"\n", v...)
}

func (l *stdLogger) Error(v ...interface{}) {
	fmt.Fprintln(l.dst, v...)
}

func (l *stdLogger) Errorf(f string, v ...interface{}) {
	fmt.Fprintf(l.dst, f+"\n", v...)
}
