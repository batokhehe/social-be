package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	entryContextKey contextKey = "logger_entry"
	requestIDField             = "request_id"
)

var Log *logrus.Logger
var Slog *slog.Logger

func Init() {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)

	Slog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	entry := FromContext(ctx).WithField(requestIDField, requestID)
	return context.WithValue(ctx, entryContextKey, entry)
}

func FromContext(ctx context.Context) *logrus.Entry {
	if Log == nil {
		Init()
	}

	if entry, ok := ctx.Value(entryContextKey).(*logrus.Entry); ok {
		return withTrace(ctx, entry)
	}

	return withTrace(ctx, logrus.NewEntry(Log))
}

func SlogFromContext(ctx context.Context) *slog.Logger {
	if Slog == nil {
		Init()
	}

	fields := []any{}
	if requestID, ok := requestIDFromContext(ctx); ok {
		fields = append(fields, requestIDField, requestID)
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		fields = append(fields,
			"trace_id", spanContext.TraceID().String(),
			"span_id", spanContext.SpanID().String(),
		)
	}

	if len(fields) == 0 {
		return Slog
	}

	return Slog.With(fields...)
}

func withTrace(ctx context.Context, entry *logrus.Entry) *logrus.Entry {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return entry
	}

	return entry.WithFields(logrus.Fields{
		"trace_id": spanContext.TraceID().String(),
		"span_id":  spanContext.SpanID().String(),
	})
}

func requestIDFromContext(ctx context.Context) (string, bool) {
	if entry, ok := ctx.Value(entryContextKey).(*logrus.Entry); ok {
		if value, ok := entry.Data[requestIDField].(string); ok {
			return value, true
		}
	}

	return "", false
}
