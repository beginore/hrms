package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func TestFieldCreation(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		wantKey  string
		wantKind string
	}{
		{
			name:     "String field",
			field:    String("key", "value"),
			wantKey:  "key",
			wantKind: "String",
		},
		{
			name:     "Int64 field",
			field:    Int64("count", 42),
			wantKey:  "count",
			wantKind: "Int64",
		},
		{
			name:     "Int field",
			field:    Int("age", 25),
			wantKey:  "age",
			wantKind: "Int64",
		},
		{
			name:     "Uint64 field",
			field:    Uint64("id", 123),
			wantKey:  "id",
			wantKind: "Uint64",
		},
		{
			name:     "Float64 field",
			field:    Float64("price", 19.99),
			wantKey:  "price",
			wantKind: "Float64",
		},
		{
			name:     "Bool field",
			field:    Bool("active", true),
			wantKey:  "active",
			wantKind: "Bool",
		},
		{
			name:     "Duration field",
			field:    Duration("latency", 100*time.Millisecond),
			wantKey:  "latency",
			wantKind: "Duration",
		},
		{
			name:     "Error field",
			field:    Error(errors.New("test error")),
			wantKey:  "error",
			wantKind: "String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := slog.Attr(tt.field)
			if attr.Key != tt.wantKey {
				t.Errorf("Expected key %q, got %q", tt.wantKey, attr.Key)
			}
			if attr.Value.Kind().String() != tt.wantKind {
				t.Errorf("Expected kind %q, got %q", tt.wantKind, attr.Value.Kind().String())
			}
		})
	}
}

func TestTimeField(t *testing.T) {
	now := time.Now()
	field := Time("timestamp", now)
	attr := slog.Attr(field)

	if attr.Key != "timestamp" {
		t.Errorf("Expected key 'timestamp', got %q", attr.Key)
	}

	gotTime := attr.Value.Time()
	if !now.Truncate(0).Equal(gotTime.Truncate(0)) {
		t.Errorf("Expected time %v, got %v", now, gotTime)
	}

	if now.Unix() != gotTime.Unix() {
		t.Errorf("Expected Unix timestamp %d, got %d", now.Unix(), gotTime.Unix())
	}
}

func TestAnyField(t *testing.T) {
	type CustomStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name  string
		value interface{}
	}{
		{"string", "test"},
		{"int", 42},
		{"struct", CustomStruct{Name: "John", Age: 30}},
		{"slice", []string{"a", "b", "c"}},
		{"map", map[string]int{"a": 1, "b": 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := Any("data", tt.value)
			attr := slog.Attr(field)

			if attr.Key != "data" {
				t.Errorf("Expected key 'data', got %q", attr.Key)
			}
		})
	}
}

func TestTraceFields(t *testing.T) {
	t.Run("Context without span", func(t *testing.T) {
		ctx := context.Background()
		fields := TraceFields(ctx)

		if fields != nil {
			t.Error("Expected nil for context without span")
		}
	})

	t.Run("Context with valid span", func(t *testing.T) {
		// Create a mock span context
		traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
		spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")

		spanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		})

		ctx := trace.ContextWithSpanContext(context.Background(), spanContext)
		fields := TraceFields(ctx)

		if len(fields) != 2 {
			t.Fatalf("Expected 2 fields, got %d", len(fields))
		}

		// Check trace_id field
		traceIDAttr := slog.Attr(fields[0])
		if traceIDAttr.Key != "trace_id" {
			t.Errorf("Expected key 'trace_id', got %q", traceIDAttr.Key)
		}
		if traceIDAttr.Value.String() != traceID.String() {
			t.Errorf("Expected trace_id %q, got %q", traceID.String(), traceIDAttr.Value.String())
		}

		// Check span_id field
		spanIDAttr := slog.Attr(fields[1])
		if spanIDAttr.Key != "span_id" {
			t.Errorf("Expected key 'span_id', got %q", spanIDAttr.Key)
		}
		if spanIDAttr.Value.String() != spanID.String() {
			t.Errorf("Expected span_id %q, got %q", spanID.String(), spanIDAttr.Value.String())
		}
	})
}

