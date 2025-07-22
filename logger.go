package deepstack

import (
	"context"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"
)

type DeepStackLogger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
	NewError(msg string, kv ...any) error
}

// idea for later: add the software version to the log so that "source" attribute deterministally references its origin
func NewDeepStackLogger(logLevel string, enableWarningsForNonDeepStackErrors bool) DeepStackLogger {
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
		AddSource: true,
		Level:     slogLogLevel,
	}

	fileHandler := slog.NewJSONHandler(logFile, opts)
	consoleHandler := leanConsoleHandler{w: os.Stdout}

	logger := slog.New(multiHandler{fileHandler, consoleHandler})
	return &DeepStackLoggerImpl{
		logger:                              &LoggingBackendImpl{slog: logger},
		enableWarningsForNonDeepStackErrors: enableWarningsForNonDeepStackErrors,
	}
}

type coloredConsoleHandler struct {
	slog.Handler
	w io.Writer
}

var lvlColor = map[slog.Level]string{
	slog.LevelDebug: "\x1b[36m", // cyan
	slog.LevelInfo:  "\x1b[32m", // green
	slog.LevelWarn:  "\x1b[33m", // yellow
	slog.LevelError: "\x1b[31m", // red
}

func (h *coloredConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	if c := lvlColor[r.Level]; c != "" {
		_, _ = io.WriteString(h.w, c)
		err := h.Handler.Handle(ctx, r)
		_, _ = io.WriteString(h.w, "\x1b[0m")
		return err
	}
	return h.Handler.Handle(ctx, r)
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
	logger                              LoggingBackend
	enableWarningsForNonDeepStackErrors bool
}

func (m *DeepStackLoggerImpl) log(level string, msg string, keyValuePairs ...any) {
	if m.logger.ShouldLogBeSkipped(level) {
		return
	}

	record := m.logger.CreateLogRecord(level, msg)
	var stackTrace string
	for i := 0; i+1 < len(keyValuePairs); i += 2 {
		key, ok := keyValuePairs[i].(string)
		if !ok {
			m.logger.LogWarning("invalid key type in log message, must always be string", "type", reflect.TypeOf(keyValuePairs[i]).String())
			continue // TODO can be removed without causing tests to fail, fix this
		}

		value := keyValuePairs[i+1]
		if key == ErrorField {
			stackTrace = m.handleErrorField(record, key, value)
		} else {
			record.AddAttrs(key, value)
		}
	}
	m.logger.HandleRecord(record)
	if stackTrace != "" {
		m.logger.Println(stackTrace)
	}
}

func (m *DeepStackLoggerImpl) handleErrorField(record *LogRecord, key string, value any) string {
	detailedError, ok := value.(*DeepStackError)
	if ok {
		for contextKey, contextValue := range detailedError.Context {
			record.AddAttrs(contextKey, contextValue)
		}
		record.AddAttrs("stack_trace", detailedError.StackTrace)
		return detailedError.StackTrace
	} else {
		if m.enableWarningsForNonDeepStackErrors {
			m.logger.LogWarning("invalid error type in log message, must be *DeepStackError")
		}
		record.AddAttrs(key, value)
		return ""
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
		Message:    msg,
		StackTrace: printStackTrace(),
		Context:    contextMap,
	}
}

func (m *DeepStackLoggerImpl) AddContext(err error) error {
	/* TODO implement and unit test
	if this is a DeepStackError, add the context to it, and return the error
	else log a warning that this is not a DeepStackError, convert it to a DeepStackError, add context and return it
	*/
	return nil
}

type leanConsoleHandler struct{ w io.Writer }

func (h leanConsoleHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h leanConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
	fileLine := fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
	var attrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == slog.SourceKey {
			return true
		}
		if a.Key == "stack_trace" {
			return true
		}
		attrs = append(attrs, a)
		return true
	})

	c, reset := lvlColor[r.Level], "\x1b[0m"
	fmt.Fprintf(h.w, "%s%s %s %s %q", c, r.Time.Format(time.RFC3339Nano), r.Level, fileLine, r.Message)
	for _, a := range attrs {
		fmt.Fprintf(h.w, " %s=%v", a.Key, a.Value)
	}
	fmt.Fprintln(h.w, reset)
	return nil
}

func (h leanConsoleHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h leanConsoleHandler) WithGroup(string) slog.Handler      { return h }
