package utils

import (
	"fmt"
	"log"
	"os"
)

type LogLevel = int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelOff
)

func LevelText(level LogLevel) string {
	switch level {
	case LevelDebug:
		return "Level=Debug"
	case LevelInfo:
		return "Level=Info"
	case LevelWarn:
		return "Level=Warn"
	case LevelError:
		return "Level=Error"
	case LevelFatal:
		return "Level=Fatal"
	case LevelOff:
		return "Level=Off"
	default:
		return "Level=Unknown"
	}
}

type Logger struct {
	*log.Logger
	level LogLevel
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
		level:  level,
	}
}

func (l *Logger) Debug(s string, a ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	ls := fmt.Sprintf("| DEBUG | %s", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

func (l *Logger) Info(s string, a ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	ls := fmt.Sprintf("|  INFO | %s", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

func (l *Logger) Warn(s string, a ...interface{}) {
	if l.level > LevelWarn {
		return
	}
	ls := fmt.Sprintf("|  WARN | %s", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

func (l *Logger) Error(s string, a ...interface{}) {
	if l.level > LevelError {
		return
	}
	ls := fmt.Sprintf("| ERROR | %s", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

func (l *Logger) Fatal(s string, a ...interface{}) {
	if l.level > LevelFatal {
		return
	}
	ls := fmt.Sprintf("| FATAL | %s", s)
	if a == nil || len(a) == 0 {
		l.Fatalln(ls)
	}
	l.Fatalf(ls, a...)
}
