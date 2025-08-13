package deepstack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorToString(t *testing.T) {
	logger, _, stackTracerMock := newLogger(t)
	stackTracerMock.EXPECT().GetStackTrace().Return("some-stack-trace")
	testError := logger.NewError("an error occurred", "key1", "value1")

	detailedTestError, ok := testError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "an error occurred", detailedTestError.Message)
	assert.Equal(t, 1, len(detailedTestError.Context))
	assert.Equal(t, "value1", detailedTestError.Context["key1"])
	assert.Equal(t, "some-stack-trace", detailedTestError.StackTrace)

	assert.Equal(t, "an error occurred", testError.Error())
	stackTracerMock.AssertExpectations(t)
}
