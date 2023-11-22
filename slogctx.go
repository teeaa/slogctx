package slogctx

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
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

type HandlerOptions struct {
	AddSource   bool
	Level       slog.Level
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	TimeFormat  string
}

func New(opts *HandlerOptions) {
	if opts == nil {
		opts = &HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelInfo,
			ReplaceAttr: ReplaceAttr,
			TimeFormat:  time.RFC3339,
		}
	}

	handler := slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{
			AddSource:   opts.AddSource,
			Level:       opts.Level,
			ReplaceAttr: opts.ReplaceAttr,
		})

	timeFormat = opts.TimeFormat

	slog.SetDefault(slog.New(Handler{handler: handler}))
}

func NewWithHandler(handler slog.Handler) {
	slog.SetDefault(slog.New(Handler{handler: handler}))
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
