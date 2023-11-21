package slogctx

import (
	"context"
	"log/slog"
	"sync"
)

func Debug(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}

func WithValue(parent context.Context, key string, val any) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(fields).(*sync.Map); ok {
		v.Store(key, val)

		return context.WithValue(parent, fields, v)
	}

	v := &sync.Map{}
	v.Store(key, val)

	return context.WithValue(parent, fields, v)
}
