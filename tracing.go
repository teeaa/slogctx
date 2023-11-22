package slogctx

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var timeFormat = time.RFC3339

func ReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	// fmt.Printf("key %v, value %v\n", a.Key, a.Value)
	if a.Key == "time" {
		if timeFormat == "" {
			return slog.Attr{}
		} else if timeFormat != time.RFC3339 {
			t, err := time.Parse("2006-01-02 15:04:05.99999 -0700 MST", a.Value.String())
			if err != nil {
				return a
			}
			return slog.Attr{
				Key:   "time",
				Value: slog.StringValue(t.Format(time.DateTime)),
			}
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

func formatError(err error) slog.Value {
	var groupValues []slog.Attr

	groupValues = append(groupValues, slog.String("msg", err.Error()))

	type StackTracer interface {
		StackTrace() errors.StackTrace
	}

	// Find the trace to the location of the first errors.New,
	// errors.Wrap, or errors.WithStack call.
	var st StackTracer
	for err := err; err != nil; err = errors.Unwrap(err) {
		if x, ok := err.(StackTracer); ok {
			st = x
		}
	}

	if st != nil {
		groupValues = append(groupValues,
			slog.Any("trace", traceLines(st.StackTrace())),
		)
	}

	return slog.GroupValue(groupValues...)
}

func traceLines(frames errors.StackTrace) []string {
	var skipped int
	traceLines := make([]string, len(frames))
	skipping := true

	for i := len(frames) - 1; i >= 0; i-- {
		pc := uintptr(frames[i]) - 1
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			traceLines[i] = "unknown"
			skipping = false
			continue
		}

		name := fn.Name()

		if skipping && strings.HasPrefix(name, "runtime.") {
			skipped++
			continue
		} else {
			skipping = false
		}

		filename, lineNr := fn.FileLine(pc)

		traceLines[i] = fmt.Sprintf("%s %s:%d", name, filename, lineNr)
	}

	return traceLines[:len(traceLines)-skipped]
}
