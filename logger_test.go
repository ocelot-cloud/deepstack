package deepstack

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TODO there is not assertions yet, that the attributes "stack_trace" and "error_cause" are set in the log record

// TODO I dont like that the output of this test is to be checked by humans. should be automated. Maybe add an option to return the output as a string?; check that afterwards not more console output is written
func TestLoggingVisually(t *testing.T) {
	logger := NewDeepStackLogger("debug", false)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	logger.Info("This is an info message", "key1", "value1", "key2", "value 2")
	logger.Error("This is an info message", ErrorField, "some-error")
	logger.Error("This is an info message", ErrorField, errors.New("some-error"))
}

// TODO during testing I also do not want to generate log files, as this is slow. architecture should be like this: only in memory unit tests; integration tests for writing to log file, maybe also for console output?; in production return the logger with real implementations
// TODO problem: we need to also insert a date producer
func TestConsoleOutput(t *testing.T) {
	consoleSpy := &bytes.Buffer{}
	logger := newDeepStackLoggerForTesting("debug", false, consoleSpy)
	logger.Info("msg", "k", "v")
	// TODO assert.Equal(t, "todo", consoleSpy.String())
}

func TestLoggingVisuallyOfNormalError(t *testing.T) {
	logger := NewDeepStackLogger("debug", true)
	logger.Error("testing normal error", ErrorField, errors.New("some-error"), "key1", "value1")
}

func TestLoggingWithStackTrace(t *testing.T) {
	logger := NewDeepStackLogger("debug", true)
	logger.Error("testing detailed error", ErrorField, subfunction(logger))
}

func subfunction(logger DeepStackLogger) error {
	return logger.NewError("an error occurred", "key1", "value1")
}

func GetSampleLogRecord() *LogRecord {
	return &LogRecord{
		level:      "debug",
		msg:        "some message",
		attributes: map[string]any{"key1": "value1", "key2": "value2"},
	}
}

func newLogger(tb testing.TB, warn bool) (*DeepStackLoggerImpl, *LoggingBackendMock) {
	tb.Helper()
	m := NewLoggingBackendMock(tb)
	return &DeepStackLoggerImpl{logger: m, enableWarningsForNonDeepStackErrors: warn}, m
}

func TestLogSkip(t *testing.T) {
	l, m := newLogger(t, false)
	m.EXPECT().ShouldLogBeSkipped("debug").Return(true)
	l.log("debug", "msg")
	m.AssertExpectations(t)
}

func TestLogDeepStackError(t *testing.T) {
	logger, backendMock := newLogger(t, false)
	err := &DeepStackError{Message: "some-error-cause", StackTrace: "trace", Context: map[string]any{"key1": "value1"}}
	backendMock.EXPECT().ShouldLogBeSkipped("error").Return(false)

	expectedLogRecord := &LogRecord{
		level:      "error",
		msg:        "msg",
		attributes: map[string]any{"key1": "value1", "stack_trace": "trace", "error_cause": "some-error-cause"},
	}
	backendMock.EXPECT().HandleRecord(expectedLogRecord)

	backendMock.EXPECT().Println("trace")
	logger.log("error", "msg", ErrorField, err)
	backendMock.AssertExpectations(t)
}

func TestLogNormalErrorNoWarning(t *testing.T) {
	l, m := newLogger(t, false)
	m.EXPECT().ShouldLogBeSkipped("error").Return(false)
	m.EXPECT().HandleRecord(mock.Anything)
	l.log("error", "msg", ErrorField, errors.New("e"))
	m.AssertExpectations(t)
}

func TestLogNormalErrorWithWarning(t *testing.T) {
	l, m := newLogger(t, true)
	m.EXPECT().ShouldLogBeSkipped("error").Return(false)
	m.EXPECT().LogWarning("invalid error type in log message, must be *DeepStackError")
	m.EXPECT().HandleRecord(mock.Anything)
	l.log("error", "msg", ErrorField, errors.New("e"))
	m.AssertExpectations(t)
}

func TestLogInvalidKeyType(t *testing.T) {
	l, m := newLogger(t, false)
	expectedLogRecord := &LogRecord{
		level:      "info",
		msg:        "msg",
		attributes: map[string]any{"key2": "value2"},
	}

	m.EXPECT().ShouldLogBeSkipped("info").Return(false)
	m.EXPECT().LogWarning(
		"invalid key type in log message, must always be string",
		[]interface{}{"type", reflect.TypeOf(0).String()},
	)
	m.EXPECT().HandleRecord(expectedLogRecord)
	l.log("info", "msg", 123, "value1", "key2", "value2")
	m.AssertExpectations(t)
}

func TestAddContextNormalError(t *testing.T) {
	logger, backendMock := newLogger(t, true)
	inputError := errors.New("some error")
	backendMock.EXPECT().LogWarning("invalid error type in log message, must be *DeepStackError")
	todoMyChangeName(t, logger, inputError, backendMock)
}

func todoMyChangeName(t *testing.T, l *DeepStackLoggerImpl, inputError error, m *LoggingBackendMock) {
	outputError := l.AddContext(inputError, "key1", "value1", "key2", "value2")

	err, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some error", err.Message)
	assert.Equal(t, 2, len(err.Context))
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
	assert.NotEqual(t, "", err.StackTrace)

	m.AssertExpectations(t)
}

func TestAddContextNormalError_DisabledWarnings(t *testing.T) {
	logger, BackendMock := newLogger(t, false)
	inputError := errors.New("some error")
	todoMyChangeName(t, logger, inputError, BackendMock)
}

func TestAddContextDeepStackError(t *testing.T) {
	logger, backendMock := newLogger(t, false)
	inputError := &DeepStackError{
		Message:    "some error",
		StackTrace: "some stack trace",
		Context:    map[string]any{"key1": "value1"},
	}
	outputError := logger.AddContext(inputError, "key2", "value2")

	err, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some error", err.Message)
	assert.Equal(t, 2, len(err.Context))
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
	assert.Equal(t, "some stack trace", err.StackTrace)

	backendMock.AssertExpectations(t)
}
