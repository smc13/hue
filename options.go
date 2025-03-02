package hue

import "log/slog"

type Options struct {
	Level         slog.Leveler
	TimeFormat    string
	SupportsColor bool
	ReplaceAttr   func(groups []string, a slog.Attr) slog.Attr
	WithCaller    bool
}
