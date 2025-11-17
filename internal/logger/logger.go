package logger

import (
	"log"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

type Logger struct {
	l     *log.Logger
	level Level
}

func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func New(levelStr string) *Logger {
	l := log.New(os.Stdout, "[goproxy] ", log.LstdFlags|log.Lshortfile)
	return &Logger{
		l:     l,
		level: ParseLevel(levelStr),
	}
}

func (lg *Logger) Debugf(format string, v ...any) {
	if lg.level <= LevelDebug {
		lg.l.Printf("[DEBUG] "+format, v...)
	}
}

func (lg *Logger) Infof(format string, v ...any) {
	if lg.level <= LevelInfo {
		lg.l.Printf("[INFO] "+format, v...)
	}
}

func (lg *Logger) Errorf(format string, v ...any) {
	lg.l.Printf("[ERROR] "+format, v...)
}
