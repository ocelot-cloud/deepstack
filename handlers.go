package deepstack

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
)

type multiHandler []slog.Handler

func (h multiHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	for _, hd := range h {
		if hd.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}

func (h multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, hd := range h {
		_ = hd.Handle(ctx, r)
	}
	return nil
}

func (h multiHandler) WithAttrs(a []slog.Attr) slog.Handler {
	out := make(multiHandler, len(h))
	for i, hd := range h {
		out[i] = hd.WithAttrs(a)
	}
	return out
}

func (h multiHandler) WithGroup(name string) slog.Handler {
	out := make(multiHandler, len(h))
	for i, hd := range h {
		out[i] = hd.WithGroup(name)
	}
	return out
}

// TODO add tests to console handler

type consoleHandler struct {
	w     io.Writer
	opts  *slog.HandlerOptions
	attrs []slog.Attr
}

func (s consoleHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	if s.opts != nil && s.opts.Level != nil {
		return lvl >= s.opts.Level.Level()
	}
	return true
}

func (s consoleHandler) Handle(_ context.Context, r slog.Record) error {
	frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
	fileLine := fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
	var recAttrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == slog.SourceKey || a.Key == "stack_trace" {
			return true
		}
		recAttrs = append(recAttrs, a)
		return true
	})
	c, reset := lvlColor[r.Level], "\x1b[0m"
	fmt.Fprintf(s.w, "%s%s %s %s %q", c, r.Time.Format("2006-01-02 15:04:05.000"), r.Level, fileLine, r.Message)
	for _, a := range append(s.attrs, recAttrs...) {
		fmt.Fprintf(s.w, " %s=%v", a.Key, a.Value)
	}
	fmt.Fprintln(s.w, reset)
	return nil
}

func (s consoleHandler) WithAttrs(a []slog.Attr) slog.Handler {
	n := s
	n.attrs = append(append([]slog.Attr{}, s.attrs...), a...)
	return n
}

func (s consoleHandler) WithGroup(string) slog.Handler { return s }
