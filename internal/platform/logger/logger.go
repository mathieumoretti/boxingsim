package logger

import (
	"fmt"
	"os"
)

type Logger struct {
	info  *os.File
	error *os.File
}

func New(name string) *Logger {
	return &Logger{
		info:  os.Stdout,
		error: os.Stderr,
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	msg := format
	if len(v) > 0 {
		msg += " " + fmt.Sprintf("%v", v...)
	}
	_, _ = l.info.WriteString("[INFO] " + msg + "\n")
}

func (l *Logger) Error(format string, v ...interface{}) {
	msg := format
	if len(v) > 0 {
		msg += " " + fmt.Sprintf("%v", v...)
	}
	_, _ = l.error.WriteString("[ERROR] " + msg + "\n")
}

func (l *Logger) Debug(format string, v ...interface{}) {
	msg := format
	if len(v) > 0 {
		msg += " " + fmt.Sprintf("%v", v...)
	}
	_, _ = l.info.WriteString("[DEBUG] " + msg + "\n")
}
