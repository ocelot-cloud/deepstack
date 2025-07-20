package deepstack

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestErrorToString(t *testing.T) {
	logger := NewDeepStackLogger("debug", true, false)
	testError := logger.NewError("an error occurred", "key1", "value1")

	detailedTestError, ok := testError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "an error occurred", detailedTestError.Message)
	assert.Equal(t, 1, len(detailedTestError.Context))
	assert.Equal(t, "value1", detailedTestError.Context["key1"])
	assert.NotEqual(t, "", detailedTestError.StackTrace)

	errorString := testError.Error()
	assert.True(t, strings.HasPrefix(errorString, "an error occurred key1=value1\nstack trace:\n"))
}
