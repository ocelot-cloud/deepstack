package deepstack

import (
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	actualTypeField = "actual_type"
	keyField        = "key"

	emptySpacesInKeyMessage      = "spaces in keys are not allowed"
	invalidKeyTypeMessage        = "invalid key type in log message, must always be string"
	invalidErrorTypeMessage      = "invalid error type in log message, must be *DeepStackError"
	oddKeyValuePairNumberMessage = "odd number of key-value pairs in log message, must always be even"
)

type DeepStackLogger interface {
	Debug(msg string, context ...any)
	Info(msg string, context ...any)
	Warn(msg string, context ...any)
	Error(msg string, context ...any)
	NewError(msg string, context ...any) error
	AddContext(err error, context ...any) error
}

func NewConsoleHandler(opts *slog.HandlerOptions) *ConsoleHandler {
	return &ConsoleHandler{w: os.Stdout, opts: opts}
}

func NewDeepStackLogger(logLevel string, additionalHandlers ...slog.Handler) DeepStackLogger {
	opts := &slog.HandlerOptions{AddSource: true, Level: convertToSlogLevel(logLevel)}
	consoleHandlerObject := NewConsoleHandler(opts)
	multiHandlerObject := multiHandler{hs: append([]slog.Handler{consoleHandlerObject}, additionalHandlers...)}
	slogLogger := slog.New(multiHandlerObject)
	return &DeepStackLoggerImpl{
		logger:      &LoggingBackendImpl{slog: slogLogger},
		stackTracer: &StackTracerImpl{},
	}
}

var lvlColor = map[slog.Level]string{
	slog.LevelDebug: "\x1b[36m", // cyan
	slog.LevelInfo:  "\x1b[32m", // green
	slog.LevelWarn:  "\x1b[33m", // yellow
	slog.LevelError: "\x1b[31m", // red
}

type DeepStackLoggerImpl struct {
	logger      LoggingBackend
	stackTracer StackTracer
}

func (m *DeepStackLoggerImpl) log(level string, msg string, context ...any) {
	if m.logger.ShouldLogBeSkipped(level) {
		return
	}
	sanitizedContext := m.sanitizeContext(context)

	record := &Record{
		level:      level,
		msg:        msg,
		attributes: make(map[string]any),
	}
	var stackTrace string
	for key, value := range sanitizedContext {
		if key == ErrorField {
			stackTrace = m.appendStackErrorToRecord(record, key, value)
		} else {
			record.AddAttrs(key, value)
		}
	}
	m.logger.LogRecord(record)
	if stackTrace != "" {
		m.logger.PrintStackTrace(stackTrace)
	}
}

func (m *DeepStackLoggerImpl) sanitizeContext(context []any) map[string]any {
	if len(context)%2 != 0 {
		m.logger.LogWarning(oddKeyValuePairNumberMessage)
	}

	result := make(map[string]any)
	for i := 0; i+1 < len(context); i += 2 {
		if key, ok := context[i].(string); ok {
			if strings.Contains(key, " ") {
				m.logger.LogWarning(emptySpacesInKeyMessage, keyField, key)
				continue
			}
			result[key] = context[i+1]
		} else {
			m.logger.LogWarning(invalidKeyTypeMessage, actualTypeField, reflect.TypeOf(context[i]).String())
		}
	}
	return result
}

func (m *DeepStackLoggerImpl) appendStackErrorToRecord(record *Record, key string, value any) string {
	detailedError, ok := value.(*DeepStackError)
	if ok {
		for contextKey, contextValue := range detailedError.Context {
			record.AddAttrs(contextKey, contextValue)
		}
		record.AddAttrs("stack_trace", detailedError.StackTrace)
		record.AddAttrs("error_cause", detailedError.Message)
		return detailedError.StackTrace
	} else {
		m.logger.LogWarning(invalidErrorTypeMessage)
		record.AddAttrs(key, value)
		return ""
	}
}

func (m *DeepStackLoggerImpl) Debug(msg string, context ...any) { m.log("debug", msg, context...) }
func (m *DeepStackLoggerImpl) Info(msg string, context ...any)  { m.log("info", msg, context...) }
func (m *DeepStackLoggerImpl) Warn(msg string, context ...any)  { m.log("warn", msg, context...) }
func (m *DeepStackLoggerImpl) Error(msg string, context ...any) { m.log("error", msg, context...) }
func (m *DeepStackLoggerImpl) NewError(msg string, context ...any) error {
	var contextMap = make(map[string]any)
	for i := 0; i+1 < len(context); i += 2 {
		if k, ok := context[i].(string); ok {
			contextMap[k] = context[i+1]
		}
	}

	return &DeepStackError{
		Message:    msg,
		StackTrace: m.stackTracer.GetStackTrace(),
		Context:    contextMap,
	}
}

func (m *DeepStackLoggerImpl) AddContext(err error, context ...any) error {
	sanitizedContext := m.sanitizeContext(context)
	deepStackError, ok := err.(*DeepStackError)
	if ok {
		m.addToContextField(sanitizedContext, deepStackError)
		return deepStackError
	} else {
		m.logger.LogWarning(invalidErrorTypeMessage)
		newDeepStackError := &DeepStackError{
			Message:    err.Error(),
			StackTrace: m.stackTracer.GetStackTrace(),
			Context:    map[string]any{},
		}
		m.addToContextField(sanitizedContext, newDeepStackError)
		return newDeepStackError
	}
}

func (m *DeepStackLoggerImpl) addToContextField(sanitizedContext map[string]any, deepStackError *DeepStackError) {
	for key, value := range sanitizedContext {
		deepStackError.Context[key] = value
	}
}

func AssertDeepStackError(t *testing.T, err error, expectedMessage string, expectedContext ...any) {
	t.Helper()

	require.NotNilf(t, err, "expected error %q, got nil", expectedMessage)

	deepErr, ok := err.(*DeepStackError)
	require.Truef(t, ok, "expected *DeepStackError, got %T (%v)", err, err)

	assert.Equal(t, expectedMessage, deepErr.Message, "message mismatch")

	require.Equalf(t, 0, len(expectedContext)%2, "expectedContext must be key/value pairs, got odd number of item: %d", len(expectedContext))

	expectedKeyValuePairs := make(map[string]any, len(expectedContext)/2)
	for i := 0; i < len(expectedContext); i += 2 {
		k, ok := expectedContext[i].(string)
		require.Truef(t, ok, "expectedContext key at index %d is not string: %T", i, expectedContext[i])
		expectedKeyValuePairs[k] = expectedContext[i+1]
	}

	for expectedKey, expectedValue := range expectedKeyValuePairs {
		actualValue, found := deepErr.Context[expectedKey]
		assert.Truef(t, found, "missing context key %q (expected %#v)", expectedKey, expectedValue)
		if found {
			assert.Equalf(t, expectedValue, actualValue, "context value mismatch for key %q", expectedKey)
		}
	}

	for errorContextKey, errorContextValue := range deepErr.Context {
		if _, ok := expectedKeyValuePairs[errorContextKey]; !ok {
			assert.Failf(t, "unexpected key found in error context", "unexpected context key %q with value %#v", errorContextKey, errorContextValue)
		}
	}
}
