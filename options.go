package hue

import (
	"fmt"
	"log/slog"
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
	// AddPrefix determines whether Prefix attributes are rendered before the message.
	AddPrefix bool
	// Styles defines the styling options for components of the log output.
	Styles *Styles
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
