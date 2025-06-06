package hue

import (
	"log/slog"

	"github.com/charmbracelet/lipgloss"
)

// Styles defines the styling options for the hue logger.
type Styles struct {
	// Styling for the time attribute.
	Time lipgloss.Style
	// Default styling for attributes.
	// Overridable by StyledAttr interface.
	Attr lipgloss.Style
	// Styling for attributes that are used as prefixes.
	Prefix lipgloss.Style
	// Styling for the source file and line number.
	Source lipgloss.Style
	// Styling for log levels.
	// Custom log levels can be added to customise the output
	// eg: `Levels[slog.LevelDebug] = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).SetString("TRC")`
	Levels map[slog.Level]lipgloss.Style
}

func DefaultStyles() *Styles {
	return &Styles{
		Time:   lipgloss.NewStyle().Foreground(lipgloss.Color("0")),
		Attr:   lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Prefix: lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
		Source: lipgloss.NewStyle().Foreground(lipgloss.Color("0")),
		Levels: map[slog.Level]lipgloss.Style{
			slog.LevelDebug: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).SetString("DBG"),
			slog.LevelInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color("4")).SetString("INF"),
			slog.LevelWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")).SetString("WRN"),
			slog.LevelError: lipgloss.NewStyle().Foreground(lipgloss.Color("1")).SetString("ERR"),
		},
	}
}
