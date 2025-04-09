package slogctx

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var defaultTimeFormat = "2006-01-02 15:04:05.999999 -0700 MST"

func ReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		if timeFormat != "" {
			t, err := time.Parse(defaultTimeFormat, a.Value.String())
			if err != nil {
				return a
			}
			return slog.Attr{
				Key:   slog.TimeKey,
				Value: slog.StringValue(t.Format(timeFormat)),
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

	groupValues = append(groupValues, slog.String(messageKey, err.Error()))

	type StackTracer interface {
		StackTrace() errors.StackTrace
	}

	// Find the trace to the location of the first errors.New,
	// errors.Wrap, or errors.WithStack call.
	var st StackTracer
	for err := err; err != nil; err = errors.Unwrap(err) {
		x, ok := err.(StackTracer)
		if ok {
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
