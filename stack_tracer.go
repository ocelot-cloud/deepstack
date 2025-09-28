package deepstack

import (
	"fmt"
	"runtime"
	"strings"
)

//go:generate mockery
type StackTracer interface {
	GetStackTrace() string
}

type StackTracerImpl struct{}

func (s *StackTracerImpl) GetStackTrace() string {
	programCounters := make([]uintptr, 32)
	indexUntilProgramCountersShouldBeSkipped := runtime.Callers(3, programCounters)
	frames := runtime.CallersFrames(programCounters[:indexUntilProgramCountersShouldBeSkipped])
	var builder strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&builder, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return builder.String()
}

// TODO I want file paths relative to the project directory; neither absolute paths (log file, stack trace) not only file names (console output)
