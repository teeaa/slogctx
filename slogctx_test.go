package slogctx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"
	"time"

	"github.com/pkg/errors"
)

func TestContext(t *testing.T) {
	ctx := context.Background()
	ts := time.Now().Format(time.DateOnly)
	var buf bytes.Buffer
	SetOutput(&buf)

	log := New(&HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelInfo,
		ReplaceAttr: ReplaceAttr,
		TimeFormat:  time.DateOnly,
		Format:      JSONFormat,
	})

	ctx = WithValue(ctx, "context key", "context value")
	log.Info(ctx, "debug message", slog.String("string key", "string value"))

	output := string(buf.Bytes())
	expected := `{"time":"` + ts + `","level":"INFO","msg":"debug message","string key":"string value","context key":"context value"}` + "\n"
	if output != expected {
		t.Errorf("expected %s, got %s", expected, output)
	}
}

func TestErrorStacktrace(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	SetOutput(&buf)

	log := New(nil)

	log.Error(ctx, "debug message", slog.Any("error", errors.New("error")))

	output := string(buf.Bytes())
	if !strings.Contains(output, "debug message") {
		t.Errorf("expected log to contain \"debug message\" in output: %s", output)
	}
	if !strings.Contains(output, "slogctx.TestErrorStacktrace") {
		t.Errorf("expected log to contain calling function in output: %s", output)
	}
	if !strings.Contains(output, "slogctx_test.go:48") {
		t.Errorf("expected log to contain calling file and line number in output: %s", output)
	}
}

func TestSlogCtx(t *testing.T) {
	ctx := context.Background()
	ts := time.Now().Format(time.DateOnly)
	var buf bytes.Buffer
	SetOutput(&buf)

	log := New(&HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelInfo,
		ReplaceAttr: ReplaceAttr,
		TimeFormat:  time.DateOnly,
		Format:      TextFormat,
	})

	tests := []struct {
		fn     func(ctx context.Context, msg string, args ...any)
		msg    string
		output string
	}{
		{
			fn:     log.Debug,
			msg:    "Debug message",
			output: "",
		},
		{
			fn:     log.Info,
			msg:    "Info message",
			output: fmt.Sprintf("time=%s level=INFO msg=\"Info message\"\n", ts),
		},
		{
			fn:     log.Warn,
			msg:    "Warn message",
			output: fmt.Sprintf("time=%s level=WARN msg=\"Warn message\"\n", ts),
		},
		{
			fn:     log.Error,
			msg:    "Error message",
			output: fmt.Sprintf("time=%s level=ERROR msg=\"Error message\"\n", ts),
		},
	}

	for _, tt := range tests {
		tt.fn(ctx, tt.msg)
		output := string(buf.Bytes())
		if tt.output != output {
			t.Errorf("expected: %s, got: %s", tt.output, output)
		}
		buf.Reset()
	}
}

func TestHandler(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)

	log := New(&HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: ReplaceAttr,
		Format:      JSONFormat,
	})

	results := func() []map[string]any {
		ms, err := parseLines(buf.Bytes(), parseJSON)
		if err != nil {
			t.Fatal(err)
		}
		return ms
	}
	if err := slogtest.TestHandler(log.GetHandler(), results); err != nil {
		t.Fatal(err)
	}
}

func parseLines(src []byte, parse func([]byte) (map[string]any, error)) ([]map[string]any, error) {
	fmt.Println(string(src))
	var records []map[string]any
	for _, line := range bytes.Split(src, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		m, err := parse(line)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", string(line), err)
		}
		records = append(records, m)
	}
	return records, nil
}

func parseJSON(bs []byte) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, err
	}
	return m, nil
}
