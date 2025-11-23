package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	// LevelDebug is for detailed debugging information.
	LevelDebug LogLevel = iota
	// LevelInfo is for general informational messages.
	LevelInfo
	// LevelWarn is for warning messages.
	LevelWarn
	// LevelError is for error messages.
	LevelError
)

var (
	levelNames = map[LogLevel]string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
	}
	levelColors = map[LogLevel]string{
		LevelDebug: "\033[36m", // Cyan
		LevelInfo:  "\033[32m", // Green
		LevelWarn:  "\033[33m", // Yellow
		LevelError: "\033[31m", // Red
	}
	colorReset = "\033[0m"
)

// Logger provides structured logging with levels.
type Logger struct {
	mu        sync.Mutex
	output    io.Writer
	minLevel  LogLevel
	useColors bool
	prefix    string
	stdLogger *log.Logger
}

// NewLogger creates a new Logger instance.
func NewLogger(output io.Writer, minLevel LogLevel, useColors bool, prefix string) *Logger {
	return &Logger{
		output:    output,
		minLevel:  minLevel,
		useColors: useColors,
		prefix:    prefix,
		stdLogger: log.New(output, "", 0),
	}
}

// NewDefaultLogger creates a logger with sensible defaults for CLI usage.
func NewDefaultLogger(minLevel LogLevel) *Logger {
	return NewLogger(os.Stderr, minLevel, true, "")
}

// SetLevel changes the minimum log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// SetPrefix changes the logger prefix.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// log is the internal logging function.
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.minLevel {
		return
	}

	levelName := levelNames[level]
	message := fmt.Sprintf(format, args...)

	var output string
	if l.useColors {
		color := levelColors[level]
		output = fmt.Sprintf("%s%-5s%s ", color, levelName, colorReset)
	} else {
		output = fmt.Sprintf("%-5s ", levelName)
	}

	if l.prefix != "" {
		output += fmt.Sprintf("[%s] ", l.prefix)
	}

	output += message

	l.stdLogger.Println(output)
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal logs an error message and exits with status 1.
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
	os.Exit(1)
}

// ParseLogLevel parses a string into a LogLevel.
func ParseLogLevel(s string) (LogLevel, error) {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level: %s", s)
	}
}

// DefaultLogger is the global logger instance used by package-level logging functions.
var DefaultLogger = NewDefaultLogger(LevelInfo)

// Debug logs a debug message to the default logger.
func Debug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

// Info logs an info message to the default logger.
func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

// Warn logs a warning message to the default logger.
func Warn(format string, args ...interface{}) {
	DefaultLogger.Warn(format, args...)
}

// Error logs an error message to the default logger.
func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

// Fatal logs an error message to the default logger and exits.
func Fatal(format string, args ...interface{}) {
	DefaultLogger.Fatal(format, args...)
}

// SetLevel sets the minimum log level for the default logger.
func SetLevel(level LogLevel) {
	DefaultLogger.SetLevel(level)
}
