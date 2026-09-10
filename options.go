package hue

import (
	"fmt"
	"log/slog"
)

// ColorMode controls whether/how hue colors its output.
type ColorMode int

const (
	// ColorAuto detects whether color should be used based on the
	// destination writer and environment (NO_COLOR, CLICOLOR, etc).
	// This is the default.
	ColorAuto ColorMode = iota
	// ColorAlways forces color output regardless of the destination or environment.
	ColorAlways
	// ColorNever disables color output entirely.
	ColorNever
)

type Options struct {
	// Level sets the minimum log level for the handler.
	Level slog.Leveler
	// TimeFormat sets the format for the time attribute.
	TimeFormat string
	// ReplaceAttr can be used to modify or remove attributes before they are logged.
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	// AddSource determines whether to include source file and line number in the log output.
	AddSource bool
	// SourceLink generates a terminal-friendly hyperlink for the source file and line number.
	// It should return a string that can be used in terminal emulators that support hyperlinks.
	// If an empty string is returned, no link will be generated.
	SourceLink func(source *slog.Source) string
	// AddScope determines whether Scope attributes are rendered as part of the log line.
	AddScope bool
	// Styles defines the styling options for components of the log output.
	Styles *Styles
	// View controls how log lines (and their attributes) are rendered.
	// Defaults to NewCompactView().
	View View
	// Color controls whether/how output is colored. Defaults to ColorAuto.
	Color ColorMode
	// Width sets a fixed wrap width for messages and attributes. A value of
	// 0 (the default) auto-detects the width from the destination terminal,
	// falling back to unbounded (no wrapping) when it isn't a terminal.
	Width int
}

func FileSourceLink(source *slog.Source) string {
	if source == nil {
		return ""
	}

	return fmt.Sprintf("file://%s#L%d", source.File, source.Line)
}

func VscodeSourceLink(source *slog.Source) string {
	if source == nil {
		return ""
	}

	return fmt.Sprintf("vscode://file/%s:%d", source.File, source.Line)
}
