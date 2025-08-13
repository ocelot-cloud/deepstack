package deepstack

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

var workDir = getWorkDir()

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

// TODO add tests to console handler, consider implementing a dependency for mocking

type ConsoleHandler struct {
	w     io.Writer
	opts  *slog.HandlerOptions
	attrs []slog.Attr
}

func (s ConsoleHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	if s.opts != nil && s.opts.Level != nil {
		return lvl >= s.opts.Level.Level()
	}
	return true
}

func (s ConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	fileLine := getFileLineRelativeToWorkDir(r)
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

func getWorkDir() string {
	currentDir, _ := os.Getwd()
	return currentDir
}

func getFileLineRelativeToWorkDir(r slog.Record) string {
	frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
	p, err := filepath.Rel(workDir, frame.File)
	if err != nil {
		p = filepath.Base(frame.File)
	}
	return fmt.Sprintf("%s:%d", p, frame.Line)
}

func (s ConsoleHandler) WithAttrs(a []slog.Attr) slog.Handler {
	n := s
	n.attrs = append(append([]slog.Attr{}, s.attrs...), a...)
	return n
}

func (s ConsoleHandler) WithGroup(string) slog.Handler { return s }
