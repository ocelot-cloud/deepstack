package deepstack

import (
	"io"
	"log/slog"
	"os"
	"reflect"

	"gopkg.in/natefinch/lumberjack.v2"
)

type DeepStackLogger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
	NewError(msg string, kv ...any) error
}

// TODO I dont like that my business logic is coupled with slog. Instead it should be completely hidden behind an interface? not sure, maybe slog.Handler interface is tolerable as dependency
func newDeepStackLoggerWithCustomWriter(logLevel string, enableWarningsForNonDeepStackErrors bool, dst io.Writer) DeepStackLogger {
	slogLogger := getSlogLogger(logLevel, dst)
	return &DeepStackLoggerImpl{
		logger:               &LoggingBackendImpl{slog: slogLogger},
		enableMisuseWarnings: enableWarningsForNonDeepStackErrors,
	}
}

func getSlogLogger(logLevel string, dst io.Writer) *slog.Logger {
	// TODO nil should be rejected?
	if dst == nil {
		dst = os.Stdout
	}

	// TODO this block should be contained in the constructor block for the production log file writer
	logDir := "data/logs"
	_ = os.MkdirAll(logDir, 0700)
	logFile := &lumberjack.Logger{Filename: logDir + "/app.log", MaxSize: 100, MaxAge: 30, Compress: true}
	opts := &slog.HandlerOptions{AddSource: true, Level: convertToSlogLevel(logLevel)}
	fileHandler := slog.NewJSONHandler(logFile, opts)

	consoleHandlerObj := consoleHandler{w: dst, opts: opts}
	logger := slog.New(multiHandler{fileHandler, consoleHandlerObj})
	return logger
}

func NewDeepStackLogger(logLevel string, enableWarningsForNonDeepStackErrors bool) DeepStackLogger {
	return newDeepStackLoggerWithCustomWriter(logLevel, enableWarningsForNonDeepStackErrors, os.Stdout)
}

var lvlColor = map[slog.Level]string{
	slog.LevelDebug: "\x1b[36m", // cyan
	slog.LevelInfo:  "\x1b[32m", // green
	slog.LevelWarn:  "\x1b[33m", // yellow
	slog.LevelError: "\x1b[31m", // red
}

type DeepStackLoggerImpl struct {
	logger LoggingBackend
	// if enabled, when the framework is not used correctly, a warning is logged
	enableMisuseWarnings bool
}

func (m *DeepStackLoggerImpl) log(level string, msg string, keyValuePairs ...any) {
	if m.logger.ShouldLogBeSkipped(level) {
		return
	}

	record := &Record{
		level:      level,
		msg:        msg,
		attributes: make(map[string]any),
	}
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
	m.logger.LogRecord(record)
	if stackTrace != "" {
		m.logger.Println(stackTrace)
	}
}

func (m *DeepStackLoggerImpl) handleErrorField(record *Record, key string, value any) string {
	detailedError, ok := value.(*DeepStackError)
	if ok {
		for contextKey, contextValue := range detailedError.Context {
			record.AddAttrs(contextKey, contextValue)
		}
		record.AddAttrs("stack_trace", detailedError.StackTrace)
		record.AddAttrs("error_cause", detailedError.Message)
		return detailedError.StackTrace
	} else {
		if m.enableMisuseWarnings {
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

func (m *DeepStackLoggerImpl) AddContext(err error, context ...any) error {
	workError, ok := err.(*DeepStackError)
	if ok {
		m.addToContextField(context, workError)
		return workError
	} else {
		if m.enableMisuseWarnings {
			m.logger.LogWarning("invalid error type in log message, must be *DeepStackError")
		}
		deepStackError := &DeepStackError{
			Message:    err.Error(),
			StackTrace: printStackTrace(),
			Context:    map[string]any{},
		}
		m.addToContextField(context, deepStackError)
		return deepStackError
	}
}

func (m *DeepStackLoggerImpl) addToContextField(context []any, workError *DeepStackError) {
	for i := 0; i+1 < len(context); i += 2 {
		if key, ok := context[i].(string); ok {
			workError.Context[key] = context[i+1]
		} else {
			m.logger.LogWarning("invalid key type in log message, must always be string", "type", reflect.TypeOf(context[i]).String())
		}
	}
}
