package utils

import (
	"errors"
	"testing"
)

func TestLoggingVisually(t *testing.T) {
	logger := NewDeepStackLogger("debug", true, false)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	logger.Info("This is an info message", "key1", "value1", "key2", "value2")
	logger.Error("This is an info message", ErrorField, "some-error")
	logger.Error("This is an info message", ErrorField, errors.New("some-error"))
}

func TestLoggingVisuallyOfNormalError(t *testing.T) {
	logger := NewDeepStackLogger("debug", true, true)
	logger.Error("testing normal error", ErrorField, errors.New("some-error"), "key1", "value1")
}

func TestLoggingWithStackTrace(t *testing.T) {
	logger := NewDeepStackLogger("debug", true, true)
	logger.Error("testing detailed error", ErrorField, subfunction(logger))
}

func subfunction(logger DeepStackLogger) error {
	return logger.NewError("an error occurred", "key1", "value1")
}
