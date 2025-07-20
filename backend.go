package deepstack

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

//go:generate mockery
type LoggingBackend interface {
	ShouldLogBeSkipped(level string) bool
	CreateLogRecord(level string, msg string) *LogRecord
	HandleRecord(logRecord *LogRecord)
	Println(message string)
	LogWarning(message string, kv ...any)
}

type LoggingBackendImpl struct {
	slog *slog.Logger
}

func (s *LoggingBackendImpl) Println(message string) {
	println(message)
}

func (s *LoggingBackendImpl) ShouldLogBeSkipped(level string) bool {
	slogLevel := convertToSlogLevel(level)
	return !s.slog.Handler().Enabled(context.Background(), slogLevel)
}

func (s *LoggingBackendImpl) HandleRecord(logRecord *LogRecord) {
	s.logRecord(logRecord, 4)
}

func (s *LoggingBackendImpl) logRecord(logRecord *LogRecord, skipFunctionTreeLevels int) {
	var pcs [1]uintptr
	runtime.Callers(skipFunctionTreeLevels, pcs[:])
	slogLevel := convertToSlogLevel(logRecord.level)
	slogRecord := slog.NewRecord(time.Now(), slogLevel, logRecord.msg, pcs[0])

	for key, value := range logRecord.attributes {
		slogRecord.AddAttrs(slog.Any(key, value))
	}

	_ = s.slog.Handler().Handle(context.Background(), slogRecord)
}

func (s *LoggingBackendImpl) CreateLogRecord(level string, msg string) *LogRecord {
	return &LogRecord{
		level:      level,
		msg:        msg,
		attributes: make(map[string]any),
	}
}

func (s *LoggingBackendImpl) LogWarning(message string, kv ...any) {
	if len(kv) == 0 {
		record := s.CreateLogRecord("warn", message)
		s.logRecord(record, 6)
	} else if len(kv) == 2 {
		key, ok := kv[0].(string)
		if !ok {
			s.slog.Warn("invalid key type in log message, must always be string", slog.Any("key", kv[0]))
			return
		}
		s.slog.Warn(message, slog.Any(key, kv[1]))
	}
}
