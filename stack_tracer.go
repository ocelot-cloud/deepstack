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
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	var b strings.Builder
	for {
		f, more := frames.Next()
		fmt.Fprintf(&b, "%s\n\t%s:%d\n", f.Function, f.File, f.Line)
		if !more {
			break
		}
	}
	return b.String()
}

// TODO I want file paths relative to the project directory; neither absolute paths (log file, stack trace) not only file names (console output)
