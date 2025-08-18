package hue

import (
	"fmt"
	"log/slog"
	"runtime"
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
	// Stacktrace specifies options for stack trace handling.
	Stacktrace StacktraceOptions
	// Styles defines the styling options for components of the log output.
	Styles *Styles
}

type StacktraceOptions struct {
	// DisplayLastFrames limits the number of frames displayed in the stack trace.
	DisplayLastFrames int
	// MaxFrames limits the total number of frames processed in the stack trace.
	// This is check before DisplayLastFrames and can be used to limit the number of frames processed.
	MaxFrames int
	// ShouldDisplay is a function that determines whether to display the stack trace for a given log record.
	ShouldDisplay func(record slog.Record) bool
	StackProvider StackProvider
}

type StackProvider func(record slog.Record) []StackFrame

func DefaultOptions(level slog.Level) Options {
	return Options{
		Level:      level,
		TimeFormat: DefaultTimeFormat,
		AddPrefix:  true,
		AddSource:  false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
		Styles:     DefaultStyles(),
		SourceLink: FileSourceLink,
		Stacktrace: StacktraceOptions{
			DisplayLastFrames: 5,
			ShouldDisplay:     func(record slog.Record) bool { return record.Level >= slog.LevelError },
			StackProvider:     DefaultStackProvider(32),
		},
	}
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

func DefaultStackProvider(maxFrames int) StackProvider {
	return func(rec slog.Record) []StackFrame {
		callers := make([]uintptr, maxFrames)
		n := runtime.Callers(0, callers)
		callers = callers[:n]

		// we need to reduce the slice down to the frames after and including the record's PC
		for i, pc := range callers {
			if pc == rec.PC {
				callers = callers[i:]
				break
			}
		}

		frames := filterStacktraceFrames(runtime.CallersFrames(callers))

		returnFrames := make([]StackFrame, 0, len(frames))
		for _, frame := range frames {
			returnFrames = append(returnFrames, StackFrame{
				Function: formatFunctionName(frame.Function),
				File:     cleanFramePath(frame.File),
				Line:     frame.Line,
			})
		}

		return returnFrames
	}
}
