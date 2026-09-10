package hue

import (
	"log/slog"

	"charm.land/lipgloss/v2"
)

// ValueStyles defines styling for attribute values, selected based on the
// underlying slog.Kind of the value. These are only used when an attribute
// doesn't implement StyledAttr, which always takes precedence.
type ValueStyles struct {
	// Styling for boolean values.
	Bool lipgloss.Style
	// Styling for numeric values (int64, uint64, float64).
	Number lipgloss.Style
	// Styling for time.Duration values.
	Duration lipgloss.Style
	// Styling for time.Time values.
	Time lipgloss.Style
}

// Styles defines the styling options for the hue logger.
type Styles struct {
	// Styling for the time attribute.
	Time lipgloss.Style
	// Default styling for attribute values.
	// Used as a fallback for strings and any value that isn't otherwise
	// styled by StyledAttr or ValueStyles.
	Attr lipgloss.Style
	// Styling for attribute keys. Applied uniformly regardless of value
	// type or custom attribute styling, so keys stay visually consistent.
	Key lipgloss.Style
	// Styling for scope attributes (see ScopeAttr).
	Scope lipgloss.Style
	// Styling for the source file and line number attribute.
	Source lipgloss.Style
	// Styling for log levels.
	// Custom log levels can be added to customise the output
	// eg: `Levels[slog.LevelDebug] = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).SetString("TRC")`
	Levels map[slog.Level]lipgloss.Style
	// Styling for attribute values, keyed by kind.
	Values ValueStyles
}

// levelStyle returns the configured style for level, falling back to the
// default attribute style (with the level's text set) if none is configured.
// Levels are always rendered bold, regardless of custom styling, so they
// stand out from the rest of the line.
func (s *Styles) levelStyle(level slog.Level) lipgloss.Style {
	if style, ok := s.Levels[level]; ok {
		return style.Bold(true)
	}

	return s.Attr.SetString(level.String()).Bold(true)
}
