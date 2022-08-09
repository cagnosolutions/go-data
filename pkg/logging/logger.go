package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// For more information on leveled logging [https://sematext.com/blog/logging-levels/]

// Level is a type.
type Level = uint8

// The logging levels are listed in order below.
const (
	_          = iota
	LevelTrace // A log level describing detailed events; useful during extended debugging sessions.
	LevelDebug // Used for events considered to be useful during software debugging.
	LevelInfo  // Purely informative and can be ignored during normal operations.
	LevelWarn  // Unexpected behavior happened, but key features are still operating as expected.
	LevelError // Something occurred that will prevent some features from working correctly.
	LevelFatal // One or more key features are not working properly; system might have stopped completely.
	LevelPanic // Usually used during a failure that should not have happened; system may have crashed.
	LevelOff   // Disables all logging output
)

// LevelText accepts a logLevel, and it will return the corresponding string value.
func LevelText(level Level) string {
	switch level {
	case LevelTrace:
		return "Level=Trace"
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
	case LevelPanic:
		return "Level=Panic"
	case LevelOff:
		return "Level=Off"
	default:
		return "Level=Unknown"
	}
}

// DefaultLogger is the pre-instantiated default logger
var (
	DefaultLevel     Level = LevelInfo
	DefaultLogger          = NewLogger(os.Stderr, DefaultLevel)
	DefaultFlags           = log.Ldate | log.Ltime | log.Lshortfile | log.Lmsgprefix
	defaultCallDepth       = 2
)

// Logger is used for logging
type Logger struct {
	*log.Logger
	level Level
	depth int
}

// NewDefaultLogger returns the default logger, which is set up to log at the info level.
func NewDefaultLogger() *Logger {
	return NewLogger(os.Stderr, DefaultLevel)
}

// NewLogger instantiates a new instance of *Logger. It accepts a log level and will ignore any
// log level above the level provided.
func NewLogger(w io.Writer, level Level) *Logger {
	return &Logger{
		Logger: log.New(w, "", log.LstdFlags),
		level:  level,
		depth:  defaultCallDepth,
	}
}

// SetCallDepth adjusts the call depth, which alters the info printed by a trace call.
func (l *Logger) SetCallDepth(depth int) {
	l.depth = depth
}

// Trace writes a log to standard output at the trace level, displaying the "| TRACE |" prefix.
func (l *Logger) Trace(s string, a ...interface{}) {
	if l.level > LevelTrace {
		return
	}
	ls := l.formatLog("| TRACE |", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

// Debug writes a log to standard output at the debug level, displaying the "| DEBUG |" prefix.
func (l *Logger) Debug(s string, a ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	ls := l.formatLog("| DEBUG |", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

// Info writes a log to standard output at the info level, displaying the "| INFO |" prefix.
func (l *Logger) Info(s string, a ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	ls := l.formatLog("|  INFO |", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

// Warn writes a log to standard output at the warning level displaying the "| WARN |" prefix.
func (l *Logger) Warn(s string, a ...interface{}) {
	if l.level > LevelWarn {
		return
	}
	ls := l.formatLog("|  WARN |", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

// Error writes a log to standard output at the error level displaying the "| ERROR |" prefix.
func (l *Logger) Error(s string, a ...interface{}) {
	if l.level > LevelError {
		return
	}
	ls := l.formatLog("| ERROR |", s)
	if a == nil || len(a) == 0 {
		l.Println(ls)
		return
	}
	l.Printf(ls, a...)
}

// Fatal writes a log to standard output at the fatal level displaying the "| FATAL |" prefix.
func (l *Logger) Fatal(s string, a ...interface{}) {
	if l.level > LevelFatal {
		return
	}
	ls := l.formatLog("| FATAL |", s)
	if a == nil || len(a) == 0 {
		l.Fatalln(ls)
	}
	l.Fatalf(ls, a...)
}

// Panic writes a log to standard output at the panic level displaying the "| PANIC |" prefix.
func (l *Logger) Panic(s string, a ...interface{}) {
	if l.level > LevelPanic {
		return
	}
	ls := l.formatLog("| PANIC |", s)
	if a == nil || len(a) == 0 {
		l.Panicln(ls)
	}
	l.Panicf(ls, a...)
}

// Default writes a log to standard output at the default level.
func (l *Logger) Default(s string, a ...interface{}) {
	l.Info(s, a...)
}

func (l *Logger) formatLog(prefix string, s string) string {
	pc, file, line, ok := runtime.Caller(l.depth)
	if !ok || l.depth < 2 {
		return fmt.Sprintf("%s %s", prefix, s)
	}
	fn := runtime.FuncForPC(pc)
	funcName := filepath.Base(fn.Name())
	fileName := filepath.Base(file)
	return fmt.Sprintf("%s [%s] %s:%d - %s", prefix, funcName, fileName, line, s)
}
