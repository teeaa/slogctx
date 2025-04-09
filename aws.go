package slogctx

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

const LevelFatal slog.Level = 16

var awsInstanceID string

func Fatal(ctx context.Context, msg string, args ...any) {
	slog.Default().Log(ctx, LevelFatal, msg, args...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, args ...any) {
	l.logger.Log(ctx, LevelFatal, msg, args...)
}

func getAWSEC2Handler(opts *HandlerOptions) slog.Handler {
	return slog.NewJSONHandler(output,
		&slog.HandlerOptions{
			AddSource:   opts.AddSource,
			Level:       opts.Level,
			ReplaceAttr: AWSReplaceAttr,
		}).
		WithAttrs([]slog.Attr{slog.String("ec2_instance", awsInstanceID)})
}

func getAWSLambdaHandler(opts *HandlerOptions, lc *lambdacontext.LambdaContext) slog.Handler {
	requestID := lc.AwsRequestID
	arn := lc.InvokedFunctionArn

	return slog.NewJSONHandler(output,
		&slog.HandlerOptions{
			AddSource:   opts.AddSource,
			Level:       opts.Level,
			ReplaceAttr: AWSReplaceAttr,
		}).
		WithAttrs([]slog.Attr{slog.String("function_arn", arn)}).
		WithAttrs([]slog.Attr{slog.String("request_id", requestID)})
}

func AWSReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		t, err := time.Parse(defaultTimeFormat, a.Value.String())
		if err != nil {
			return a
		}
		return slog.Attr{
			Key:   slog.TimeKey,
			Value: slog.StringValue(t.Format(time.RFC3339Nano)),
		}
	// "msg" => "message"
	case slog.MessageKey:
		return slog.Attr{
			Key:   "message",
			Value: a.Value,
		}
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

func isAWSEC2() bool {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	client := imds.NewFromConfig(cfg)

	// Set a short timeout for the check
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try to retrieve instance ID
	output, err := client.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "instance-id",
	})
	if err != nil {
		return false
	}
	defer output.Content.Close()
	instanceID, _ := io.ReadAll(output.Content)
	awsInstanceID = string(instanceID)
	return true
}
