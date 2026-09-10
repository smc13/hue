package hue

import (
	"log/slog"

	"charm.land/lipgloss/v2"
)

const ErrKey = "err"
const ScopeKey = "scope"

// ServiceKey is a deprecated alias for ScopeKey.
//
// Deprecated: use ScopeKey instead.
const ServiceKey = ScopeKey

// StyledAttr is an interface that defines a custom style for a slog.Attr.
type StyledAttr interface {
	Style() lipgloss.Style
}

// ScopeAttr is an interface that marks an attribute as being used to scope
// (prefix) log lines.
type ScopeAttr interface {
	Prefix() bool
}

// PrefixAttr is a deprecated alias for ScopeAttr.
//
// Deprecated: use ScopeAttr instead.
type PrefixAttr = ScopeAttr

type errorAttr struct{ error }

// Err is a custom `slog.Attr` that is used to style errors and provides a consistent error attribute.
// Works as a normal attribute when used with other handlers.
func Err(err error) slog.Attr {
	if err != nil {
		err = errorAttr{err}
	}

	return slog.Any(ErrKey, err)
}

func (e errorAttr) Style() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
}

// scopeAttr is a custom slog.Attr that is used to style scope names and mark them as log scopes.
type scopeAttr string

func (s scopeAttr) Style() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
}
func (s scopeAttr) Prefix() bool { return true }

// Scope is a custom `slog.Attr` that is used to style scope names and mark them as log scopes.
// If not used with logger.WithAttrs, it will not be used as a scope and instead be displayed as a regular attribute.
// Works as a normal attribute when used with other handlers.
func Scope(name string) slog.Attr {
	return slog.Any(ScopeKey, scopeAttr(name))
}

// Service is a deprecated alias for Scope.
//
// Deprecated: use Scope instead.
func Service(name string) slog.Attr {
	return Scope(name)
}
