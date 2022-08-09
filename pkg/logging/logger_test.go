package logging

import (
	"fmt"
	"os"
	"testing"
)

// LevelTrace Level = iota
// LevelDebug
// LevelInfo
// LevelWarn
// LevelError
// LevelFatal
// LevelPanic
// LevelDefault
// LevelOff

func TestLevelText(t *testing.T) {

	levels := []Level{
		LevelTrace,
		LevelDebug,
		LevelInfo,
		LevelWarn,
		LevelError,
		LevelFatal,
		LevelPanic,
		LevelOff,
	}

	for _, level := range levels {
		l := NewLogger(os.Stderr, level)
		l.Printf("New Logger %s\n", LevelText(level))
		l.Trace(">> TRACE level log I am printing (Current %s)\n", LevelText(level))
		l.Debug(">> DEBUG level log I am printing (Current %s)\n", LevelText(level))
		l.Info(">> INFO level log I am printing (Current %s)\n", LevelText(level))
		l.Warn(">> WARNING level log I am printing (Current %s)\n", LevelText(level))
		l.Error(">> ERROR level log I am printing (Current %s)\n", LevelText(level))
		fmt.Println()
	}

}

func ARandomFunction() {
	// this function doesn't really do anything

}
