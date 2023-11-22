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

To add new fields to context, do it with `slogctx.WithValue()`:

```Go
// ctx is already defined somewhere
slogctx.WithValue(ctx, "var name", var)
```

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

## Benefits

If you get context from, for example your service's REST endpoint being called, you can add values to context that might help with debugging. Then when you carry the context forwards, these values will be carried as well so identifying bugs and things will be easier.

Also our `otel` package adds a trace id to the context by default, so if you log context, you will much more easily find all the log entries from that call by filtering with that trace id.