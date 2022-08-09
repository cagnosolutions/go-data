package logging

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	White = iota
	Black = iota + 30
	Red
	Green
	Yellow
	Blue
	Purple
	Cyan
	Grey
)

const (
	color = iota
	prefix
)

var colors = map[int]string{
	White:  "\033[0m",
	Black:  "\033[30m",
	Red:    "\033[31m",
	Green:  "\033[32m",
	Yellow: "\033[33m",
	Blue:   "\033[34m",
	Purple: "\033[35m",
	Cyan:   "\033[36m",
	Grey:   "\033[37m",
}

var levelColor = map[Level][2]string{
	LevelTrace: {colors[Grey], "TRCE"},
	LevelDebug: {colors[Grey], "DBUG"},
	LevelInfo:  {colors[Blue], "INFO"},
	LevelWarn:  {colors[Yellow], "WARN"},
	LevelError: {colors[Red], "EROR"},
	LevelFatal: {colors[Red], "FATL"},
	LevelPanic: {colors[Red], "PANC"},
}

type cLogger struct {
	lock      sync.Mutex    // sync
	log       *log.Logger   // actual logger
	buf       *bytes.Buffer // buffer
	printFunc bool
	printFile bool
	dep       int // call depth
}

func NewCLogger() *cLogger {
	_ = log.Ldate | log.Ltime | log.Lshortfile | log.Lmsgprefix
	l := &cLogger{
		log: log.New(os.Stderr, "", log.LstdFlags),
		buf: new(bytes.Buffer),
		dep: 5,
	}
	return l
}

func (l *cLogger) formatLog(level Level, depth int, format string, args ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	levelInfo, ok := levelColor[level]
	if !ok {
		levelInfo = levelColor[LevelInfo]
	}
	l.buf.Reset()
	l.buf.WriteString("| ")
	l.buf.WriteString(levelInfo[color])
	l.buf.WriteString(levelInfo[prefix])
	l.buf.WriteString(colors[White])
	l.buf.WriteString(" | ")
	if l.printFunc || l.printFile {
		if level == LevelFatal {
			depth += 1
		}
		if level != LevelPanic {
			fn, file := trace(depth)
			if l.printFunc {
				l.buf.WriteByte('[')
				l.buf.WriteString(strings.Split(fn, ".")[1])
				l.buf.WriteByte(']')
			}
			if l.printFunc && l.printFile {
				l.buf.WriteByte(' ')
			}
			if l.printFile {
				l.buf.WriteString(file)
			}
			l.buf.WriteString(" - ")
		}
	}
	l.buf.WriteString(format)
	if args == nil || len(args) == 0 {
		l.log.Print(l.buf.String())
		return
	}
	l.log.Printf(l.buf.String(), args...)
}

func (l *cLogger) SetPrefix(prefix string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.log.SetPrefix(prefix)
}

func (l *cLogger) SetPrintFunc(ok bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.printFunc = ok
}

func (l *cLogger) SetPrintFile(ok bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.printFile = ok
}

func (l *cLogger) SetCallDepth(depth int) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.dep = depth
}

func (l *cLogger) Trace(message string) {
	l.formatLog(LevelTrace, l.dep, message)
}

func (l *cLogger) Tracef(format string, args ...interface{}) {
	l.formatLog(LevelTrace, l.dep, format, args...)
}

func (l *cLogger) Debug(message string) {
	l.formatLog(LevelDebug, l.dep, message)
}

func (l *cLogger) Debugf(format string, args ...interface{}) {
	l.formatLog(LevelDebug, l.dep, format, args...)
}

func (l *cLogger) Info(message string) {
	l.formatLog(LevelInfo, l.dep, message)
}

func (l *cLogger) Infof(format string, args ...interface{}) {
	l.formatLog(LevelInfo, l.dep, format, args...)
}

func (l *cLogger) Warn(message string) {
	l.formatLog(LevelWarn, l.dep, message)
}

func (l *cLogger) Warnf(format string, args ...interface{}) {
	l.formatLog(LevelWarn, l.dep, format, args...)
}

func (l *cLogger) Error(message string) {
	l.formatLog(LevelError, l.dep, message)
}

func (l *cLogger) Errorf(format string, args ...interface{}) {
	l.formatLog(LevelError, l.dep, format, args...)
}

func (l *cLogger) Fatal(message string) {
	l.formatLog(LevelFatal, l.dep, message)
	os.Exit(1)
}

func (l *cLogger) Fatalf(format string, args ...interface{}) {
	l.formatLog(LevelFatal, l.dep, format, args...)
	os.Exit(1)
}

func (l *cLogger) Panic(message string) {
	l.formatLog(LevelPanic, l.dep, message)
	panic(message)
}

func (l *cLogger) Panicf(format string, args ...interface{}) {
	l.formatLog(LevelPanic, l.dep, format, args...)
	panic(fmt.Sprintf(format, args...))
}

func trace0(calldepth int) (string, string) {
	// pc, file, line, ok := runtime.Caller(calldepth)
	pc, file, line, _ := runtime.Caller(calldepth)
	fn := runtime.FuncForPC(pc)
	// funcName := filepath.Base(fn.Name())
	// fileName := filepath.Base(file)
	// return fmt.Sprintf("%s %s:%d", funcName, fileName, line)
	return filepath.Base(fn.Name()), fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// trace is used internally to get more information
func trace(calldepth int) (string, string) {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(calldepth, pc)
	fn := runtime.FuncForPC(pc[0])
	file, line := fn.FileLine(pc[0])
	return filepath.Base(fn.Name()), fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
