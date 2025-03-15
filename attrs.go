package hue

import (
	"log/slog"

	"github.com/charmbracelet/lipgloss"
)

const ErrKey = "err"
const ServiceKey = "service"

// StyledAttr is an interface that defines a custom style for a slog.Attr.
type StyledAttr interface {
	Style() lipgloss.Style
}

// PrefixAttr is an interface that marks an attribute as being used as a prefix.
type PrefixAttr interface {
	Prefix() bool
}

type errorAttr struct{ error }

func Err(err error) slog.Attr {
	if err != nil {
		err = errorAttr{err}
	}

	return slog.Any(ErrKey, err)
}

func (e errorAttr) Style() lipgloss.Style { return errorAttrStyle }

// serviceAttr is a custom slog.Attr that is used to style service names and mark them as log prefixes.
type serviceAttr string

func (s serviceAttr) Style() lipgloss.Style { return serviceAttrStyle }
func (s serviceAttr) Prefix() bool          { return true }

func Service(name string) slog.Attr {
	return slog.Any(ServiceKey, serviceAttr(name))
}
