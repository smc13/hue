package hue

import (
	"log/slog"

	"github.com/charmbracelet/lipgloss"
)

type StyledValue interface {
	Style() lipgloss.Style
}

const ErrKey = "err"

type errorAttr struct{ error }

func Err(err error) slog.Attr {
	if err != nil {
		err = errorAttr{err}
	}

	return slog.Any(ErrKey, err)
}

func (e errorAttr) Style() lipgloss.Style {
	return errorLevelStyle.Faint(true)
}
