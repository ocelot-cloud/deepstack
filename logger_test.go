package deepstack

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

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
	l, m := newLogger(t, false)
	err := &DeepStackError{Message: "m", StackTrace: "trace", Context: map[string]any{"k": "v"}}
	m.EXPECT().ShouldLogBeSkipped("error").Return(false)
	m.EXPECT().CreateLogRecord("error", "msg").Return(GetSampleLogRecord())
	m.EXPECT().HandleRecord(mock.Anything)
	m.EXPECT().Println("trace")
	l.log("error", "msg", ErrorField, err)
	m.AssertExpectations(t)
}

func TestLogNormalErrorNoWarning(t *testing.T) {
	l, m := newLogger(t, false)
	m.EXPECT().ShouldLogBeSkipped("error").Return(false)
	m.EXPECT().CreateLogRecord("error", "msg").Return(GetSampleLogRecord())
	m.EXPECT().HandleRecord(mock.Anything)
	l.log("error", "msg", ErrorField, errors.New("e"))
	m.AssertExpectations(t)
}

func TestLogNormalErrorWithWarning(t *testing.T) {
	l, m := newLogger(t, true)
	m.EXPECT().ShouldLogBeSkipped("error").Return(false)
	m.EXPECT().CreateLogRecord("error", "msg").Return(GetSampleLogRecord())
	m.EXPECT().LogWarning("invalid error type in log message, must be *DeepStackError")
	m.EXPECT().HandleRecord(mock.Anything)
	l.log("error", "msg", ErrorField, errors.New("e"))
	m.AssertExpectations(t)
}

func TestLogInvalidKeyType(t *testing.T) {
	l, m := newLogger(t, false)
	rec := GetSampleLogRecord()

	m.EXPECT().ShouldLogBeSkipped("info").Return(false)
	m.EXPECT().CreateLogRecord("info", "msg").Return(rec)
	m.EXPECT().LogWarning(
		"invalid key type in log message, must always be string",
		[]interface{}{"type", reflect.TypeOf(0).String()},
	)
	m.EXPECT().HandleRecord(rec)

	l.log("info", "msg", 123, "value1", "key2", "value2")
	m.AssertExpectations(t)
}

func TestAddContextNormalError(t *testing.T) {
	l, m := newLogger(t, false)
	inputError := errors.New("some error")
	m.EXPECT().LogWarning("invalid error type in log message, must be *DeepStackError")
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

func TestAddContextDeepStackError(t *testing.T) {
	l, m := newLogger(t, false)
	inputError := &DeepStackError{
		Message:    "some error",
		StackTrace: "some stack trace",
		Context:    map[string]any{"key1": "value1"},
	}
	outputError := l.AddContext(inputError, "key2", "value2")

	err, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some error", err.Message)
	assert.Equal(t, 2, len(err.Context))
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
	assert.Equal(t, "some stack trace", err.StackTrace)

	m.AssertExpectations(t)
}
