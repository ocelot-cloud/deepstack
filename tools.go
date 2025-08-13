package deepstack

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	ErrorField = "error"
)

var (
	dataDir       = "data"
	workDirectory string

	logLevelMap = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
)

func init() {
	var err error
	workDirectory, err = os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("cannot determine working dir: %v", err))
	}
	if _, err = os.Stat(dataDir); os.IsNotExist(err) {
		if err = os.MkdirAll(dataDir, 0700); err != nil {
			panic(fmt.Sprintf("Error creating data directory: %v", err))
		}
	}
}

func convertToSlogLevel(logLevel string) slog.Level {
	lvl, ok := logLevelMap[strings.ToLower(logLevel)]
	if ok {
		return lvl
	} else {
		return slog.LevelInfo
	}
}

type Record struct {
	level      string
	msg        string
	attributes map[string]any
}

func (r *Record) AddAttrs(key string, value any) {
	r.attributes[key] = value
}
