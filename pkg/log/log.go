package log

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Field represents a structured logging field used by the log facade.
type Field slog.Attr

// String creates a string logging field.
func String(key, value string) Field {
	return Field(slog.String(key, value))
}

// Int64 creates an int64 logging field.
func Int64(key string, value int64) Field {
	return Field(slog.Int64(key, value))
}

// Int creates an int logging field.
func Int(key string, value int) Field {
	return Field(slog.Int(key, value))
}

// Uint64 creates an uint64 logging field.
func Uint64(key string, value uint64) Field {
	return Field(slog.Uint64(key, value))
}

// Float64 creates a float64 logging field.
func Float64(key string, value float64) Field {
	return Field(slog.Float64(key, value))
}

// Bool creates a bool logging field.
func Bool(key string, value bool) Field {
	return Field(slog.Bool(key, value))
}

// Time creates a time logging field.
func Time(key string, value time.Time) Field {
	return Field(slog.Time(key, value))
}

// Duration creates a duration time field in nanoseconds.
func Duration(key string, value time.Duration) Field {
	return Field(slog.Duration(key, value))
}

// Any creates a logging field with an arbitrary value.
func Any(key string, value interface{}) Field {
	return Field(slog.Any(key, value))
}

// Error creates a logging field for an error value.
func Error(err error) Field {
	return Field(slog.String("error", err.Error()))
}

// TraceFields extracts trace and span identifiers from the context
// and returns them as logging fields.
func TraceFields(ctx context.Context) []Field {
	spanContext := trace.SpanContextFromContext(ctx)

	if !spanContext.IsValid() {
		return nil
	}

	return []Field{
		String("trace_id", spanContext.TraceID().String()),
		String("span_id", spanContext.SpanID().String()),
	}
}

type Log struct {
	logger *slog.Logger
}

func (l *Log) Debug(msg string, fields ...Field) {
	attrs := fieldsToAttrs(fields)
	l.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs...)
}

func (l *Log) Info(msg string, fields ...Field) {
	attrs := fieldsToAttrs(fields)
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

func (l *Log) Warn(msg string, fields ...Field) {
	attrs := fieldsToAttrs(fields)
	l.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs...)
}

func (l *Log) Error(msg string, fields ...Field) {
	attrs := fieldsToAttrs(fields)
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

func (l *Log) Fatal(msg string, fields ...Field) {
	attrs := fieldsToAttrs(fields)
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
	os.Exit(1)
}

func (l *Log) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return l
	}

	attrs := fieldsToAny(fields)
	return &Log{
		logger: l.logger.With(attrs...),
	}
}

func (l *Log) WithContext(ctx context.Context) Logger {
	traceFields := TraceFields(ctx)
	if len(traceFields) == 0 {
		return l
	}

	return l.With(traceFields...)
}

func fieldsToAny(fields []Field) []any {
	attrs := make([]any, len(fields))
	for i, f := range fields {
		attrs[i] = slog.Attr(f)
	}
	return attrs
}

func fieldsToAttrs(fields []Field) []slog.Attr {
	attrs := make([]slog.Attr, len(fields))
	for i, f := range fields {
		attrs[i] = slog.Attr(f)
	}
	return attrs
}

func NewLog(env string) *Log {
	var handler slog.Handler

	switch env {
	case "local":
		opts := LocalHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}
		handler = opts.NewLocalHandler(os.Stdout)
	case "dev":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	case "prod", "production":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return &Log{
		logger: slog.New(handler),
	}
}
