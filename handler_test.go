package hue_test

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/smc13/hue"
)

func TestHandlerCompact(t *testing.T) {
	log := slog.New(hue.New(os.Stderr, &hue.Options{
		Level:      slog.LevelDebug,
		AddSource:  true,
		AddScope:   true,
		Color:      hue.ColorAlways,
		Width:      100,
		TimeFormat: hue.DefaultTimeFormat,
	}))

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warning message", hue.Scope("my-scope"))
	log.Error("error message", hue.Err(errors.New("something broke!")))

	log.Debug("debug message with attrs",
		slog.String("key", "value"),
		slog.Bool("enabled", true),
		slog.Int("count", 42),
		slog.Group("testing", slog.Time("now", time.Now()), slog.Duration("duration", 5*time.Second)),
	)

	// Service is a deprecated alias for Scope; kept working for backwards compatibility.
	serviceLogger := log.With(hue.Service("database"))
	serviceLogger.Info("database info message")

	subServiceLogger := serviceLogger.With(hue.Scope("sub"))
	subServiceLogger.Info("database sub message")

	log.Info("this is a long message that should wrap onto a new line once it exceeds the configured width, with continuation lines indented to align under the start of the message")
}

// TestOptionsMerge verifies that a partially-populated Options only
// overrides the fields it explicitly sets, and falls back to defaults
// (eg. TimeFormat, SourceLink, Styles, View) for everything else.
func TestOptionsMerge(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(hue.New(&buf, &hue.Options{
		Level:     slog.LevelDebug,
		AddSource: true,
		Color:     hue.ColorNever,
	}))

	log.Info("hello")

	out := buf.String()

	// TimeFormat should fall back to hue.DefaultTimeFormat (time.TimeOnly,
	// eg. "15:04:05") rather than being wiped to "" by the partial Options.
	if !regexp.MustCompile(`\d{2}:\d{2}:\d{2}`).MatchString(out) {
		t.Fatalf("expected default TimeFormat to be applied, got: %q", out)
	}

	// SourceLink should fall back to hue.FileSourceLink and produce a
	// "source=" attribute since AddSource was explicitly set to true.
	if !regexp.MustCompile(`source=handler_test\.go:\d+`).MatchString(out) {
		t.Fatalf("expected default SourceLink to be applied, got: %q", out)
	}
}

func TestHandlerTree(t *testing.T) {
	log := slog.New(hue.New(os.Stderr, &hue.Options{
		Level:      slog.LevelDebug,
		AddScope:   true,
		Color:      hue.ColorAlways,
		Width:      100,
		View:       hue.NewTreeView(),
		TimeFormat: hue.DefaultTimeFormat,
	}))

	log = log.With(hue.Scope("auth-service"))

	log.Info("user authenticated",
		slog.Int("userID", 123),
		slog.String("requestID", "abc-123"),
		slog.Group("testing", slog.Time("now", time.Now()), slog.Duration("duration", 5*time.Second)),
	)
	log.Error("login failed", hue.Err(errors.New("invalid credentials")), slog.Bool("locked", false))
}

func TestHandlerSingleLine(t *testing.T) {
	log := slog.New(hue.New(os.Stderr, &hue.Options{
		Level:      slog.LevelDebug,
		AddSource:  true,
		AddScope:   true,
		Color:      hue.ColorAlways,
		Width:      100,
		View:       hue.NewSingleLineView(),
		TimeFormat: hue.DefaultTimeFormat,
	}))

	log = log.With(hue.Scope("auth-service"))

	log.Info("user authenticated",
		slog.Int("userID", 123),
		slog.String("requestID", "abc-123"),
		slog.Group("testing", slog.Time("now", time.Now()), slog.Duration("duration", 5*time.Second)),
	)
	log.Error("login failed", hue.Err(errors.New("invalid credentials")), slog.Bool("locked", false))
	log.Info("this is a long message that would wrap in the other views, but stays on a single line here, along with any attributes", slog.Int("count", 42))
}
