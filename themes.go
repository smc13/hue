package hue

import (
	"log/slog"

	"charm.land/lipgloss/v2"
)

// DefaultStyles returns hue's default styling.
func DefaultStyles() *Styles {
	return &Styles{
		Time:   lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Attr:   lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Key:    lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Scope:  lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
		Source: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Levels: map[slog.Level]lipgloss.Style{
			slog.LevelDebug: lipgloss.NewStyle().Foreground(lipgloss.Color("8")).SetString("DBG"),
			slog.LevelInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color("12")).SetString("INF"),
			slog.LevelWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("11")).SetString("WRN"),
			slog.LevelError: lipgloss.NewStyle().Foreground(lipgloss.Color("9")).SetString("ERR"),
		},
		Values: ValueStyles{
			Bool:     lipgloss.NewStyle().Foreground(lipgloss.Color("141")),
			Number:   lipgloss.NewStyle().Foreground(lipgloss.Color("114")),
			Duration: lipgloss.NewStyle().Foreground(lipgloss.Color("211")),
			Time:     lipgloss.NewStyle().Foreground(lipgloss.Color("211")),
		},
	}
}
