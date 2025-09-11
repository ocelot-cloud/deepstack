package deepstack

import "log/slog"

const (
	ErrorField = "error"
)

type Record struct {
	level      slog.Level
	msg        string
	attributes map[string]any
}

func (r *Record) AddAttrs(key string, value any) {
	r.attributes[key] = value
}
