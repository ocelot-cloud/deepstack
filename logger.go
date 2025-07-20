package utils

import (
	"context"
	"fmt"
	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type DeepStackLogger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
	NewError(msg string, kv ...any) error
}

// idea for later: add the software version to the log so that "source" attribute deterministally references its origin
func NewDeepStackLogger(logLevel string, showCaller bool) DeepStackLogger {
	logDir := "data/logs"
	if err := os.MkdirAll(logDir, 0700); err != nil {
		panic(fmt.Sprintf("Failed to create logs directory: %v", err))
	}

	logFile := &lumberjack.Logger{
		Filename:   logDir + "/app.log",
		MaxSize:    100,
		MaxBackups: 0,
		MaxAge:     30,
		Compress:   true,
	}

	slogLogLevel := convertToSlogLevel(logLevel)

	opts := &slog.HandlerOptions{
		AddSource:   showCaller,
		Level:       slogLogLevel,
		ReplaceAttr: replaceSource,
	}

	fileHandler := slog.NewJSONHandler(logFile, opts)
	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:   showCaller,
		Level:       slogLogLevel,
		ReplaceAttr: dropStackTrace,
	})

	logger := slog.New(multiHandler{fileHandler, consoleHandler})
	return &DeepStackLoggerImpl{logger, &LoggingBackendImpl{slog: logger}}
}

func dropStackTrace(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "stack_trace" {
		return slog.Attr{}
	}
	return replaceSource(groups, a)
}

func replaceSource(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		src := a.Value.Any().(*slog.Source)
		if rel, ok := strings.CutPrefix(src.File, workDirectory+string(os.PathSeparator)); ok {
			src.File = rel
		} else {
			src.File = filepath.Base(src.File)
		}
		return slog.Any(a.Key, src)
	}
	return a
}

type multiHandler []slog.Handler

func (h multiHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	for _, hd := range h {
		if hd.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}
func (h multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, hd := range h {
		_ = hd.Handle(ctx, r)
	}
	return nil
}
func (h multiHandler) WithAttrs(a []slog.Attr) slog.Handler {
	out := make(multiHandler, len(h))
	for i, hd := range h {
		out[i] = hd.WithAttrs(a)
	}
	return out
}

func (h multiHandler) WithGroup(name string) slog.Handler {
	out := make(multiHandler, len(h))
	for i, hd := range h {
		out[i] = hd.WithGroup(name)
	}
	return out
}

type DeepStackLoggerImpl struct {
	l      *slog.Logger // TODO to be removed, use LoggingBackend instead
	logger LoggingBackend
}

// TODO this should be unit tested using mockery;
func (m *DeepStackLoggerImpl) log(level string, msg string, kv ...any) {
	if m.logger.ShouldLogBeSkipped(level) {
		return
	}

	rec := m.logger.CreateLogRecord(level, msg)
	var stackTrace string

	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			m.Warn("invalid key type in log message, must always be string", "key", key)
			continue
		}

		if key == ErrorField {
			detailedError, ok := kv[i+1].(*DeepStackError)
			if ok {
				for k, v := range detailedError.Context {
					rec.AddAttrs(k, v)
				}
				rec.AddAttrs("stack_trace", detailedError.ErrorStack)
				stackTrace = detailedError.ErrorStack
				m.log(level, msg)
			} else {
				m.Warn("invalid error type in log message, must be *DeepStackError")
				rec.AddAttrs(key, kv[i+1])
			}
		} else {
			rec.AddAttrs(key, kv[i+1])
		}
	}
	m.logger.HandleRecord(rec)
	if stackTrace != "" {
		m.logger.Println(stackTrace)
	}
}

func (m *DeepStackLoggerImpl) Debug(msg string, kv ...any) { m.log("debug", msg, kv...) }
func (m *DeepStackLoggerImpl) Info(msg string, kv ...any)  { m.log("info", msg, kv...) }
func (m *DeepStackLoggerImpl) Warn(msg string, kv ...any)  { m.log("warn", msg, kv...) }
func (m *DeepStackLoggerImpl) Error(msg string, kv ...any) { m.log("error", msg, kv...) }
func (m *DeepStackLoggerImpl) NewError(msg string, kv ...any) error {
	var contextMap = make(map[string]any)
	for i := 0; i+1 < len(kv); i += 2 {
		if k, ok := kv[i].(string); ok {
			contextMap[k] = kv[i+1]
		}
	}

	return &DeepStackError{
		ErrorMessage: msg,
		ErrorStack:   printStackTrace(),
		Context:      contextMap,
	}
}
