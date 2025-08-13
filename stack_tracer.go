package deepstack

type StackTracer interface {
	GetStackTrace() string
}

type StackTracerImpl struct{}

func (s *StackTracerImpl) GetStackTrace() string {
	return printStackTrace()
}
