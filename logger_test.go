package deepstack

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TODO there is not assertions yet, that the attributes "stack_trace" and "error_cause" are set in the log record

func TestLoggingVisuallyOfNormalError(t *testing.T) {
	logger := NewDeepStackLogger("debug", true)
	logger.Error("testing normal error", ErrorField, errors.New("some-error"), "key1", "value1")
}

func newLogger(t *testing.T, warn bool) (*DeepStackLoggerImpl, *LoggingBackendMock, *StackTracerMock) {
	loggingBackendMock := NewLoggingBackendMock(t)
	stackTracerMock := NewStackTracerMock(t)
	return &DeepStackLoggerImpl{
		logger:               loggingBackendMock,
		enableMisuseWarnings: warn,
		stackTracer:          stackTracerMock,
	}, loggingBackendMock, stackTracerMock
}

func TestLogSkip(t *testing.T) {
	logger, backendMock, _ := newLogger(t, false)
	backendMock.EXPECT().ShouldLogBeSkipped("debug").Return(true)
	logger.log("debug", "msg")
	backendMock.AssertExpectations(t)
}

func TestLogDeepStackError(t *testing.T) {
	logger, backendMock, _ := newLogger(t, false)
	err := &DeepStackError{Message: "some-error-cause", StackTrace: "trace", Context: map[string]any{"key1": "value1"}}
	backendMock.EXPECT().ShouldLogBeSkipped("error").Return(false)

	expectedLogRecord := &Record{
		level:      "error",
		msg:        "msg",
		attributes: map[string]any{"key1": "value1", "stack_trace": "trace", "error_cause": "some-error-cause"},
	}
	backendMock.EXPECT().LogRecord(expectedLogRecord)

	backendMock.EXPECT().Println("trace")
	logger.log("error", "msg", ErrorField, err)
	backendMock.AssertExpectations(t)
}

func TestLogNormalErrorNoWarning(t *testing.T) {
	l, m, _ := newLogger(t, false)
	m.EXPECT().ShouldLogBeSkipped("error").Return(false)
	m.EXPECT().LogRecord(mock.Anything) // TODO replace all occurrences of mock.Anything with something more specific
	l.log("error", "msg", ErrorField, errors.New("e"))
	m.AssertExpectations(t)
}

func TestLogNormalErrorWithWarning(t *testing.T) {
	l, m, _ := newLogger(t, true)
	m.EXPECT().ShouldLogBeSkipped("error").Return(false)
	m.EXPECT().LogWarning("invalid error type in log message, must be *DeepStackError")
	m.EXPECT().LogRecord(mock.Anything)
	l.log("error", "msg", ErrorField, errors.New("e"))
	m.AssertExpectations(t)
}

func TestLogInvalidKeyType(t *testing.T) {
	l, m, _ := newLogger(t, false)
	expectedLogRecord := &Record{
		level:      "info",
		msg:        "msg",
		attributes: map[string]any{"key2": "value2"},
	}

	m.EXPECT().ShouldLogBeSkipped("info").Return(false)
	m.EXPECT().LogWarning(
		"invalid key type in log message, must always be string",
		[]interface{}{"type", reflect.TypeOf(0).String()},
	)
	m.EXPECT().LogRecord(expectedLogRecord)
	l.log("info", "msg", 123, "value1", "key2", "value2")
	m.AssertExpectations(t)
}

func TestAddContextNormalError(t *testing.T) {
	logger, backendMock, stackTracerMock := newLogger(t, true)
	inputError := errors.New("some error")
	backendMock.EXPECT().LogWarning("invalid error type in log message, must be *DeepStackError")
	stackTracerMock.EXPECT().GetStackTrace().Return("some-stack-trace")
	createAndAssertDeepstackError(t, logger, inputError)
	backendMock.AssertExpectations(t)
	stackTracerMock.AssertExpectations(t)
}

func createAndAssertDeepstackError(t *testing.T, l *DeepStackLoggerImpl, inputError error) {
	outputError := l.AddContext(inputError, "key1", "value1", "key2", "value2")

	err, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some error", err.Message)
	assert.Equal(t, 2, len(err.Context))
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
	assert.Equal(t, "some-stack-trace", err.StackTrace)
}

func TestAddContextNormalError_DisabledWarnings(t *testing.T) {
	logger, backendMock, stackTracerMock := newLogger(t, false)
	inputError := errors.New("some error")
	stackTracerMock.EXPECT().GetStackTrace().Return("some-stack-trace")
	createAndAssertDeepstackError(t, logger, inputError)
	stackTracerMock.AssertExpectations(t)
	backendMock.AssertExpectations(t)
}

func TestAddContextDeepStackError(t *testing.T) {
	logger, backendMock, _ := newLogger(t, false)
	inputError := &DeepStackError{
		Message:    "some error",
		StackTrace: "some-stack-trace",
		Context:    map[string]any{"key1": "value1"},
	}
	outputError := logger.AddContext(inputError, "key2", "value2")

	err, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some error", err.Message)
	assert.Equal(t, 2, len(err.Context))
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
	assert.Equal(t, "some-stack-trace", err.StackTrace)

	backendMock.AssertExpectations(t)
}
