package hue

import (
	"log/slog"
)

type Options struct {
	Level       slog.Leveler
	TimeFormat  string
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	AddSource   bool
	AddPrefix   bool
	Styles      *Styles
}
