package slogctx

import (
	"context"
	"log/slog"
	"time"

	"cloud.google.com/go/compute/metadata"
)

const (
	gcpTimeFormat            = "2006-01-02T15:04:05.99Z"
	LevelCritical slog.Level = 16
)

var (
	projectID    string
	instanceName string
)

func Critical(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, LevelCritical, msg, args...)
}

func (l *Logger) Critical(ctx context.Context, msg string, args ...any) {
	l.logger.Log(ctx, LevelCritical, msg, args...)
}

func getGCPHandler(opts *HandlerOptions) slog.Handler {
	projectID, _ = metadata.ProjectIDWithContext(context.Background())
	instanceName, _ = metadata.InstanceNameWithContext(context.Background())

	return slog.NewJSONHandler(output,
		&slog.HandlerOptions{
			AddSource:   opts.AddSource,
			Level:       opts.Level,
			ReplaceAttr: GCPReplaceAttr,
		}).
		WithAttrs([]slog.Attr{slog.String("project_id", projectID)}).
		WithAttrs([]slog.Attr{slog.String("instance_name", instanceName)})
}

func GCPReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		t, err := time.Parse(defaultTimeFormat, a.Value.String())
		if err != nil {
			return a
		}
		return slog.Attr{
			Key:   slog.TimeKey,
			Value: slog.StringValue(t.Format(gcpTimeFormat)),
		}
	// "level" => "severity"
	case slog.LevelKey:
		return slog.Attr{
			Key:   "severity",
			Value: a.Value,
		}
	// "msg" => "message"
	case slog.MessageKey:
		return slog.Attr{
			Key:   "message",
			Value: a.Value,
		}
	// "source" => "logging.googleapis.com/sourceLocation"
	case slog.SourceKey:
		source, ok := a.Value.Any().(*slog.Source)
		if !ok || source == nil {
			return a
		}
		return slog.Any("logging.googleapis.com/sourceLocation", source)
	}

	// Only add trace to errors
	switch a.Value.Kind() {
	case slog.KindAny:
		switch v := a.Value.Any().(type) {
		case error:
			a.Value = formatError(v)
		}
	}

	return a
}
