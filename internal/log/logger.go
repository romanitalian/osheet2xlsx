package log

import (
	"encoding/json"
	"io"
	stdlog "log"
	"os"
	"strings"
	"time"
)

// Level represents logging verbosity.
type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

// Logger provides minimal leveled logging without external deps.
type Logger struct {
	level   Level
	base    *stdlog.Logger
	json    bool
	noColor bool
}

var global *Logger

// SetupLogger configures the global logger.
func SetupLogger(levelStr string, quiet bool, noColor bool, json bool) error {
	var lvl Level
	switch strings.ToLower(levelStr) {
	case "error":
		lvl = LevelError
	case "warn":
		lvl = LevelWarn
	case "info":
		lvl = LevelInfo
	case "debug":
		lvl = LevelDebug
	default:
		lvl = LevelInfo
	}

	var output io.Writer = os.Stderr
	if quiet {
		output = io.Discard
	}

	flag := stdlog.LstdFlags
	if json {
		flag = 0
	}
	global = &Logger{
		level:   lvl,
		base:    stdlog.New(output, "", flag),
		json:    json,
		noColor: noColor,
	}
	return nil
}

// Get returns the global logger.
func Get() *Logger {
	if global == nil {
		if err := SetupLogger("info", false, false, false); err != nil {
			// Fallback to basic logger if setup fails
			global = &Logger{
				level:   LevelInfo,
				base:    stdlog.New(os.Stderr, "", stdlog.LstdFlags),
				json:    false,
				noColor: false,
			}
		}
	}
	return global
}

func (l *Logger) log(at Level, message string) {
	if at > l.level {
		return
	}
	if l.json {
		type record struct {
			Time  string `json:"time"`
			Level string `json:"level"`
			Msg   string `json:"msg"`
		}
		lvl := "info"
		if at == LevelError {
			lvl = "error"
		} else if at == LevelWarn {
			lvl = "warn"
		} else if at == LevelDebug {
			lvl = "debug"
		}
		rec := record{Time: time.Now().UTC().Format(time.RFC3339), Level: lvl, Msg: message}
		b, err := json.Marshal(rec)
		if err == nil {
			l.base.Print(string(b))
			return
		}
	}
	l.base.Print(message)
}

func (l *Logger) Error(message string) { l.log(LevelError, "ERROR: "+message) }
func (l *Logger) Warn(message string)  { l.log(LevelWarn, "WARN: "+message) }
func (l *Logger) Info(message string)  { l.log(LevelInfo, "INFO: "+message) }
func (l *Logger) Debug(message string) { l.log(LevelDebug, "DEBUG: "+message) }
