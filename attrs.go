package hue

import (
	"log/slog"

	"github.com/charmbracelet/lipgloss"
)

const ErrKey = "err"
const ServiceKey = "service"

type StyledAttr interface {
	Style() lipgloss.Style
}

type errorAttr struct{ error }

func Err(err error) slog.Attr {
	if err != nil {
		err = errorAttr{err}
	}

	return slog.Any(ErrKey, err)
}

func (e errorAttr) Style() lipgloss.Style {
	return errorAttrStyle
}

// serviceAttr is a custom slog.Attr that is used to style service names.
type serviceAttr string

func (s serviceAttr) Style() lipgloss.Style {
	return serviceAttrStyle
}

func Service(name string) slog.Attr {
	return slog.Any(ServiceKey, serviceAttr(name))
}
