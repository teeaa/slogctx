package slogctx

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type (
	contextKey string
	LogFormat  string
	Handler    struct {
		handler slog.Handler
	}
)

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

var (
	fields     contextKey   = "slog_fields"
	_          slog.Handler = Handler{}
	output     io.Writer    = os.Stdout
	timeFormat              = time.RFC3339
	messageKey              = "msg"
)

type HandlerOptions struct {
	AddSource   bool                                         // add calling file and line to logs
	Level       slog.Level                                   // log level, default is Info
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr // function to replace attributes
	TimeFormat  string                                       // time format, default is RFC3339
	Format      LogFormat                                    // log format, default is Text for readability
	EnableAWS   bool                                         // enable AWS specific logging
	EnableGCP   bool                                         // enable GCP specific logging
}

func New(opts *HandlerOptions) *Logger {
	if opts == nil {
		opts = &HandlerOptions{
			Level:       slog.LevelInfo,
			ReplaceAttr: ReplaceAttr,
			TimeFormat:  time.RFC3339,
			Format:      TextFormat,
		}
	}

	// Reformat when running in GCP environment
	if opts.EnableGCP && metadata.OnGCE() {
		logger := slog.New(Handler{handler: getGCPHandler(opts)})
		slog.SetDefault(logger)
		return &Logger{logger: logger}
	}

	// Reformat when running in AWS Lambda environment
	if opts.EnableAWS {
		lc, found := lambdacontext.FromContext(context.Background())
		if found {
			logger := slog.New(Handler{handler: getAWSLambdaHandler(opts, lc)})
			slog.SetDefault(logger)
			return &Logger{logger: logger}
		}

		if isAWSEC2() {
			logger := slog.New(Handler{handler: getAWSEC2Handler(opts)})
			slog.SetDefault(logger)
			return &Logger{logger: logger}
		}
	}

	slogOpts := &slog.HandlerOptions{
		AddSource:   opts.AddSource,
		Level:       opts.Level,
		ReplaceAttr: opts.ReplaceAttr,
	}

	var handler slog.Handler = slog.NewJSONHandler(output, slogOpts)
	if opts.Format == TextFormat {
		handler = slog.NewTextHandler(output, slogOpts)
	}

	timeFormat = opts.TimeFormat

	logger := slog.New(Handler{handler: handler})
	slog.SetDefault(logger)
	return &Logger{logger: logger}
}

func SetOutput(out io.Writer) {
	output = out
}

func NewWithHandler(handler slog.Handler) *Logger {
	logger := slog.New(Handler{handler: handler})
	slog.SetDefault(logger)
	return &Logger{logger: logger}
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

// WithValue adds a key-value pair to the *logging* context
// NOTE: It will not add it to normal context
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
