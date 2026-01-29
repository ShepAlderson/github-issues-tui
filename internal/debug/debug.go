package debug

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	logFile *os.File
	mu      sync.Mutex
	enabled = false
)

// Enable enables debug logging
func Enable() error {
	mu.Lock()
	defer mu.Unlock()

	if enabled {
		return nil
	}

	f, err := os.OpenFile("/tmp/ghissues_debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	logFile = f
	enabled = true
	// Log directly without calling Log() to avoid deadlock
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(logFile, "[%s] Debug logging enabled\n", timestamp)
	return nil
}

// Log writes a log message
func Log(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if !enabled || logFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "[%s] %s\n", timestamp, msg)
}

// Close closes the log file
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	enabled = false
}
