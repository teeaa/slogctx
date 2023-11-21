package slogctx

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

type (
	contextKey string
	Handler    struct {
		handler slog.Handler
	}
)

var (
	fields contextKey   = "slog_fields"
	_      slog.Handler = Handler{}
)

func New(handler slog.Handler) *slog.Logger {
	if handler == nil {
		handler = slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{
				AddSource:   false,
				Level:       slog.LevelInfo,
				ReplaceAttr: replaceAttr,
			})
	}

	log := slog.New(Handler{handler: handler})
	slog.SetDefault(log)

	return log
}

func (h Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h Handler) Handle(ctx context.Context, record slog.Record) error {
	if v, ok := ctx.Value(fields).(*sync.Map); ok {
		v.Range(func(key, val any) bool {
			if keyString, ok := key.(string); ok {
				record.AddAttrs(slog.Any(keyString, val))
			}
			return true
		})
	}
	return h.handler.Handle(ctx, record)
}

func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return Handler{h.handler.WithAttrs(attrs)}
}

func (h Handler) WithGroup(name string) slog.Handler {
	return h.handler.WithGroup(name)
}
