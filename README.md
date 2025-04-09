# slogctx - context based logging

## Usage

Initalise the logger by adding this in your `main.go`
```Go
slogctx.New(nil)
```

After this you can log with context with:

```Go
slogctx.Info(ctx, "log message",
    slog.String("var name", var),
    )
```

You can also use other methods, such as `slog.Int()`, `slog.Bool()`, `slog.Any()` and others. Avoid using `slog.Any()` when possible, as it uses reflect and that is an expensive operation.

_Note: If you don't have context available you can replace it with `nil`_

### Using context

To add an entry to be in every log line afterwards, add it with `slogctx.WithValue()`:

```Go
// ctx is already defined somewhere
slogctx.WithValue(ctx, "ctx key", "ctx value")
```

This will cause `slogctx.Info("foo")` to print out `"msg"="foo" "ctx key"="ctx value"`

## Modifying log level and time format

If you want to change the log level from the default (debug) you can pass new HandlerOptions to slogctx.New():

```Go
slogctx.New(&slogctx.HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelDebug,
			ReplaceAttr: slogctx.ReplaceAttr,
			TimeFormat:  time.DateTime,
		})
```

Or if you want a new handler altogether, you can use slogctx.NewWithHandler():

```Go
slogctx.NewWithHandler(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{
			AddSource:   opts.AddSource,
			Level:       opts.Level,
			ReplaceAttr: opts.ReplaceAttr,
		}))
```

_Note: You can't define time format with the default slog handler_

## AWS and GCP specific logging

To enable AWS and GCP specific logging (for field names and value formatting) use
```Go
slogctx.New(&slogctx.HandlerOptions{
			EnableAWS: true,
			EnableGCP: true,
		})
```

This will auto-detect if the code is running in AWS or GCP environment and format accordingly. If it isn't running in either environment it will fall back to default behaviour (with a short delay)