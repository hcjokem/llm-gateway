package util

import (
	"log"
	"os"
)

type LogLevel string

const (
	DEBUG LogLevel = "debug"
	INFO  LogLevel = "info"
	WARN  LogLevel = "warn"
	ERROR LogLevel = "error"
)

type Logger struct {
	level LogLevel
	logger *log.Logger
}

func NewLogger(level string) *Logger {
	return &Logger{
		level:  LogLevel(level),
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Debug(msg string) {
	if l.level == DEBUG {
		l.logger.Printf("[DEBUG] %s", msg)
	}
}

func (l *Logger) Info(msg string) {
	l.logger.Printf("[INFO] %s", msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Printf("[WARN] %s", msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Printf("[ERROR] %s", msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.level == DEBUG {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Printf("[WARN] "+format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}
