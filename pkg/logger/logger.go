package logger

import (
	"fmt"
	"log"
	"os"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level  Level
	logger *log.Logger
}

type Options struct {
	Level  Level
	Prefix string
}

func New(opts Options) *Logger {
	flags := log.Ldate | log.Ltime | log.Lshortfile
	logger := log.New(os.Stdout, opts.Prefix, flags)
	return &Logger{
		level:  opts.Level,
		logger: logger,
	}
}

func (l *Logger) log(level Level, prefix, msg string, args ...any) {
	if level < l.level {
		return
	}
	if len(args) > 0 {
		msg = fmt.Sprintf("%s %s %v", prefix, msg, args)
	} else {
		msg = fmt.Sprintf("%s %s", prefix, msg)
	}
	l.logger.Output(3, msg)
}

func (l *Logger) Info(msg string, args ...any) {
	l.log(LevelInfo, "[INFO]", msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.log(LevelWarn, "[WARN]", msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.log(LevelError, "[ERROR]", msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log(LevelDebug, "[DEBUG]", msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.log(LevelError, "[FATAL]", msg, args...)
	os.Exit(1)
}

func (l *Logger) With(prefix string) *Logger {
	return &Logger{
		level:  l.level,
		logger: log.New(l.logger.Writer(), prefix+" ", l.logger.Flags()),
	}
}
