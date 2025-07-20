package utils

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

type LoggingBackend interface {
	ShouldLogBeSkipped(level string) bool
	CreateLogRecord(level string, msg string) *LogRecord
	HandleRecord(logRecord *LogRecord)
	Println(message string)
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
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
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