func TestNewLog(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantNil bool
	}{
		{
			name:    "Local environment",
			env:     "local",
			wantNil: false,
		},
		{
			name:    "Dev environment",
			env:     "dev",
			wantNil: false,
		},
		{
			name:    "Prod environment",
			env:     "prod",
			wantNil: false,
		},
		{
			name:    "Production environment",
			env:     "production",
			wantNil: false,
		},
		{
			name:    "Unknown environment defaults to prod",
			env:     "unknown",
			wantNil: false,
		},
		{
			name:    "Empty environment defaults to prod",
			env:     "",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLog(tt.env)

			if (logger == nil) != tt.wantNil {
				t.Errorf("NewLog() returned nil=%v, want nil=%v", logger == nil, tt.wantNil)
			}

			if logger != nil && logger.logger == nil {
				t.Error("Logger's internal slog.Logger should not be nil")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := &Log{
		logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	tests := []struct {
		name      string
		logFunc   func(string, ...Field)
		message   string
		fields    []Field
		wantLevel string
	}{
		{
			name:      "Debug level",
			logFunc:   logger.Debug,
			message:   "debug message",
			fields:    []Field{String("key", "value")},
			wantLevel: "DEBUG",
		},
		{
			name:      "Info level",
			logFunc:   logger.Info,
			message:   "info message",
			fields:    []Field{Int("count", 42)},
			wantLevel: "INFO",
		},
		{
			name:      "Warn level",
			logFunc:   logger.Warn,
			message:   "warn message",
			fields:    []Field{Bool("warning", true)},
			wantLevel: "WARN",
		},
		{
			name:      "Error level",
			logFunc:   logger.Error,
			message:   "error message",
			fields:    []Field{Error(errors.New("test error"))},
			wantLevel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.message, tt.fields...)

			output := buf.String()
			if !strings.Contains(output, tt.message) {
				t.Errorf("Expected log output to contain message %q, got %q", tt.message, output)
			}

			if !strings.Contains(output, tt.wantLevel) {
				t.Errorf("Expected log output to contain level %q, got %q", tt.wantLevel, output)
			}
		})
	}
}

func TestLogWith(t *testing.T) {
	var buf bytes.Buffer
	logger := &Log{
		logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	t.Run("With adds persistent fields", func(t *testing.T) {
		buf.Reset()
		childLogger := logger.With(String("request_id", "123"), String("user_id", "456"))
		childLogger.Info("test message")

		output := buf.String()
		if !strings.Contains(output, "request_id") || !strings.Contains(output, "123") {
			t.Error("Expected output to contain request_id field")
		}
		if !strings.Contains(output, "user_id") || !strings.Contains(output, "456") {
			t.Error("Expected output to contain user_id field")
		}
	})

	t.Run("With empty fields returns same logger", func(t *testing.T) {
		childLogger := logger.With()
		if childLogger != logger {
			t.Error("With() with no fields should return the same logger instance")
		}
	})

	t.Run("Chained With calls", func(t *testing.T) {
		buf.Reset()
		childLogger := logger.
			With(String("field1", "value1")).
			With(String("field2", "value2"))

		childLogger.Info("chained message")

		output := buf.String()
		if !strings.Contains(output, "field1") || !strings.Contains(output, "field2") {
			t.Error("Expected output to contain both chained fields")
		}
	})
}

func TestLogWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &Log{
		logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	t.Run("Context without trace", func(t *testing.T) {
		ctx := context.Background()
		childLogger := logger.WithContext(ctx)

		if childLogger != logger {
			t.Error("WithContext with no trace should return same logger")
		}
	})

	t.Run("Context with trace", func(t *testing.T) {
		buf.Reset()
		traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
		spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")

		spanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		})

		ctx := trace.ContextWithSpanContext(context.Background(), spanContext)
		childLogger := logger.WithContext(ctx)
		childLogger.Info("traced message")

		output := buf.String()
		if !strings.Contains(output, "trace_id") || !strings.Contains(output, traceID.String()) {
			t.Error("Expected output to contain trace_id")
		}
		if !strings.Contains(output, "span_id") || !strings.Contains(output, spanID.String()) {
			t.Error("Expected output to contain span_id")
		}
	})
}

func TestLogOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := &Log{
		logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	logger.Info("test message",
		String("string_field", "value"),
		Int("int_field", 42),
		Bool("bool_field", true),
	)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg 'test message', got %v", logEntry["msg"])
	}

	if logEntry["string_field"] != "value" {
		t.Errorf("Expected string_field 'value', got %v", logEntry["string_field"])
	}

	if logEntry["int_field"] != float64(42) {
		t.Errorf("Expected int_field 42, got %v", logEntry["int_field"])
	}

	if logEntry["bool_field"] != true {
		t.Errorf("Expected bool_field true, got %v", logEntry["bool_field"])
	}
}

func TestErr(t *testing.T) {
	testErr := errors.New("test error message")
	attr := Err(testErr)

	if attr.Key != "error" {
		t.Errorf("Expected key 'error', got %q", attr.Key)
	}

	if attr.Value.String() != "test error message" {
		t.Errorf("Expected value 'test error message', got %q", attr.Value.String())
	}
}

func TestLocalHandlerCreation(t *testing.T) {
	var buf bytes.Buffer
	opts := LocalHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewLocalHandler(&buf)

	if handler == nil {
		t.Fatal("NewLocalHandler should not return nil")
	}

	if handler.l == nil {
		t.Error("LocalHandler's logger should not be nil")
	}

	if handler.Handler == nil {
		t.Error("LocalHandler's Handler should not be nil")
	}
}

func TestLocalHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	opts := LocalHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewLocalHandler(&buf)
	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	newHandler := handler.WithAttrs(attrs)

	if newHandler == nil {
		t.Fatal("WithAttrs should not return nil")
	}

	localHandler, ok := newHandler.(*LocalHandler)
	if !ok {
		t.Fatal("WithAttrs should return *LocalHandler")
	}

	if len(localHandler.attrs) != len(attrs) {
		t.Errorf("Expected %d attrs, got %d", len(attrs), len(localHandler.attrs))
	}
}

func TestLocalHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	opts := LocalHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewLocalHandler(&buf)
	newHandler := handler.WithGroup("test_group")

	if newHandler == nil {
		t.Fatal("WithGroup should not return nil")
	}

	_, ok := newHandler.(*LocalHandler)
	if !ok {
		t.Fatal("WithGroup should return *LocalHandler")
	}
}
